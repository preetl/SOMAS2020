package team3

import (
	"math"

	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

type president struct {
	// Base implementation
	*baseclient.BasePresident
	// Our client
	c *client
}

func (p *president) PaySpeaker(salary shared.Resources) (shared.Resources, bool) {
	// Use the base implementation
	return p.BasePresident.PaySpeaker(salary)
}

func (p *president) DecideNextSpeaker(winner shared.ClientID) shared.ClientID {
	p.c.clientPrint("choosing speaker")
	// Naively choose group 0
	return mostTrusted(p.c.trustMapAgg)
}

// Computes average request, excluding top and bottom
func findAvgNoTails(resourceRequest map[shared.ClientID]shared.Resources) shared.Resources {
	var sum shared.Resources
	minClient := shared.TeamIDs[0]
	maxClient := shared.TeamIDs[0]

	// Find min and max requests
	for island, request := range resourceRequest {
		if request < resourceRequest[minClient] {
			minClient = island
		}
		if request > resourceRequest[maxClient] {
			maxClient = island
		}
	}

	// Compute average ignoring highest and lowest
	for island, request := range resourceRequest {
		if island != minClient || island != maxClient {
			sum += request
		}
	}

	return shared.Resources(int(sum) / len(resourceRequest))
}

// EvaluateAllocationRequests sets allowed resource allocation based on each islands requests
func (p *president) EvaluateAllocationRequests(resourceRequest map[shared.ClientID]shared.Resources, availCommonPool shared.Resources) (map[shared.ClientID]shared.Resources, bool) {
	p.c.Logf("Evaluating allocations...")
	var allocations, allocWeights map[shared.ClientID]float64
	var avgResource, avgRequest shared.Resources
	var resources map[shared.ClientID]shared.Resources
	var allocSum float64
	var commonPoolThreshold shared.Resources
	var finalAllocations map[shared.ClientID]shared.Resources
	var sumRequest shared.Resources

	// Make sure resource skew is greater than 1
	resourceSkew := math.Max(float64(p.c.params.resourcesSkew), 1)

	for island, req := range resourceRequest {
		sumRequest += req
		resources[island] = p.c.declaredResources[island] * shared.Resources(math.Pow(resourceSkew, 1-p.c.trustScore[island]))
	}

	avgRequest = findAvgNoTails(resourceRequest)
	avgResource = findAvgNoTails(resources)

	for island, resource := range resources {
		allocations[island] = float64(avgRequest) + p.c.params.equity*(float64(avgResource-resource)+float64(resourceRequest[island]-avgRequest))

		if island == id {
			allocations[island] += math.Max(float64(resourceRequest[island])-allocations[island]*p.c.params.selfishness, 0)
		} else {
			allocations[island] += float64(resourceRequest[island]) - allocations[island]*(1/p.c.params.selfishness)
			allocations[island] = math.Min(float64(resourceRequest[island]), allocations[island]) // to prevent overallocating
		}
	}

	// Collect weights
	for _, alloc := range allocations {
		allocSum += alloc
	}
	// Normalise
	for island, alloc := range allocations {
		allocWeights[island] = alloc / allocSum
	}

	commonPoolThreshold = shared.Resources(p.commonPoolThresholdFactor * float64(availCommonPool))
	if p.c.params.saveCriticalIsland {
		for island := range resourceRequest {
			if resources[island] < p.c.criticalStatePrediction.lowerBound {
				finalAllocations[island] = shared.Resources(math.Max((allocWeights[island] * float64(commonPoolThreshold)), float64(p.c.criticalStatePrediction.lowerBound-resources[island])))
			} else {
				finalAllocations[island] = 0
			}
		}
	}

	for island := range resourceRequest {
		if finalAllocations[island] == 0 {
			if sumRequest < commonPoolThreshold {
				finalAllocations[island] = shared.Resources(allocWeights[island] * float64(sumRequest))
			} else {
				finalAllocations[island] = shared.Resources(allocWeights[island] * float64(commonPoolThreshold))
			}
		}
	}

	// Curently always evaluate, would there be a time when we don't want to?
	return finalAllocations, true
}
