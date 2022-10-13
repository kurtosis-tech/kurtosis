package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
	"strings"
)

const (
	portSpecTypeName = "port_spec"
	portNumberAttr   = "number"
	portProtocolAttr = "protocol"
)

// PortSpec A starlark.Value that represents a port number & protocol
// TODO add a Make method so that this can be added as a built in
// TODO use this in the add_service primitive while passing service config
type PortSpec struct {
	number   starlark.Int
	protocol starlark.String
}

func NewPortSpec(number starlark.Int, protocol starlark.String) *PortSpec {
	return &PortSpec{
		number:   number,
		protocol: protocol,
	}
}

// String the starlark.Value interface
func (ps *PortSpec) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(portSpecTypeName + "(")
	buffer.WriteString(portNumberAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", ps.number))
	buffer.WriteString(portProtocolAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", ps.protocol))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (ps *PortSpec) Type() string {
	return portSpecTypeName
}

// Freeze implements the starlark.Value interface
func (ps *PortSpec) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (ps *PortSpec) Truth() starlark.Bool {
	return ps.protocol != "" && ps.number != starlark.MakeUint(0)
}

// Hash implements the starlark.Value interface
// TODO maybe implement this, otherwise this can't be used as a key to a dictionary
func (ps *PortSpec) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", portSpecTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *PortSpec) Attr(name string) (starlark.Value, error) {
	switch name {
	case portNumberAttr:
		return ps.number, nil
	case portProtocolAttr:
		return ps.protocol, nil
	default:
		return nil, fmt.Errorf("'%v' has no attribute '%v", portSpecTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PortSpec) AttrNames() []string {
	return []string{portNumberAttr, portProtocolAttr}
}
