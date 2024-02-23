package port_spec

import (
	"encoding/json"
	"github.com/kurtosis-tech/stacktrace"
)

type PortSpec struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privatePortSpec *privatePortSpec
}

type privatePortSpec struct {
	Number              uint16
	TransportProtocol   TransportProtocol
	ApplicationProtocol *string
	Wait                *Wait
	Url                 *string
}

// This method accepts port number, transportProtocol, and application protocol (which is optional), and port wait
func NewPortSpec(
	number uint16,
	transportProtocol TransportProtocol,
	maybeApplicationProtocol string,
	wait *Wait,
	maybeUrl string,
) (*PortSpec, error) {
	var appProtocol *string
	if maybeApplicationProtocol != "" {
		appProtocol = &maybeApplicationProtocol
	}

	// throw an error if the method receives more than 3 parameters.
	if !transportProtocol.IsATransportProtocol() {
		return nil, stacktrace.NewError("Unrecognized transportProtocol '%v'", transportProtocol.String())
	}

	var url *string
	if maybeUrl != "" {
		url = &maybeUrl
	}

	internalPortSpec := &privatePortSpec{
		Number:              number,
		TransportProtocol:   transportProtocol,
		ApplicationProtocol: appProtocol,
		Wait:                wait,
		Url:                 url,
	}

	portSpecObj := &PortSpec{internalPortSpec}

	return portSpecObj, nil
}

func (spec *PortSpec) GetNumber() uint16 {
	return spec.privatePortSpec.Number
}

func (spec *PortSpec) GetTransportProtocol() TransportProtocol {
	return spec.privatePortSpec.TransportProtocol
}

func (spec *PortSpec) GetMaybeApplicationProtocol() *string {
	return spec.privatePortSpec.ApplicationProtocol
}

func (spec *PortSpec) GetUrl() *string {
	return spec.privatePortSpec.Url
}

func (spec *PortSpec) GetWait() *Wait {
	return spec.privatePortSpec.Wait
}

func (spec *PortSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(spec.privatePortSpec)
}

func (spec *PortSpec) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privatePortSpec{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	spec.privatePortSpec = unmarshalledPrivateStructPtr
	return nil
}
