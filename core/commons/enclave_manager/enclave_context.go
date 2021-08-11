/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_manager

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"net"
)

type EnclaveContext struct {
	networkId string
	networkName string
	networkIpAndMask *net.IPNet

	// TODO We'd like to remove this, so that all interactions go through the networkCtx
	//  However, this isn't possible as of 2021-08-11 because the initializer container needs to stream the testsuite
	//  container's logs, there's no way to stream a container's logs via the API container, and building that functionality
	//  is going to be very complex.
	dockerManager *docker_manager.DockerManager

	// Connection to API container
	networkCtx *networks.NetworkContext
}

func NewEnclaveContext(networkId string, networkName string, networkIpAndMask *net.IPNet, networkCtx *networks.NetworkContext) *EnclaveContext {
	return &EnclaveContext{networkId: networkId, networkName: networkName, networkIpAndMask: networkIpAndMask, networkCtx: networkCtx}
}
