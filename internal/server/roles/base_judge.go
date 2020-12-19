package roles

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/pkg/errors"
)

// base Judge object
type BaseJudge struct {
	id                int
	budget            int
	presidentSalary   int
	ballotID          int
	resAllocID        int
	speakerID         int
	presidentID       int
	evaluationResults map[int]EvaluationReturn
}

func (j *BaseJudge) init() {
	j.ballotID = 0
	j.resAllocID = 0
}

func (j *BaseJudge) withdrawPresidentSalary() {
	// Withdraw president salary from the common pool
	// Call common withdraw function with president as parameter
}

func (j *BaseJudge) payPresident() {
	// Pay the president
	// Call common pay function with president as parameter
}

func (j *BaseJudge) setSpeakerAndPresidentIDs(speakerId int, presidentId int) {
	j.speakerID = speakerId
	j.presidentID = presidentId
}

type EvaluationReturn struct {
	rules       []rules.RuleMatrix
	evaluations []bool
}

func (j *BaseJudge) inspectHistory() {
	outputMap := map[int]EvaluationReturn{}
	for _, v := range TurnHistory {
		variablePairs := v.pairs
		clientID := v.clientID
		var rulesAffected []string
		for _, v2 := range variablePairs {
			valuesToBeAdded, err := PickUpRulesByVariable(v2.VariableName, rules.RulesInPlay)
			if err == nil {
				rulesAffected = append(rulesAffected, valuesToBeAdded...)
			}
			err = rules.UpdateVariable(v2.VariableName, v2)
			if err != nil {
				return
			}
		}
		if _, ok := outputMap[clientID]; !ok {
			tempTemp := EvaluationReturn{
				rules:       []rules.RuleMatrix{},
				evaluations: []bool{},
			}
			outputMap[clientID] = tempTemp
		}
		tempReturn := outputMap[clientID]
		for _, v3 := range rulesAffected {
			evaluation, _ := rules.BasicBooleanRuleEvaluator(v3)
			tempReturn.rules = append(tempReturn.rules, rules.RulesInPlay[v3])
			tempReturn.evaluations = append(tempReturn.evaluations, evaluation)
		}
	}
	j.evaluationResults = outputMap
}

func (j *BaseJudge) inspectBallot() (bool, error) {
	// 1. Evaluate difference between newRules and oldRules to check
	//    rule changes are in line with ruleToVote in previous ballot
	// 2. Compare each ballot action adheres to rules in ruleSet matrix
	rulesAffectedBySpeaker := j.evaluationResults[j.speakerID]
	indexOfBallotRule, err := searchForRule("inspect_ballot_rule", rulesAffectedBySpeaker.rules)
	if err == nil {
		return rulesAffectedBySpeaker.evaluations[indexOfBallotRule], nil
	} else {
		return true, errors.Errorf("Speaker did not conduct any ballots")
	}
}

func (j *BaseJudge) inspectAllocation() (bool, error) {
	// 1. Evaluate difference between commonPoolNew and commonPoolOld
	//    to check resource allocation changes are in line with resourceRequests
	//    in previous resourceAllocation
	// 2. Compare each resource allocation action adheres to rules in ruleSet
	//    matrix
	rulesAffectedByPresident := j.evaluationResults[j.presidentID]
	indexOfAllocRule, err := searchForRule("inspect_allocation_rule", rulesAffectedByPresident.rules)
	if err == nil {
		return rulesAffectedByPresident.evaluations[indexOfAllocRule], nil
	} else {
		return true, errors.Errorf("President didn't conduct any allocations")
	}
}

func searchForRule(ruleName string, listOfRuleMatrices []rules.RuleMatrix) (int, error) {
	for i, v := range listOfRuleMatrices {
		if v.RuleName == ruleName {
			return i, nil
		}
	}
	return 0, errors.Errorf("The rule name '%v' was not found", ruleName)
}

func (j *BaseJudge) declareSpeakerPerformance() (int, bool, int, bool) {

	j.ballotID++
	result, err := j.inspectBallot()

	conductedRole := err == nil

	return j.ballotID, result, j.speakerID, conductedRole
}

func (j *BaseJudge) declarePresidentPerformance() (int, bool, int, bool) {

	j.resAllocID++
	result, err := j.inspectAllocation()

	conductedRole := err == nil

	return j.resAllocID, result, j.presidentID, conductedRole
}
