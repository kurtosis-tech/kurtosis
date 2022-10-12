package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
)

const (
	addServiceReturnValueTypeName = "add_service_return_type"
)

type AddServiceInstructionReturnType struct {
	ipAddress starlark.String
	ports     *starlark.Dict
}

func NewAddServiceInstructionReturnType(ipAddress starlark.String, ports *starlark.Dict) *AddServiceInstructionReturnType {
	return &AddServiceInstructionReturnType{
		ipAddress: ipAddress,
		ports:     ports,
	}
}

func (rv *AddServiceInstructionReturnType) String() string {
	return fmt.Sprintf("%v: ip_address:'%v', ports:'%v'", addServiceReturnValueTypeName, rv.ipAddress, rv.ports)
}

func (rv *AddServiceInstructionReturnType) Type() string {
	return addServiceReturnValueTypeName
}

func (rv *AddServiceInstructionReturnType) Freeze() {
	// this is a no-op its already immutable
}

func (rv *AddServiceInstructionReturnType) Truth() starlark.Bool {
	return rv.ipAddress != "" && rv.ports != nil
}

func (rv *AddServiceInstructionReturnType) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", addServiceReturnValueTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (rv *AddServiceInstructionReturnType) Attr(name string) (starlark.Value, error) {
	switch name {
	case "ip_address":
		return rv.ipAddress, nil
	case "ports":
		return rv.ports, nil
	default:
		return nil, nil
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (rv *AddServiceInstructionReturnType) AttrNames() []string {
	return []string{"ip_address", "ports"}
}
