/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package services

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"math"
)

// Use a type alias here to make this a bit more user-friendly
type TransportProtocol kurtosis_core_rpc_api_bindings.Port_TransportProtocol

const (
	MaxPortNum               = math.MaxUint16
	TransportProtocol_TCP    = TransportProtocol(kurtosis_core_rpc_api_bindings.Port_TCP)
	TransportProtocol_UDP    = TransportProtocol(kurtosis_core_rpc_api_bindings.Port_UDP)
	emptyApplicationProtocol = ""
)

// "Set" of allowed port protocols
var allowedTransportProtocols = map[TransportProtocol]bool{
	TransportProtocol_TCP: true,
	TransportProtocol_UDP: true,
}

func (protocol TransportProtocol) IsValid() bool {
	_, found := allowedTransportProtocols[protocol]
	return found
}

type PortSpec struct {
	number                   uint16
	transportProtocol        TransportProtocol
	maybeApplicationProtocol string
}

func NewPortSpec(number uint16, transportProtocol TransportProtocol, maybeApplicationProtocol string) *PortSpec {
	return &PortSpec{number: number, transportProtocol: transportProtocol, maybeApplicationProtocol: maybeApplicationProtocol}
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetTransportProtocol() TransportProtocol {
	return spec.transportProtocol
}

func (spec *PortSpec) String() string {
	if spec.maybeApplicationProtocol == emptyApplicationProtocol {
		return fmt.Sprintf("%d/%v", spec.number, spec.transportProtocol)
	}
	return fmt.Sprintf("%v:%d/%v", spec.maybeApplicationProtocol, spec.number, spec.transportProtocol)
}

func (spec *PortSpec) GetMaybeApplicationProtocol() string {
	return spec.maybeApplicationProtocol
}
