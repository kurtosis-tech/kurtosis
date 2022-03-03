package port_spec

import "github.com/kurtosis-tech/stacktrace"

type PortSpec struct {
	number   uint16
	protocol PortProtocol
}

func NewPortSpec(number uint16, protocol PortProtocol) (*PortSpec, error) {
	if !protocol.IsAPortProtocol() {
		return nil, stacktrace.NewError("Unrecognized protocol '%v'", protocol.String())
	}
	return &PortSpec{
		number:   number,
		protocol: protocol,
	}, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetProtocol() PortProtocol {
	return spec.protocol
}
