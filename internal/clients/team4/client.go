// Package team4 contains code for team 4's client implementation
package team4

import (
	"math"
	"sort"
	"testing"

	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	"gonum.org/v1/gonum/mat"
)

const id = shared.Team4

func init() {
	baseclient.RegisterClientFactory(id, func() baseclient.Client { return NewClient(id) })
}

func newClientInternal(clientID shared.ClientID, testing *testing.T) client {
	// have some config json file or something?
	internalConfig := internalParameters{
		greediness:       0,
		selfishness:      0,
		fairness:         0,
		collaboration:    0,
		riskTaking:       0,
		maxPardonTime:    3,
		maxTierToPardon:  shared.SanctionTier3,
		minTrustToPardon: 0.6,
	}

	iigoObs := iigoObservation{
		allocationGranted: shared.Resources(0),
		taxDemanded:       shared.Resources(0),
	}
	iifoObs := iifoObservation{}
	iitoObs := iitoObservation{}

	obs := observation{
		iigoObs: &iigoObs,
		iifoObs: &iifoObs,
		iitoObs: &iitoObs,
	}

	judgeHistory := accountabilityHistory{
		history: map[uint]map[shared.ClientID]judgeHistoryInfo{},
		updated: false,
	}

	emptyRuleCache := map[string]rules.RuleMatrix{}
	trustMatrix := trust{
		trustMap: map[shared.ClientID]float64{},
	}
	trustMatrix.initialise()

	importancesMatrix := importances{
		requestAllocationImportance:                mat.NewVecDense(6, []float64{5.0, 1.0, -1.0, -1.0, 5.0, 1.0}),
		commonPoolResourceRequestImportance:        mat.NewVecDense(6, []float64{4.0, 1.0, -1.0, -1.0, 1.0, 1.0}),
		resourceReportImportance:                   mat.NewVecDense(6, []float64{5.0, 5.0, -5.0, -5.0, 1.0, 5.0}),
		getTaxContributionImportance:               mat.NewVecDense(4, []float64{-2.0, -2.0, 4.0, 1.0}),
		decideIIGOMonitoringAnnouncementImportance: mat.NewVecDense(3, []float64{1.0, -1.0, 1.0}),
		getGiftRequestsImportance:					mat.NewVecDense(4, []float64{2.0, 1.0, -1.0, -1.0}),
	}

	team4client := client{
		BaseClient:         baseclient.NewClient(id),
		clientJudge:        judge{BaseJudge: &baseclient.BaseJudge{}, t: testing},
		clientSpeaker:      speaker{BaseSpeaker: &baseclient.BaseSpeaker{}},
		obs:                &obs,
		internalParam:      &internalConfig,
		idealRulesCachePtr: &emptyRuleCache,
		savedHistory:       &judgeHistory,
		trustMatrix:        &trustMatrix,
		importances:        &importancesMatrix,
	}

	team4client.updateParents()

	return team4client
}

// NewClient is a function that creates a new empty client
func NewClient(clientID shared.ClientID) baseclient.Client {
	team4client := newClientInternal(clientID, nil)
	return &team4client
}

type client struct {
	*baseclient.BaseClient //client struct has access to methods and fields of the BaseClient struct which implements implicitly the Client interface.

	//custom fields
	clientJudge        judge
	clientSpeaker      speaker
	obs                *observation        //observation is the raw input into our client
	internalParam      *internalParameters //internal parameter store the useful parameters for the our agent
	idealRulesCachePtr *map[string]rules.RuleMatrix
	savedHistory       *accountabilityHistory
	trustMatrix        *trust
	importances        *importances
}

type importances struct {
	requestAllocationImportance                *mat.VecDense
	commonPoolResourceRequestImportance        *mat.VecDense
	resourceReportImportance                   *mat.VecDense
	getTaxContributionImportance               *mat.VecDense
	decideIIGOMonitoringAnnouncementImportance *mat.VecDense
	getGiftRequestsImportance				   *mat.VecDense
}

// Store extra information which is not in the server and is helpful for our client
type observation struct {
	iigoObs           *iigoObservation
	iifoObs           *iifoObservation
	iitoObs           *iitoObservation
	pastDisastersList baseclient.PastDisastersList
}

type iigoObservation struct {
	allocationGranted shared.Resources
	taxDemanded       shared.Resources
}

type iifoObservation struct {
	receivedDisasterPredictions shared.ReceivedDisasterPredictionsDict
	ourDisasterPrediction       shared.DisasterPredictionInfo
	finalDisasterPrediction     shared.DisasterPrediction
}

type iitoObservation struct {
}

// all parameters are from 0 to 1 and they determine the personality of the agent.
type internalParameters struct {
	//trustMatrix *mat.Dense //this shouldn't be in internal parameters
	greediness    float64
	selfishness   float64
	fairness      float64
	collaboration float64
	riskTaking    float64

	// Judge GetPardonIslands config
	// days left on the sanction after which we can even considering pardoning other islands
	maxPardonTime int
	// specifies the maximum sanction tier after which we will no longer consider pardoning others
	maxTierToPardon shared.IIGOSanctionsTier
	// we will only consider pardoning islands which we trust with at least this value
	minTrustToPardon float64
}

// type personality struct {
// }

//Overriding and extending the Initialise method of the BaseClient to initilise our client. This function happens after the init() function. At this point server has just initialised and the ServerReadHandle is available.
func (c *client) Initialise(serverReadHandle baseclient.ServerReadHandle) {
	c.BaseClient.Initialise(serverReadHandle)

	//custom things below, trust matrix initilised to values of 0
	c.idealRulesCachePtr = deepCopyRulesCache(c.ServerReadHandle.GetGameState().RulesInfo.AvailableRules)
	c.updateParents()
}

func (c *client) updateParents() {
	c.clientJudge.parent = c
	c.clientSpeaker.parent = c

}

func deepCopyRulesCache(AvailableRules map[string]rules.RuleMatrix) *map[string]rules.RuleMatrix {
	idealRulesCache := map[string]rules.RuleMatrix{}
	for k, v := range AvailableRules {
		idealRulesCache[k] = v
	}
	return &idealRulesCache
}

//Overriding the StartOfTurn method of the BaseClient
// func (c *client) StartOfTurn() {
// }

// GetVoteForRule returns the client's vote in favour of or against a rule.
// COMPULSORY: vote to represent your island's opinion on a rule
func (c *client) VoteForRule(ruleMatrix rules.RuleMatrix) shared.RuleVoteType {
	// TODO implement decision on voting that considers the rule
	ruleDistance := c.decideRuleDistance(ruleMatrix)
	if ruleDistance < 5 { // TODO: calibrate the distance ranges
		return shared.Reject
	} else if ruleDistance < 15 {
		return shared.Abstain
	} else if ruleDistance >= 15 {
		return shared.Approve
	}
	return shared.Abstain
}

// decideRuleDistance returns the evaluated distance for the rule given in the argument
func (c *client) decideRuleDistance(ruleMatrix rules.RuleMatrix) float64 {
	// link rules

	// find rule corresponding to the rule that you need to evaluate
	idealRuleMatrix := (*c.idealRulesCachePtr)[ruleMatrix.RuleName]

	// calculate a distance
	distance := 0.0
	for i := 0; i < ruleMatrix.AuxiliaryVector.Len(); i++ {
		currentAuxValue := ruleMatrix.AuxiliaryVector.AtVec(i)
		for j := range ruleMatrix.RequiredVariables {

			idealValue := idealRuleMatrix.ApplicableMatrix.At(i, j)
			actualValue := ruleMatrix.ApplicableMatrix.At(i, j)

			if currentAuxValue == 0 {
				// ==0 condition
				if idealValue != 0 {
					distance += math.Abs(idealValue-actualValue) / idealValue
				} else {
					distance += math.Abs(idealValue - actualValue)
				}
			} else if currentAuxValue == 1 {
				// TODO: ACTUALLY IMPLEMENT THESE CONDITIONS
				// >0 condition
				distance += math.Abs(idealValue-actualValue) / idealValue
			} else if currentAuxValue == 2 {
				// <=0 condition
				distance += math.Abs(idealValue-actualValue) / idealValue
			} else if currentAuxValue == 3 {
				// !=0 condition
				if idealValue != 0 {
					distance += math.Abs(idealValue-actualValue) / idealValue
				} else {
					distance += math.Abs(idealValue - actualValue)
				}
			} else if currentAuxValue == 4 {
				distance += math.Abs(idealValue-actualValue) / idealValue
				// it returns the value of the calculation
			}
		}

	}

	return distance
}

// GetVoteForElection returns the client's Borda vote for the role to be elected.
// COMPULSORY: use opinion formation to decide a rank for islands for the role
func (c *client) VoteForElection(roleToElect shared.Role, candidateList []shared.ClientID) []shared.ClientID {

	trustToID := map[float64]shared.ClientID{}
	trustList := []float64{}
	returnList := []shared.ClientID{}
	for i := 0; i < len(candidateList); i++ {
		trustScore := c.trustMatrix.GetClientTrust(candidateList[i]) //c.internalParam.agentsTrust[candidateList[i]]
		trustToID[trustScore] = candidateList[i]
		trustList = append(trustList, trustScore)
	}
	sort.Float64s(trustList)

	for i := len(trustList) - 1; i >= 0; i-- {
		// The idea is to have the very untrusted island to split the points in order
		// to increase the gap with good islands that we include and that we want to be elected.
		if trustList[i] > 0.25 || (len(trustList)-1)-i < 2 { //TODO: calibrate the trustScore so we don't always not rank //currently the infra does not support not ranking someone
			returnList = append(returnList, trustToID[trustList[i]])
		}
	}

	return returnList
}

func (c *client) StartOfTurn() {
	c.updateTrustFromSavedHistory()
}

func (c *client) updateTrustFromSavedHistory() {
	if c.savedHistory.updated {

	}
}

//MonitorIIGORole decides whether to perform monitoring on a role
//COMPULOSRY: must be implemented
func (c *client) MonitorIIGORole(roleName shared.Role) bool {

	presidentID := c.getPresident()
	speakerID := c.getSpeaker()
	judgeID := c.getJudge()
	clientID := id
	ourResources:= c.getOurResources()
	// TODO: Choose sensible thresholds!
	trustThreshold := 0.5
	resourcesThreshold := shared.Resources(100)
	monitoring := false
	switch clientID {
	case presidentID:
		// If we are the president.
		monitoring = (c.getTrust(speakerID) < trustThreshold ||
			c.getTrust(judgeID) < trustThreshold) &&
			(ourResources > resourcesThreshold)

	case speakerID:
		// If we are the Speaker.
		monitoring = (c.getTrust(presidentID) < trustThreshold ||
			c.getTrust(judgeID) < trustThreshold) &&
			(ourResources > resourcesThreshold)
	case judgeID:
		// If we are the Judge.
		monitoring = (c.getTrust(speakerID) < trustThreshold ||
			c.getTrust(judgeID) < trustThreshold) &&
			(ourResources > resourcesThreshold)
	}
	return monitoring
}

//DecideIIGOMonitoringAnnouncement decides whether to share the result of monitoring a role and what result to share
//COMPULSORY: must be implemented
func (c *client) DecideIIGOMonitoringAnnouncement(monitoringResult bool) (resultToShare bool, announce bool) {
	collaborationThreshold := 0.5
	importance := c.importances.decideIIGOMonitoringAnnouncementImportance

	parameters := mat.NewVecDense(3, []float64{
		c.internalParam.selfishness,
		c.internalParam.fairness,
		c.internalParam.collaboration,
	})
	// Initialise Return values.
	announce = false
	resultToShare = monitoringResult

	// Calculate collaborationLevel based on the current personality of the client.
	collaborationLevel := mat.Dot(importance, parameters)

	if collaborationLevel > collaborationThreshold {
		// announce only if we are collaborative enough.
		announce = true
	}
	return resultToShare, announce
}
