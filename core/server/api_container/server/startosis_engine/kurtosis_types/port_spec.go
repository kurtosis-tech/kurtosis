package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
)

const (
	portSpecTypeName = "port_spec"
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

func (ps *PortSpec) String() string {
	return fmt.Sprintf("%v: number:'%v', protocol:'%v'", portSpecTypeName, ps.number, ps.protocol)
}

func (ps *PortSpec) Type() string {
	return portSpecTypeName
}

func (ps *PortSpec) Freeze() {
	// this is a no-op its already immutable
}

func (ps *PortSpec) Truth() starlark.Bool {
	return ps.protocol != "" && ps.number != starlark.MakeUint(0)
}

func (ps *PortSpec) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", portSpecTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (ps *PortSpec) Attr(name string) (starlark.Value, error) {
	switch name {
	case "number":
		return ps.number, nil
	case "protocol":
		return ps.protocol, nil
	default:
		return nil, nil
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (ps *PortSpec) AttrNames() []string {
	return []string{"number", "protocol"}
}
