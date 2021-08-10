/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"net"
)

type EnclaveContext struct {
	networkId string
	networkName string
	networkIpAndMask *net.IPNet

	networkCtx *networks.NetworkContext
}

// TODO constructor
