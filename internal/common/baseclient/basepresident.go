package baseclient

import (
	"math/rand"

	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

type BasePresident struct{}

// EvaluateAllocationRequests sets allowed resource allocation based on each islands requests
func (p *BasePresident) EvaluateAllocationRequests(resourceRequest map[shared.ClientID]shared.Resources, availCommonPool shared.Resources) shared.PresidentReturnContent {
	var requestSum shared.Resources
	resourceAllocation := make(map[shared.ClientID]shared.Resources)

	for _, request := range resourceRequest {
		requestSum += request
	}

	if requestSum < 0.75*availCommonPool || requestSum == 0 {
		resourceAllocation = resourceRequest
	} else {
		for id, request := range resourceRequest {
			resourceAllocation[id] = shared.Resources(request * availCommonPool * 3 / (4 * requestSum))
		}
	}

	return shared.PresidentReturnContent{
		ContentType: shared.PresidentAllocation,
		ResourceMap: resourceAllocation,
		ActionTaken: true,
	}
}

// PickRuleToVote chooses a rule proposal from all the proposals
func (p *BasePresident) PickRuleToVote(rulesProposals []string) shared.PresidentReturnContent {
	// DefaulContentType: No rules were proposed by the islands
	proposedRule := ""
	actionTaken := false

	// if some rules were proposed
	if len(rulesProposals) != 0 {
		proposedRule = rulesProposals[rand.Intn(len(rulesProposals))]
		actionTaken = true
	}

	return shared.PresidentReturnContent{
		ContentType:  shared.PresidentRuleProposal,
		ProposedRule: proposedRule,
		ActionTaken:  actionTaken,
	}
}

// SetTaxationAmount sets taxation amount for all of the living islands
// islandsResources: map of all the living islands and their remaining resources
func (p *BasePresident) SetTaxationAmount(islandsResources map[shared.ClientID]shared.Resources) shared.PresidentReturnContent {
	taxAmountMap := make(map[shared.ClientID]shared.Resources)
	for id, resourceLeft := range islandsResources {
		taxAmountMap[id] = shared.Resources(float64(resourceLeft) * rand.Float64())
	}
	return shared.PresidentReturnContent{
		ContentType: shared.PresidentTaxation,
		ResourceMap: taxAmountMap,
		ActionTaken: true,
	}
}

// PaySpeaker pays the speaker a salary.
func (p *BasePresident) PaySpeaker(salary shared.Resources) shared.PresidentReturnContent {
	// TODO : Implement opinion based salary payment.
	return shared.PresidentReturnContent{
		ContentType:   shared.PresidentSpeakerSalary,
		SpeakerSalary: salary,
		ActionTaken:   true,
	}
}

// CallSpeakerElection is called by the executive to decide on power-transfer
func (p *BasePresident) CallSpeakerElection(turnsInPower int, allIslands []shared.ClientID) shared.ElectionSettings {
	var electionsettings = shared.ElectionSettings{
		VotingMethod:  shared.Plurality,
		IslandsToVote: allIslands,
		HoldElection:  true,
	}
	return electionsettings
}

// DecideNextSpeaker returns the ID of chosen next Speaker
func (p *BasePresident) DecideNextSpeaker(winner shared.ClientID) shared.ClientID {
	return winner
}
