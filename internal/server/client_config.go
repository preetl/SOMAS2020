package server

import (
	"github.com/SOMAS2020/SOMAS2020/internal/clients/team3"
	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

type ClientFactory func(shared.ClientID) baseclient.Client

func DefaultClientConfig() map[shared.ClientID]ClientFactory {
	return map[shared.ClientID]ClientFactory{
		shared.Team1: team3.DefaultClient,
		shared.Team2: team3.DefaultClient,
		shared.Team3: team3.DefaultClient,
		shared.Team4: team3.DefaultClient,
		shared.Team5: team3.DefaultClient,
		shared.Team6: team3.DefaultClient,
	}
}
