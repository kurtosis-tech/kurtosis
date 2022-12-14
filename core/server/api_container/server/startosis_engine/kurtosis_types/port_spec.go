package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	portNumberAttr   = "number"
	portProtocolAttr = "protocol"
	PortSpecTypeName = "PortSpec"

	optionalPortProtocolAttr = "protocol?"
	maxPortNumber            = 65535
	minPortNumber            = 1

	emptyProtocol = ""
)

// PortSpec A starlark.Value that represents a port number & protocol
// TODO add a Make method so that this can be added as a built in
// TODO use this in the add_service primitive while passing service config
type PortSpec struct {
	number   uint32
	protocol kurtosis_core_rpc_api_bindings.Port_Protocol
}

func NewPortSpec(number uint32, portProtocol kurtosis_core_rpc_api_bindings.Port_Protocol) *PortSpec {
	return &PortSpec{
		number:   number,
		protocol: portProtocol,
	}
}

func MakePortSpec() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var number int
		var protocol string

		if err := starlark.UnpackArgs(builtin.Name(), args, kwargs, portNumberAttr, &number, optionalPortProtocolAttr, &protocol); err != nil {
			return nil, startosis_errors.NewInterpretationError("Cannot construct a PortSpec from the provided arguments. Error was: \n%v", err.Error())
		}

		portProtocol, interpretationError := parsePortProtocol(protocol)
		if interpretationError != nil {
			return nil, interpretationError
		}

		uint32Number, interpretationError := parsePortNumber(number)
		if interpretationError != nil {
			return nil, interpretationError
		}

		return NewPortSpec(uint32Number, portProtocol), nil
	}
}

// String the starlark.Value interface
func (ps *PortSpec) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(PortSpecTypeName + "(")
	buffer.WriteString(portNumberAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", ps.number))
	buffer.WriteString(portProtocolAttr + "=")
	buffer.WriteString(fmt.Sprintf("%q)", ps.protocol.String()))
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
	return 0, fmt.Errorf("unhashable type: '%v'", PortSpecTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *PortSpec) Attr(name string) (starlark.Value, error) {
	switch name {
	case portNumberAttr:
		return starlark.MakeInt(int(ps.number)), nil
	case portProtocolAttr:
		return starlark.String(ps.protocol.String()), nil
	default:
		return nil, fmt.Errorf("'%v' has no attribute '%v;", PortSpecTypeName, name)
	}
}

func (ps *PortSpec) GetNumber() uint32 {
	return ps.number
}

func (ps *PortSpec) GetProtocol() kurtosis_core_rpc_api_bindings.Port_Protocol {
	return ps.protocol
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PortSpec) AttrNames() []string {
	return []string{portNumberAttr, portProtocolAttr}
}

func parsePortProtocol(portProtocol string) (kurtosis_core_rpc_api_bindings.Port_Protocol, *startosis_errors.InterpretationError) {
	if portProtocol == emptyProtocol {
		return kurtosis_core_rpc_api_bindings.Port_TCP, nil
	}

	parsedPortProtocol, found := kurtosis_core_rpc_api_bindings.Port_Protocol_value[portProtocol]
	if !found {
		return -1, startosis_errors.NewInterpretationError("Port protocol should be one of %s", strings.Join(port_spec.PortProtocolStrings(), ", "))
	}
	return kurtosis_core_rpc_api_bindings.Port_Protocol(parsedPortProtocol), nil
}

func parsePortNumber(number int) (uint32, *startosis_errors.InterpretationError) {
	if number > maxPortNumber || number < minPortNumber {
		return 0, startosis_errors.NewInterpretationError("Port number should be in range [%d - %d]", minPortNumber, maxPortNumber)
	}
	return uint32(number), nil
}
