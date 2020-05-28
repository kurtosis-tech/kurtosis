package services

import (
	"github.com/kurtosis-tech/kurtosis/commons/testnet"
)

type AvaService interface {
	testnet.Service

	GetStakingSocket() testnet.ServiceSocket
}
