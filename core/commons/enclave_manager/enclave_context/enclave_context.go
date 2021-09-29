/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_context

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"net"
)

// Package class containing information about a Kurtosis enclave
type EnclaveContext struct {
	enclaveId string
	networkId string
	networkIpAndMask *net.IPNet
	apiContainerId string
	apiContainerIpAddr net.IP
	apiContainerHostPortBinding *nat.PortBinding

	// A preallocated IP for a REPL (if any) that can be created inside the enclave
	// NOTE: this is kiiiinda janky; a better way would be having the API container create it
	replContainerIpAddr net.IP

	// TODO This is a gigantic, jankalicious hack!!!!! The problem:
	//  1. The initializer container needs to stream logs from the testsuite container, running inside the enclave
	//  2. The only way we currently have to stream container logs is via DockerManager using the container ID - we'd like
	// 		to be able to do it through the API container, but that's going to be a very tricky problem
	//  3. Once the API container is created, it keeps the canonical state about what IP addresses are available (which is necessary
	//  	because enclaves can live beyond any execution of Kurtosis)
	//  4. The testsuite container is created after the API container because the testsuite container needs a connection to the API container
	//  5. So after the enclave network + API container are created, if we create the testsuite container directly (using the docker manager)
	//		then the API container will think that the IP we used for the testsuite container is free. If we create the testsuite
	//      container using the API container (perhaps using a hypothetical LoadModule call), then the IP bookkeeping will be correct
	//      but the API container would either need to return a container ID (breaks the abstraction) or support log-streaming (which
	//		is the long-term solution but is a tricky problem to solve)
	//  Our workaround is to pre-allocate an IP for the testsuite container which can be used by the testing framework. This isn't ideal
	//   because it builds a knowledge about the testing framework into enclaves.
	//  We can get rid of this once we have a container log-streaming endpoint in Kurtosis Client NetworkContext.
	testsuiteContainerIpAddr net.IP

	// TODO This suffers from the same problem as container IP - the testsuite container _should_ just be a module loaded
	//  via an API container "LoadModule" call. Right now it's "special" in that it's started by teh initializer, and we can't
	//  make it unspecial until the API container supports log streaming
	testsuiteContainerName string

	testsuiteContainerLabels map[string]string

	// A DockerManager that logs to the log passed in when the enclave was created
	dockerManager *docker_manager.DockerManager
}

func NewEnclaveContext(enclaveId string, networkId string, networkIpAndMask *net.IPNet, apiContainerId string, apiContainerIpAddr net.IP, apiContainerHostPortBinding *nat.PortBinding, replContainerIpAddr net.IP, testsuiteContainerIpAddr net.IP, testsuiteContainerName string, testsuiteContainerLabels map[string]string, dockerManager *docker_manager.DockerManager) *EnclaveContext {
	return &EnclaveContext{enclaveId: enclaveId, networkId: networkId, networkIpAndMask: networkIpAndMask, apiContainerId: apiContainerId, apiContainerIpAddr: apiContainerIpAddr, apiContainerHostPortBinding: apiContainerHostPortBinding, replContainerIpAddr: replContainerIpAddr, testsuiteContainerIpAddr: testsuiteContainerIpAddr, testsuiteContainerName: testsuiteContainerName, testsuiteContainerLabels: testsuiteContainerLabels, dockerManager: dockerManager}
}

func (enclaveCtx *EnclaveContext) GetEnclaveID() string {
	return enclaveCtx.enclaveId
}

func (enclaveCtx *EnclaveContext) GetNetworkID() string {
	return enclaveCtx.networkId
}

func (enclaveCtx *EnclaveContext) GetNetworkIPAndMask() *net.IPNet {
	return enclaveCtx.networkIpAndMask
}

func (enclaveCtx *EnclaveContext) GetAPIContainerID() string {
	return enclaveCtx.apiContainerId
}

func (enclaveCtx *EnclaveContext) GetAPIContainerIPAddr() net.IP {
	return enclaveCtx.apiContainerIpAddr
}

func (enclaveCtx *EnclaveContext) GetAPIContainerHostPortBinding() *nat.PortBinding {
	return enclaveCtx.apiContainerHostPortBinding
}

func (enclaveCtx *EnclaveContext) GetREPLContainerIPAddr() net.IP {
	return enclaveCtx.replContainerIpAddr
}

func (enclaveCtx *EnclaveContext) GetTestsuiteContainerIPAddr() net.IP {
	return enclaveCtx.testsuiteContainerIpAddr
}

func (enclaveCtx *EnclaveContext) GetTestsuiteContainerName() string {
	return enclaveCtx.testsuiteContainerName
}

func (enclaveCtx *EnclaveContext) GetDockerManager() *docker_manager.DockerManager {
	return enclaveCtx.dockerManager
}
