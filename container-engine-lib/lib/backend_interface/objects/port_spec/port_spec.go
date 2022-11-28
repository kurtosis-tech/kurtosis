package port_spec

import (
	"github.com/kurtosis-tech/stacktrace"
)

type PortSpec struct {
	number              uint16
	protocol            PortProtocol
	applicationProtocol *string
}

/*
	This method accepts port number, protocol and application protocol ( which is optional)
*/
func NewPortSpec(number uint16, protocol PortProtocol, applicationProtocols ...string) (*PortSpec, error) {
	var applicationProtocol *string

	// throw an error if the method receives more than 3 parameters.
	if len(applicationProtocols) > 1 {
		return nil, stacktrace.NewError("Application Protocol can have at most 1 value")
	}

	if !protocol.IsAPortProtocol() {
		return nil, stacktrace.NewError("Unrecognized protocol '%v'", protocol.String())
	}

	if len(applicationProtocols) == 1 {
		applicationProtocol = &applicationProtocols[0]
	}

	portSpec := &PortSpec{
		number:              number,
		protocol:            protocol,
		applicationProtocol: applicationProtocol,
	}

	return portSpec, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetProtocol() PortProtocol {
	return spec.protocol
}

func (spec *PortSpec) GetApplicationProtocol() *string {
	return spec.applicationProtocol
}
