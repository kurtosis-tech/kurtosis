package services

import (
	"github.com/gmarchetti/kurtosis/commons/testnet"
)

type AvaService interface {
	testnet.Service

	GetStakingSocket() testnet.ServiceSocket
}
