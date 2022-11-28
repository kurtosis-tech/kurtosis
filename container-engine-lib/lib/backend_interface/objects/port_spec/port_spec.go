package port_spec

import (
	"github.com/kurtosis-tech/stacktrace"
)

type PortSpec struct {
	number              uint16
	protocol            PortProtocol
	applicationProtocol *ApplicationProtocol
}

/*
	This method accepts port number, protocol and application protocol ( which is optional)
*/
func NewPortSpec(number uint16, protocol PortProtocol, applicationProtocol ...ApplicationProtocol) (*PortSpec, error) {
	// throw an error if the method receives more than 3 parameters.
	if len(applicationProtocol) > 1 {
		return nil, stacktrace.NewError("Application Protocol can have at most 1 value")
	}

	if !protocol.IsAPortProtocol() {
		return nil, stacktrace.NewError("Unrecognized protocol '%v'", protocol.String())
	}

	portSpec := &PortSpec{
		number:   number,
		protocol: protocol,
	}

	if len(applicationProtocol) == 1 {
		portSpec.applicationProtocol = &applicationProtocol[0]
	}
	return portSpec, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetProtocol() PortProtocol {
	return spec.protocol
}

func (spec *PortSpec) GetApplicationProtocol() *ApplicationProtocol {
	return spec.applicationProtocol
}
