package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"regexp"
	"strings"
)

const (
	portNumberAttr              = "number"
	transportProtocolAttr       = "transport_protocol"
	portApplicationProtocolAttr = "application_protocol"

	PortSpecTypeName              = "PortSpec"
	maxPortNumber                 = 65535
	minPortNumber                 = 1
	optionalStringIdentifier      = ""
	validApplicationProtocolRegex = `[a-zA-Z0-9+.-]`
)

var (
	validApplicationProtocolMatcher = regexp.MustCompile(fmt.Sprintf(`^%v*$`, validApplicationProtocolRegex))
)

// PortSpec A starlark.Value that represents a port number & protocol
type PortSpec struct {
	number                   uint32
	transportProtocol        kurtosis_core_rpc_api_bindings.Port_TransportProtocol
	maybeApplicationProtocol string
}

func NewPortSpec(number uint32, transportProtocol kurtosis_core_rpc_api_bindings.Port_TransportProtocol, maybeApplicationProtocol string) *PortSpec {
	return &PortSpec{
		number:                   number,
		transportProtocol:        transportProtocol,
		maybeApplicationProtocol: maybeApplicationProtocol,
	}
}

func MakePortSpec(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var number int
	var transportProtocol string
	var maybeApplicationProtocol string

	if err := starlark.UnpackArgs(builtin.Name(), args, kwargs,
		portNumberAttr, &number,
		makeOptional(transportProtocolAttr), &transportProtocol,
		makeOptional(portApplicationProtocolAttr), &maybeApplicationProtocol,
	); err != nil {
		return nil, startosis_errors.NewInterpretationError("Cannot construct a PortSpec from the provided arguments. Error was: \n%v", err.Error())
	}

	parsedTransportProtocol, interpretationError := parseTransportProtocol(transportProtocol)
	if interpretationError != nil {
		return nil, interpretationError
	}

	uint32Number, interpretationError := parsePortNumber(number)

	if interpretationError != nil {
		return nil, interpretationError
	}

	if interpretationError = validateApplicationProtocol(maybeApplicationProtocol); interpretationError != nil {
		return nil, interpretationError
	}

	return NewPortSpec(uint32Number, parsedTransportProtocol, maybeApplicationProtocol), nil
}

// String the starlark.Value interface
func (ps *PortSpec) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(PortSpecTypeName + "(")
	buffer.WriteString(portNumberAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", ps.number))
	buffer.WriteString(transportProtocolAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q, ", ps.transportProtocol.String()))
	buffer.WriteString(portApplicationProtocolAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q)", ps.maybeApplicationProtocol))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (ps *PortSpec) Type() string {
	return PortSpecTypeName
}

// Freeze implements the starlark.Value interface
func (ps *PortSpec) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (ps *PortSpec) Truth() starlark.Bool {
	return ps.number != 0
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed, users should use a portId instead
func (ps *PortSpec) Hash() (uint32, error) {
	return 0, startosis_errors.NewInterpretationError("unhashable type: '%v'", PortSpecTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *PortSpec) Attr(name string) (starlark.Value, error) {
	switch name {
	case portNumberAttr:
		return starlark.MakeInt(int(ps.number)), nil
	case transportProtocolAttr:
		return starlark.String(ps.transportProtocol.String()), nil
	case portApplicationProtocolAttr:
		return starlark.String(ps.maybeApplicationProtocol), nil
	default:
		return nil, startosis_errors.NewInterpretationError("'%v' has no attribute '%v;", PortSpecTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PortSpec) AttrNames() []string {
	return []string{portNumberAttr, transportProtocolAttr, portApplicationProtocolAttr}
}

func (ps *PortSpec) ToKurtosisType() *kurtosis_core_rpc_api_bindings.Port {
	return binding_constructors.NewPort(ps.number, ps.transportProtocol, ps.maybeApplicationProtocol)
}

func parseTransportProtocol(portProtocol string) (kurtosis_core_rpc_api_bindings.Port_TransportProtocol, *startosis_errors.InterpretationError) {
	if portProtocol == optionalStringIdentifier {
		return kurtosis_core_rpc_api_bindings.Port_TCP, nil
	}

	parsedTransportProtocol, found := kurtosis_core_rpc_api_bindings.Port_TransportProtocol_value[portProtocol]
	if !found {
		return -1, startosis_errors.NewInterpretationError("Port protocol should be one of %s", strings.Join(port_spec.TransportProtocolStrings(), ", "))
	}
	return kurtosis_core_rpc_api_bindings.Port_TransportProtocol(parsedTransportProtocol), nil
}

func parsePortNumber(number int) (uint32, *startosis_errors.InterpretationError) {
	if number > maxPortNumber || number < minPortNumber {
		return 0, startosis_errors.NewInterpretationError("Port number should be in range [%d - %d]", minPortNumber, maxPortNumber)
	}
	return uint32(number), nil
}

func validateApplicationProtocol(maybeApplicationProtocol string) *startosis_errors.InterpretationError {
	if maybeApplicationProtocol == optionalStringIdentifier {
		return nil
	}

	doesApplicationProtocolContainsValidChar := validApplicationProtocolMatcher.MatchString(maybeApplicationProtocol)
	if !doesApplicationProtocolContainsValidChar {
		return startosis_errors.NewInterpretationError(
			`application protocol '%v' contains invalid character(s). It must only contain %v`,
			maybeApplicationProtocol,
			validApplicationProtocolRegex,
		)
	}

	return nil
}
