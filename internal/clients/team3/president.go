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

	// stores the declared resources of each island for that turn
	declaredResources map[shared.ClientID]shared.Resources

	// Parameters

	resourceSkew              uint
	equity                    float64
	commonPoolThresholdFactor float64
	saveCriticalIslands       bool
}

func (p *president) PaySpeaker(salary shared.Resources) (shared.Resources, bool) {
	// Use the base implementation
	return p.BasePresident.PaySpeaker(salary)
}

func (p *president) DecideNextSpeaker(winner shared.ClientID) shared.ClientID {
	// p.c.Logf("choosing speaker")
	// Naively choose group 0
	return shared.ClientID(0)
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

	return shared.Resources(int(sum) / len(shared.TeamIDs))
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
	resourceSkew := math.Max(float64(p.resourceSkew), 1)

	for island, req := range resourceRequest {
		sumRequest += req
		resources[island] = p.declaredResources[island] * shared.Resources(math.Pow(resourceSkew, 1-p.c.trustScore[island]))
	}

	avgRequest = findAvgNoTails(resourceRequest)
	avgResource = findAvgNoTails(resources)

	for island, resource := range resources {
		allocations[island] = float64(avgRequest) + p.equity*(float64(avgResource-resource)+float64(resourceRequest[island]-avgRequest))

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

	if p.saveCriticalIslands {
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
func (p *president) SetTaxationAmount(islandsResources map[shared.ClientID]shared.Resources) (map[shared.ClientID]shared.Resources, bool) {
	//decide if we want to run SetTaxationAmount
	p.c.declaredResources = islandsResources
	gameState := p.c.BaseClient.ServerReadHandle.GetGameState()
	var resourcesRequired float64
	if len(p.c.disasterPredictions) != 0 {
		disaster := p.c.disasterPredictions[int(gameState.Turn)][(p.c.BaseClient.GetID())]
		resourcesRequired = (disaster.Magnitude - float64(gameState.CommonPool)/float64(disaster.TimeLeft))
	}
	resourcesRequired = 100.0 - float64(gameState.CommonPool) //will change to average magnitude when we got it somehow magick
	AveTax := resourcesRequired / float64(len(islandsResources))
	var adjustedResources []float64
	adjustedResourcesMap := make(map[shared.ClientID]shared.Resources)
	for island, resource := range islandsResources {
		adjustedResource := shared.Resources(math.Pow(float64(resource)*p.c.params.resourcesSkew, (1 - p.c.trustScore[island])))
		adjustedResources = append(adjustedResources, float64(adjustedResource))
		adjustedResourcesMap[island] = adjustedResource
	}
	AveAdjustedResources := getAverage(adjustedResources)
	taxationMap := make(map[shared.ClientID]shared.Resources)
	for island, resources := range adjustedResourcesMap {
		taxation := shared.Resources(AveTax) + shared.Resources(p.c.params.equity)*(resources-shared.Resources(AveAdjustedResources))
		if island == p.c.BaseClient.GetID() {
			taxation -= shared.Resources(p.c.params.selfishness) * taxation
		}
		taxation = shared.Resources(math.Max(float64(taxation), 0.0))
		taxationMap[island] = taxation
	}
	return taxationMap, true

}
