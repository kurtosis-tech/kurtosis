/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package services

import "github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"

// Use a type alias here to make this a bit more user-friendly
type PortProtocol kurtosis_core_rpc_api_bindings.Port_Protocol
const (
	PortProtocol_TCP = PortProtocol(kurtosis_core_rpc_api_bindings.Port_TCP)
	PortProtocol_UDP = PortProtocol(kurtosis_core_rpc_api_bindings.Port_UDP)
)
// "Set" of allowed port protocols
var allowedPortProtocols = map[PortProtocol]bool{
	PortProtocol_TCP: true,
	PortProtocol_UDP: true,
}
func (protocol PortProtocol) IsValid() bool {
	_, found := allowedPortProtocols[protocol]
	return found
}


type PortSpec struct {
	number uint16
	protocol PortProtocol
}

func NewPortSpec(number uint16, protocol PortProtocol) *PortSpec {
	return &PortSpec{number: number, protocol: protocol}
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetProtocol() PortProtocol {
	return spec.protocol
}