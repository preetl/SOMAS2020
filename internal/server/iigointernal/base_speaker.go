package iigointernal

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	//"math/rand"
	_ "time"

	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/voting"
	"github.com/pkg/errors"
)

type baseSpeaker struct {
	Id            int
	budget        int
	judgeSalary   int
	ruleToVote    string
	ballotBox     voting.BallotBox
	votingResult  bool
	clientSpeaker roles.Speaker
}

//withdrawJudgeSalary deducts appropriate amount of Judge salary from the CP
func (s *baseSpeaker) withdrawJudgeSalary(gameState *gamestate.GameState) error {
	//Note, Should make it so that the speaker cannot keep the resources, they have the power to pay the judge,
	//dont have the power to pocket the money
	var judgeSalary = int(rules.VariableMap["judgeSalary"].Values[0])
	var withdrawError = WithdrawFromCommonPool(judgeSalary, gameState)
	if withdrawError != nil {
		featureSpeaker.judgeSalary = judgeSalary
	}
	return withdrawError
}

//sendJudgeSalary asks the speaker how much to add to the budget of the Judge
func (s *baseSpeaker) sendJudgeSalary() {
	if s.clientSpeaker != nil {
		amount, err := s.clientSpeaker.PayJudge()
		if err == nil {
			//TODO: reorder actions so speaker could not provide a larger amount than withdrawn (what is withdrawn is sent)
			featureJudge.budget = amount
			return
		}
	}
	amount, _ := s.PayJudge()
	featureJudge.budget = amount
}

//setRuleToVote sets the rule to be voted on.
//This action has no cost. Serves as a utility function for non-compliant agents
func (s *baseSpeaker) setRuleToVote(givenRuleID string) {
	var chosenRuleID string
	var err error

	if s.clientSpeaker != nil {
		chosenRuleID, err = s.clientSpeaker.DecideAgenda(s.ruleToVote)
		if err != nil {
			chosenRuleID = givenRuleID
		}
	} else {
		chosenRuleID = givenRuleID
	}
	//TODO: log of chosenRule vs givenRuleID
	s.ruleToVote = chosenRuleID
}

//TODO: Write tests for setVotingResult
//setVotingResult is called by orchestration and provides the Speaker with the power to conduct a vote on a rule.
func (s *baseSpeaker) setVotingResult() {

	//TODO: for loop should not be done here
	var clientIDs []shared.ClientID
	for id := range getIslandAlive() {
		clientIDs = append(clientIDs, shared.ClientID(id))
	}

	if s.clientSpeaker != nil {
		ruleID, participatingIslands, err := s.clientSpeaker.DecideVote(s.ruleToVote, clientIDs)
		if err != nil {
			s.ballotBox = s.RunVote(s.ruleToVote, clientIDs)
		} else {
			s.ballotBox = s.RunVote(ruleID, participatingIslands)
		}
	} else {
		s.ballotBox = s.RunVote(s.ruleToVote, clientIDs)
	}

	//Vote counting always happens and the cost incurred through running the vote
	s.votingResult = s.ballotBox.CountVotesMajority()

}

//RunVote creates the voting object, returns votes by category (for, against) in BallotBox.
//Passing in empty ruleID or empty clientIDs results in no vote occurring
func (s *baseSpeaker) RunVote(ruleID string, clientIDs []shared.ClientID) voting.BallotBox {

	if ruleID == "" || len(clientIDs) == 0 {
		return voting.BallotBox{}
	}
	s.budget -= 10
	ruleVote := voting.RuleVote{}

	//TODO: check if rule is valid, otherwise return empty ballot, raise error?
	ruleVote.SetRule(ruleID)

	//TODO: intersection of islands alive and islands chosen to vote in case of client error
	//TODO: check if remaining slice is >0, otherwise return empty ballot, raise error?
	ruleVote.SetVotingIslands(clientIDs)

	ruleVote.GatherBallots(iigoClients)
	//TODO: log of vote occurring with ruleID, clientIDs
	//TODO: log of clientIDs vs islandsAllowedToVote
	//TODO: log of ruleID vs s.RuleToVote
	return ruleVote.GetBallotBox()
}

//announceVotingResult gives the speaker the power to declare a result of a vote
//(see spec to see conditions on what this means for a rule-abiding speaker).
//Called by orchestration.
func (s *baseSpeaker) announceVotingResult() error {

	var rule string
	var result bool
	var err error

	if s.clientSpeaker != nil {
		//Power to change what is declared completely, return "", _ for no announcement to occur
		rule, result, err = s.clientSpeaker.DecideAnnouncement(s.ruleToVote, s.votingResult)
		//TODO: log of given vs. returned rule and result
		if err != nil {
			rule, result, _ = s.DecideAnnouncement(s.ruleToVote, s.votingResult)
		}
	} else {
		rule, result, _ = s.DecideAnnouncement(s.ruleToVote, s.votingResult)
	}

	if rule != "" {
		//Deduct action cost
		s.budget -= 10

		s.reset()

		//Perform announcement
		broadcastToAllIslands(shared.TeamIDs[s.Id], generateVotingResultMessage(rule, result))
		return s.updateRules(rule, result)
	}

	s.reset()
	return nil
}

//reset resets internal variables for safety
func (s *baseSpeaker) reset() {
	s.ruleToVote = ""
	s.ballotBox = voting.BallotBox{}
	s.votingResult = false
}

//generateVotingResultMessage packs up the ruleID and the result in a package to be sent to the clients
func generateVotingResultMessage(ruleID string, result bool) map[int]baseclient.Communication {
	returnMap := map[int]baseclient.Communication{}

	returnMap[RuleName] = baseclient.Communication{
		T:        baseclient.CommunicationString,
		TextData: ruleID,
	}
	returnMap[RuleVoteResult] = baseclient.Communication{
		T:           baseclient.CommunicationBool,
		BooleanData: result,
	}

	return returnMap
}

//updateRules updates the rules depending on the ruleID provided and the result
func (s *baseSpeaker) updateRules(ruleName string, ruleVotedIn bool) error {
	s.budget -= 10
	//TODO: might want to log the errors as normal messages rather than completely ignoring them? But then Speaker needs access to client's logger
	notInRulesCache := errors.Errorf("Rule '%v' is not available in rules cache", ruleName)
	if ruleVotedIn {
		// _ = rules.PullRuleIntoPlay(ruleName)
		err := rules.PullRuleIntoPlay(ruleName)
		if err != nil {
			if err.Error() == notInRulesCache.Error() {
				return err
			}
		}
	} else {
		// _ = rules.PullRuleOutOfPlay(ruleName)
		err := rules.PullRuleOutOfPlay(ruleName)
		if err != nil {
			if err.Error() == notInRulesCache.Error() {
				return err
			}
		}

	}
	return nil

}

//appointNextJudge runs an election for next Judge
//This is not MVP, needs to be turned into a power rather than an external event
func (s *baseSpeaker) appointNextJudge(clientIDs []shared.ClientID) int {
	s.budget -= 10
	var election voting.Election
	election.ProposeElection(baseclient.Judge, voting.Plurality)
	election.OpenBallot(clientIDs)
	election.Vote(iigoClients)
	return int(election.CloseBallot())
}

//---- Speaker Interface Implementations ----
//TODO: move to separate file?

//PayJudge is the interface implementation and example of a well behaved Speaker,
//who decides to pay the speaker everything that has been withdrawn from the CP
//(should be decides to pay the judge what the rule dictates instead)
func (s *baseSpeaker) PayJudge() (int, error) {
	hold := s.judgeSalary
	s.judgeSalary = 0
	return hold, nil
}

//DecideAgenda the interface implementation and example of a well behaved Speaker
//who sets the vote to be voted on to be the rule the President provided
func (s *baseSpeaker) DecideAgenda(ruleID string) (string, error) {
	return ruleID, nil
}

//DecideVote is the interface implementation and example of a well behaved Speaker
//who calls a vote on the proposed rule and asks all available islands to vote.
//Return an empty string or empty []shared.ClientID for no vote to occur
func (s *baseSpeaker) DecideVote(ruleID string, aliveClients []shared.ClientID) (string, []shared.ClientID, error) {
	//TODO: disregard islands with sanctions
	return ruleID, aliveClients, nil
}

//DecideAnnouncement is the interface implementation and example of a well behaved Speaker
//A well behaved speaker announces what had been voted on and the corresponding result
//Return "", _ for no announcement to occur
func (s *baseSpeaker) DecideAnnouncement(ruleId string, result bool) (string, bool, error) {
	return ruleId, result, nil
}
