package iigointernal

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	"github.com/SOMAS2020/SOMAS2020/internal/common/voting"
)

type executive struct {
	ID               shared.ClientID
	clientPresident  roles.President
	budget           shared.Resources
	speakerSalary    shared.Resources
	RulesProposals   []string
	ResourceRequests map[shared.ClientID]shared.Resources
}

// returnSpeakerSalary returns the salary to the common pool.
func (e *executive) returnSpeakerSalary() shared.Resources {
	x := e.speakerSalary
	e.speakerSalary = 0
	return x
}

// Get rule proposals to be voted on from remaining islands
// Called by orchestration
func (e *executive) setRuleProposals(rulesProposals []string) {
	e.RulesProposals = rulesProposals
}

// Set approved resources request for all the remaining islands
// Called by orchestration
func (e *executive) setAllocationRequest(resourceRequests map[shared.ClientID]shared.Resources) {
	e.ResourceRequests = resourceRequests
}

// Get rules to be voted on to Speaker
// Called by orchestration at the end of the turn
func (e *executive) getRuleForSpeaker() string {
	e.budget -= 10
	result, _ := e.clientPresident.PickRuleToVote(e.RulesProposals)
	return result
}

// Send Tax map all the remaining islands
// Called by orchestration at the end of the turn
func (e *executive) getTaxMap(islandsResources map[shared.ClientID]shared.Resources) map[shared.ClientID]shared.Resources {
	e.budget -= 10
	result, _ := e.clientPresident.SetTaxationAmount(islandsResources)
	return result
}

func (e *executive) broadcastTaxation(islandsResources map[shared.ClientID]shared.Resources) {
	e.budget -= 10
	taxAmountMap := e.getTaxMap(islandsResources)
	for _, v := range getIslandAlive() {
		d := baseclient.Communication{T: baseclient.CommunicationInt, IntegerData: int(taxAmountMap[shared.ClientID(int(v))])}
		data := make(map[int]baseclient.Communication)
		data[TaxAmount] = d
		communicateWithIslands(shared.TeamIDs[int(v)], shared.TeamIDs[e.ID], data)
	}
}

// Send Tax map all the remaining islands
// Called by orchestration at the end of the turn
func (e *executive) getAllocationRequests(commonPool shared.Resources) map[shared.ClientID]shared.Resources {
	e.budget -= 10
	result, _ := e.clientPresident.EvaluateAllocationRequests(e.ResourceRequests, commonPool)
	return result
}

func (e *executive) requestAllocationRequest() {
	allocRequests := make(map[shared.ClientID]shared.Resources)
	for _, v := range getIslandAlive() {
		allocRequests[shared.ClientID(int(v))] = iigoClients[shared.ClientID(int(v))].CommonPoolResourceRequest()
	}
	AllocationAmountMapExport = allocRequests
	e.setAllocationRequest(allocRequests)

}

func (e *executive) replyAllocationRequest(commonPool shared.Resources) {
	e.budget -= 10
	allocationMap := e.getAllocationRequests(commonPool)
	for _, v := range getIslandAlive() {
		d := baseclient.Communication{T: baseclient.CommunicationInt, IntegerData: int(allocationMap[shared.ClientID(int(v))])}
		data := make(map[int]baseclient.Communication)
		data[AllocationAmount] = d
		communicateWithIslands(shared.TeamIDs[int(v)], shared.TeamIDs[e.ID], data)
	}
}

func (e *executive) appointNextSpeaker(clientIDs []shared.ClientID) shared.ClientID {
	e.budget -= 10
	var election voting.Election
	election.ProposeElection(baseclient.Speaker, voting.Plurality)
	election.OpenBallot(clientIDs)
	election.Vote(iigoClients)
	return election.CloseBallot()
}

func (e *executive) withdrawSpeakerSalary(gameState *gamestate.GameState) error {
	var speakerSalary = shared.Resources(rules.VariableMap["speakerSalary"].Values[0])
	var withdrawError = WithdrawFromCommonPool(speakerSalary, gameState)
	if withdrawError != nil {
		e.speakerSalary = speakerSalary
	}
	return withdrawError
}

func (e *executive) sendSpeakerSalary() {
	amount, _ := e.clientPresident.PaySpeaker(e.speakerSalary)
	e.budget = amount
}

func (e *executive) reset(val string) error {
	e.ID = 0
	e.clientPresident = nil
	e.budget = 0
	e.ResourceRequests = map[shared.ClientID]shared.Resources{}
	e.RulesProposals = []string{}
	e.speakerSalary = 0
	return nil
}

func (e *executive) requestRuleProposal() {
	e.budget -= 10
	var rules []string
	for _, v := range getIslandAlive() {
		rules = append(rules, iigoClients[shared.ClientID(int(v))].RuleProposal())
	}

	e.setRuleProposals(rules)
}

func getIslandAlive() []float64 {
	return rules.VariableMap["islands_alive"].Values
}