package team3

import (
	"math"
	// "github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	// "github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	// "github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

/*
	DecideForage() (shared.ForageDecision, error)
	ForageUpdate(shared.ForageDecision, shared.Resources)
*/

func (c *client) DecideForage() (shared.ForageDecision, error) {

	// No risk -> minimum is 2 times the critical threshold, Full risk -> minimum is the critical threshold
	safetyFactor := 1.0 + 2*(1-c.params.riskFactor)

	//we want to have more than the critical threshold leftover after foraging
	var minimumLeftoverResources = float64(c.criticalStatePrediction.upperBound) * safetyFactor

	var foragingInvestment = 0.0
	//for now we invest everything we can, because foraging is iffy.
	if c.ServerReadHandle.GetGameState().ClientInfo.LifeStatus == shared.Alive {
		foragingInvestment = math.Max(float64(c.ServerReadHandle.GetGameState().ClientInfo.Resources)-minimumLeftoverResources, 0)
	}

	var forageType shared.ForageType

	if c.computeRecentExpectedROI(shared.DeerForageType) > c.computeRecentExpectedROI(shared.FishForageType) {
		forageType = shared.DeerForageType
	} else {
		forageType = shared.FishForageType
	}

	return shared.ForageDecision{
		Type:         forageType,
		Contribution: shared.Resources(foragingInvestment * ((1 + (1 - c.params.riskFactor)) / 2)),
	}, nil
}

func (c *client) computeRecentExpectedROI(forageType shared.ForageType) float64 {
	data := c.forageData[forageType]
	var sumOfROI float64
	var numberOfROI uint

	if len(data) == 0 {
		return 0
	}

	for _, forage := range data {
		if uint(forage.turn) == c.ServerReadHandle.GetGameState().Turn-1 || uint(forage.turn) == c.ServerReadHandle.GetGameState().Turn-2 {
			sumOfROI += float64((forage.amountReturned / forage.amountContributed) * 100)
			numberOfROI++
		}
	}
	return sumOfROI / float64(numberOfROI)
}

func (c *client) ForageUpdate(forageDecision shared.ForageDecision, outcome shared.Resources, numberCaught uint) {
	c.forageData[forageDecision.Type] =
		append(
			c.forageData[forageDecision.Type],
			ForageData{
				amountContributed: forageDecision.Contribution,
				amountReturned:    outcome,
				turn:              c.ServerReadHandle.GetGameState().Turn,
			},
		)
}
