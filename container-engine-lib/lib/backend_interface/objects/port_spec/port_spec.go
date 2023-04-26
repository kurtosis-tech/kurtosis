package port_spec

import (
	"github.com/kurtosis-tech/stacktrace"
)

type PortSpec struct {
	number              uint16
	transportProtocol   TransportProtocol
	applicationProtocol *string
	wait                *Wait
}

/*
	This method accepts port number, transportProtocol, and application protocol (which is optional), and port wait
*/
func NewPortSpec(
	number uint16,
	transportProtocol TransportProtocol,
	maybeApplicationProtocol string,
	wait *Wait,
) (*PortSpec, error) {
	var appProtocol *string
	if maybeApplicationProtocol != "" {
		appProtocol = &maybeApplicationProtocol
	}

	// throw an error if the method receives more than 3 parameters.
	if !transportProtocol.IsATransportProtocol() {
		return nil, stacktrace.NewError("Unrecognized transportProtocol '%v'", transportProtocol.String())
	}

	portSpec := &PortSpec{
		number:              number,
		transportProtocol:   transportProtocol,
		applicationProtocol: appProtocol,
		wait:                wait,
	}

	return portSpec, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetTransportProtocol() TransportProtocol {
	return spec.transportProtocol
}

func (spec *PortSpec) GetMaybeApplicationProtocol() *string {
	return spec.applicationProtocol
}

func (spec *PortSpec) GetWait() *Wait {
	return spec.wait
}
