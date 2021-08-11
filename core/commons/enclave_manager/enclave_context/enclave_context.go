/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_context

import (
	"net"
)

type EnclaveContext struct {
	networkId string
	networkName string
	networkIpAndMask *net.IPNet
	apiContainerIpAddr net.IP
}

func NewEnclaveContext(networkId string, networkName string, networkIpAndMask *net.IPNet, apiContainerIpAddr net.IP) *EnclaveContext {
	return &EnclaveContext{networkId: networkId, networkName: networkName, networkIpAndMask: networkIpAndMask, apiContainerIpAddr: apiContainerIpAddr}
}

func (enclaveCtx *EnclaveContext) GetNetworkID() string {
	return enclaveCtx.networkId
}

func (enclaveCtx *EnclaveContext) GetNetworkName() string {
	return enclaveCtx.networkName
}

func (enclaveCtx *EnclaveContext) GetNetworkIPAndMask() *net.IPNet {
	return enclaveCtx.networkIpAndMask
}

func (enclaveCtx *EnclaveContext) GetAPIContainerIPAddr() net.IP {
	return enclaveCtx.apiContainerIpAddr
}
