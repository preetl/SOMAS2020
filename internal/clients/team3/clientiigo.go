package team3

import (
	// "github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	// "github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

/*
	//IIGO: COMPULSORY
	MonitorIIGORole(shared.Role) bool
	DecideIIGOMonitoringAnnouncement(bool) (bool, bool)

	GetVoteForRule(ruleName string) bool
	GetVoteForElection(roleToElect shared.Role) []shared.ClientID

	CommonPoolResourceRequest() shared.Resources
	ResourceReport() shared.Resources
	RuleProposal() string
	GetClientPresidentPointer() roles.President
	GetClientJudgePointer() roles.Judge
	GetClientSpeakerPointer() roles.Speaker
	TaxTaken(shared.Resources)
	GetTaxContribution() shared.Resources
	RequestAllocation() shared.Resources
*/

func (c *client) GetClientSpeakerPointer() roles.Speaker {
	// c.Logf("became speaker")
	return &speaker{c: c}
}

func (c *client) GetClientJudgePointer() roles.Judge {
	// c.Logf("became judge")
	return &judge{c: c}
}

func (c *client) GetClientPresidentPointer() roles.President {
	// c.Logf("became president")
	return &president{c: c}
}

//resetIIGOInfo clears the island's information regarding IIGO at start of turn
func (c *client) resetIIGOInfo() {
	c.iigoInfo.commonPoolAllocation = 0
	c.iigoInfo.taxationAmount = 0
	c.iigoInfo.ruleVotingResults = make(map[string]*ruleVoteInfo)
	c.iigoInfo.monitoringOutcomes = make(map[shared.Role]bool)
	c.iigoInfo.monitoringDeclared = make(map[shared.Role]bool)
}

// ReceiveCommunication is a function called by IIGO to pass the communication sent to the client.
// This function is overridden to receive information and update local info accordingly.
func (c *client) ReceiveCommunication(sender shared.ClientID, data map[shared.CommunicationFieldName]shared.CommunicationContent) {
	c.Communications[sender] = append(c.Communications[sender], data)

	for contentType, content := range data {
		switch contentType {
		case shared.TaxAmount:
			c.iigoInfo.taxationAmount = shared.Resources(content.IntegerData)
		case shared.AllocationAmount:
			c.iigoInfo.commonPoolAllocation = shared.Resources(content.IntegerData)
		case shared.RuleName:
			currentRuleID := content.TextData
			if _, ok := c.iigoInfo.ruleVotingResults[currentRuleID]; ok {
				c.iigoInfo.ruleVotingResults[currentRuleID].resultAnnounced = true
				c.iigoInfo.ruleVotingResults[currentRuleID].result = data[shared.RuleVoteResult].BooleanData
			} else {
				c.iigoInfo.ruleVotingResults[currentRuleID] = &ruleVoteInfo{resultAnnounced: true, result: data[shared.RuleVoteResult].BooleanData}
			}
		case shared.RoleMonitored:
			c.iigoInfo.monitoringDeclared[content.IIGORole] = true
			c.iigoInfo.monitoringOutcomes[content.IIGORole] = data[shared.MonitoringResult].BooleanData
		}
	}
}
