package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
)

const (
	portSpecTypeName = "port_spec"
	portNumberAttr   = "number"
	portProtocolAttr = "protocol"
)

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
	return fmt.Sprintf("%v: number:'%v', protocol:'%v'", portSpecTypeName, ps.number, ps.protocol)
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
		return nil, nil
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PortSpec) AttrNames() []string {
	return []string{portNumberAttr, portProtocolAttr}
}
