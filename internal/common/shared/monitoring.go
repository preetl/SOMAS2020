package shared

import (
	"fmt"

	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	"github.com/pkg/errors"
)

type Monitor struct {
	GameState         *gamestate.GameState
	InternalIIGOCache []shared.Accountability
	TermLengths       map[shared.Role]uint
	IIGOClients       map[shared.ClientID]baseclient.Client
	Logger            shared.Logger
}

func (m *monitor) Logf(format string, a ...interface{}) {
	m.logger("[MONITORING]: %v", fmt.Sprintf(format, a...))
}

func (m *monitor) addToCache(roleToMonitorID shared.ClientID, variables []rules.VariableFieldName, values [][]float64) {
	pairs := []rules.VariableValuePair{}
	if len(variables) == len(values) {
		for index, variable := range variables {
			pairs = append(pairs, rules.MakeVariableValuePair(variable, values[index]))
		}
		m.InternalIIGOCache = append(m.InternalIIGOCache, shared.Accountability{
			ClientID: roleToMonitorID,
			Pairs:    pairs,
		})
	}
}

func (m *monitor) monitorRole(roleAccountable baseclient.Client) shared.MonitorResult {
	roleToMonitor, roleName, err := m.findRoleToMonitor(roleAccountable.GetID())
	if err == nil {
		decideToMonitor := roleAccountable.MonitorIIGORole(roleName)
		evaluationResult := false
		if decideToMonitor {
			evaluationResult = m.evaluateCache(roleToMonitor, rules.RulesInPlay)
		}

		m.Logf("Monitoring of %v result %v ", roleToMonitor, evaluationResult)

		evaluationResultAnnounce, announce := roleAccountable.DecideIIGOMonitoringAnnouncement(evaluationResult)

		//announce == decideToMonitor
		variablesToCache := []rules.VariableFieldName{rules.MonitorRoleAnnounce, rules.MonitorRoleDecideToMonitor}
		valuesToCache := [][]float64{{boolToFloat(decideToMonitor)}, {boolToFloat(announce)}}
		m.addToCache(roleAccountable.GetID(), variablesToCache, valuesToCache)

		if announce {
			//check if evalResult = o.g. evalResult
			variablesToCache := []rules.VariableFieldName{rules.MonitorRoleEvalResult, rules.MonitorRoleEvalResultDecide}
			valuesToCache := [][]float64{{boolToFloat(evaluationResult)}, {boolToFloat(evaluationResultAnnounce)}}
			m.addToCache(roleAccountable.GetID(), variablesToCache, valuesToCache)

			message := generateMonitoringMessage(roleName, evaluationResult)
			broadcastToAllIslands(m.IIGOClients, roleAccountable.GetID(), message)

			m.GameState.IIGOTurnsInPower[roleName] = m.TermLengths[roleName] + 1
		}

		result := shared.MonitorResult{Performed: decideToMonitor, Result: evaluationResult}
		return result
	}
	result := shared.MonitorResult{Performed: false, Result: false}
	return result
}

func (m *monitor) evaluateCache(roleToMonitorID shared.ClientID, ruleStore map[string]rules.RuleMatrix) bool {
	performedRoleCorrectly := true
	for _, entry := range m.InternalIIGOCache {
		if entry.ClientID == roleToMonitorID {
			variablePairs := entry.Pairs
			var rulesAffected []string
			for _, variable := range variablePairs {
				valuesToBeAdded, foundRules := rules.PickUpRulesByVariable(variable.VariableName, ruleStore, rules.VariableMap)
				if foundRules {
					rulesAffected = append(rulesAffected, valuesToBeAdded...)
				}
				rules.UpdateVariable(variable.VariableName, variable)
			}
			for _, rule := range rulesAffected {
				ret := rules.EvaluateRule(rule)
				if ret.EvalError == nil {
					performedRoleCorrectly = ret.RulePasses && performedRoleCorrectly
					if !ret.RulePasses {
						m.Logf("Rule: %v , broken by: %v", rule, roleToMonitorID)
					}
				}
			}
		}
	}
	return performedRoleCorrectly
}

func (m *monitor) findRoleToMonitor(roleAccountable shared.ClientID) (shared.ClientID, shared.Role, error) {
	switch roleAccountable {
	case m.GameState.SpeakerID:
		return m.GameState.PresidentID, shared.President, nil
	case m.GameState.PresidentID:
		return m.GameState.JudgeID, shared.Judge, nil
	case m.GameState.JudgeID:
		return m.GameState.SpeakerID, shared.Speaker, nil
	default:
		return shared.ClientID(-1), shared.Speaker, errors.Errorf("Monitoring by island that is not an IIGO Role")
	}
}

func generateMonitoringMessage(role shared.Role, result bool) map[shared.CommunicationFieldName]shared.CommunicationContent {
	returnMap := map[shared.CommunicationFieldName]shared.CommunicationContent{}

	returnMap[shared.RoleMonitored] = shared.CommunicationContent{
		T:            shared.CommunicationIIGORole,
		IIGORoleData: role,
	}
	returnMap[shared.MonitoringResult] = shared.CommunicationContent{
		T:           shared.CommunicationBool,
		BooleanData: result,
	}

	return returnMap
}

func (m *monitor) clearCache() {
	m.InternalIIGOCache = []shared.Accountability{}
}