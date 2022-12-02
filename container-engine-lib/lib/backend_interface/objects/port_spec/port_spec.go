package port_spec

import (
	"github.com/kurtosis-tech/stacktrace"
)

type PortSpec struct {
	number              uint16
	transportProtocol   PortProtocol
	applicationProtocol *string
}

/*
	This method accepts port number, transportProtocol and application transportProtocol ( which is optional)
*/
func NewPortSpec(number uint16, transportProtocol PortProtocol, maybeApplicationProtocol string) (*PortSpec, error) {
	var appProtocol *string
	if maybeApplicationProtocol != "" {
		appProtocol = &maybeApplicationProtocol
	}

	// throw an error if the method receives more than 3 parameters.
	if !transportProtocol.IsAPortProtocol() {
		return nil, stacktrace.NewError("Unrecognized transportProtocol '%v'", transportProtocol.String())
	}
	portSpec := &PortSpec{
		number:              number,
		transportProtocol:   transportProtocol,
		applicationProtocol: appProtocol,
	}

	return portSpec, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.number
}

func (spec *PortSpec) GetTransportProtocol() PortProtocol {
	return spec.transportProtocol
}

func (spec *PortSpec) GetMaybeApplicationProtocol() *string {
	return spec.applicationProtocol
}
