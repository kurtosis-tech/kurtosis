/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_context

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_labels_providers"
	"github.com/kurtosis-tech/kurtosis-core/commons/object_name_providers"
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

	// A DockerManager that logs to the log passed in when the enclave was created
	dockerManager *docker_manager.DockerManager

	objNameProvider *object_name_providers.EnclaveObjectNameProvider

	objLabelProvider *object_labels_providers.EnclaveObjectLabelsProvider
}

func NewEnclaveContext(enclaveId string, networkId string, networkIpAndMask *net.IPNet, apiContainerId string, apiContainerIpAddr net.IP, apiContainerHostPortBinding *nat.PortBinding, dockerManager *docker_manager.DockerManager, objNameProvider *object_name_providers.EnclaveObjectNameProvider, objLabelProvider *object_labels_providers.EnclaveObjectLabelsProvider) *EnclaveContext {
	return &EnclaveContext{enclaveId: enclaveId, networkId: networkId, networkIpAndMask: networkIpAndMask, apiContainerId: apiContainerId, apiContainerIpAddr: apiContainerIpAddr, apiContainerHostPortBinding: apiContainerHostPortBinding, dockerManager: dockerManager, objNameProvider: objNameProvider, objLabelProvider: objLabelProvider}
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

func (enclaveCtx *EnclaveContext) GetDockerManager() *docker_manager.DockerManager {
	return enclaveCtx.dockerManager
}

func (enclaveCtx *EnclaveContext) GetObjectNameProvider() *object_name_providers.EnclaveObjectNameProvider {
	return enclaveCtx.objNameProvider
}


func (enclaveCtx *EnclaveContext) GetObjectLabelsProvider() *object_labels_providers.EnclaveObjectLabelsProvider {
	return enclaveCtx.objLabelProvider
}
