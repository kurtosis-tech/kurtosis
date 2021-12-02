/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_container_launcher

import "github.com/kurtosis-tech/stacktrace"

// EnclaveContainerPortProtocol
// Type representing a port that the enclave container is listening on (either privately, inside the enclave, or
//  publicy, outside the enclave), which is actually a string representing a Docker port protocol
type EnclaveContainerPortProtocol string
const (
	// NOTE: Unfortunately, Docker doesn't have public constants for the protocols it accepts
	// However, we can see all the valid protos in the 'nat.validateProto' method and we've created
	//  constants for each of them

	EnclaveContainerPortProtocol_TCP EnclaveContainerPortProtocol = "tcp"
	EnclaveContainerPortProtocol_SCTP EnclaveContainerPortProtocol = "sctp"
	EnclaveContainerPortProtocol_UDP EnclaveContainerPortProtocol = "udp"
)
// "Set" of the allowed enclave container port protocols
var AllEnclaveContainerPortProtocols = map[EnclaveContainerPortProtocol]bool{
	EnclaveContainerPortProtocol_TCP: true,
	EnclaveContainerPortProtocol_SCTP: true,
	EnclaveContainerPortProtocol_UDP: true,
}

// EnclaveContainerPort
// Represents a port (either public, on the host machine, or private, inside the enclave) that a container
//  inside an enclave is listening on
type EnclaveContainerPort struct {
	number uint16
	protocol EnclaveContainerPortProtocol
}

func NewEnclaveContainerPort(number uint16, protocol EnclaveContainerPortProtocol) (*EnclaveContainerPort, error) {
	if _, found := AllEnclaveContainerPortProtocols[protocol]; !found {
		return nil, stacktrace.NewError("Invalid enclave container port protocol '%v'", protocol)
	}
	return &EnclaveContainerPort{
		number: number,
		protocol: protocol,
	}, nil
}
func (port *EnclaveContainerPort) GetNumber() uint16 {
	return port.number
}
func (port *EnclaveContainerPort) GetProtocol() EnclaveContainerPortProtocol {
	return port.protocol
}


