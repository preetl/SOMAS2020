package team3

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

// getLocalResources retrieves our islands resrouces from server
func (c *client) getLocalResources() shared.Resources {
	currentState := c.BaseClient.ServerReadHandle.GetGameState()
	return currentState.ClientInfo.Resources
}

// getIslandsAlive retrives number of islands still alive
func (c *client) getIslandsAlive() int {
	var lifeStatuses map[shared.ClientID]shared.ClientLifeStatus
	var aliveCount int

	currentState := c.BaseClient.ServerReadHandle.GetGameState()
	lifeStatuses = currentState.ClientLifeStatuses
	for _, statusInfo := range lifeStatuses {
		if statusInfo == shared.Alive {
			aliveCount++
		}
	}
	return aliveCount
}

// getAverage returns the average of the list
func getAverage(lst []float64) float64 {

	if len(lst) == 0 {
		return 0.0
	}

	total := 0.0
	for _, val := range lst {
		total += val
	}

	return (float64(total) / float64(len(lst)))
}

type mockServerReadHandle struct {
	gameState gamestate.ClientGameState
}

func (m mockServerReadHandle) GetGameState() gamestate.ClientGameState {
	return m.gameState
}
