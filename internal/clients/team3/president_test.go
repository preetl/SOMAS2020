package team3

// IIGO client functions testing

import (
	"reflect"
	"testing"

	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

func TestSetTaxationAmount(t *testing.T) {
	cases := []struct {
		name              string
		president         president
		declaredResources map[shared.ClientID]shared.Resources
		expected          map[shared.ClientID]shared.Resources
	}{
		{
			name: "Normal",
			president: president{c: &client{
				BaseClient: &baseclient.BaseClient{ServerReadHandle: mockServerReadHandle{gameState: gamestate.ClientGameState{
					ClientInfo: gamestate.ClientInfo{LifeStatus: shared.Alive,
						Resources: 100,
					},
					CommonPool: shared.Resources(40),
				}}},
				criticalStatePrediction: criticalStatePrediction{upperBound: 10, lowerBound: 0},
				params:                  islandParams{escapeCritcaIsland: true, selfishness: 0.3, riskFactor: 0.5, resourcesSkew: 1.3},
				trustScore: map[shared.ClientID]float64{
					0: 50,
					1: 50,
					2: 50,
					3: 50,
					4: 50,
					5: 50,
				},
				compliance: 1,
			}},
			declaredResources: map[shared.ClientID]shared.Resources{
				0: 100,
				1: 100,
				2: 100,
				3: 100,
				4: 100,
				5: 100,
			},
			expected: map[shared.ClientID]shared.Resources{
				0: 7,
				1: 10,
				2: 10,
				3: 10,
				4: 10,
				5: 10,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ans, _ := tc.president.SetTaxationAmount(tc.declaredResources)
			if !reflect.DeepEqual(ans, tc.expected) {
				t.Errorf("got %v, want %v", ans, tc.expected)
			}
		})
	}
}
