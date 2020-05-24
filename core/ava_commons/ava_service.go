package ava_commons

import "github.com/gmarchetti/kurtosis/commons"

type AvaService interface {
	commons.Service

	GetStakingSocket() commons.ServiceSocket
}
