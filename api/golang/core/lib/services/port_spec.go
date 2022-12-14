/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package services

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
)

// Use a type alias here to make this a bit more user-friendly
type TransportProtocol kurtosis_core_rpc_api_bindings.Port_TransportProtocol

const (
	PortProtocol_TCP = TransportProtocol(kurtosis_core_rpc_api_bindings.Port_TCP)
	PortProtocol_UDP = TransportProtocol(kurtosis_core_rpc_api_bindings.Port_UDP)
)

// "Set" of allowed port protocols
var allowedPortProtocols = map[TransportProtocol]bool{
	PortProtocol_TCP: true,
	PortProtocol_UDP: true,
}

func (protocol TransportProtocol) IsValid() bool {
	_, found := allowedPortProtocols[protocol]
	return found
}

type PortSpec struct {
	number            uint16
	transportProtocol TransportProtocol
}

func NewPortSpec(number uint16, transportProtocol TransportProtocol) *PortSpec {
	return &PortSpec{number: number, transportProtocol: transportProtocol}
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetTransportProtocol() TransportProtocol {
	return spec.transportProtocol
}

func (spec *PortSpec) String() string {
	return fmt.Sprintf("%d/%v", spec.number, spec.transportProtocol)
}
