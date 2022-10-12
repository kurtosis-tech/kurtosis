package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
)

const (
	serviceTypeName = "service_type"
)

type Service struct {
	ipAddress starlark.String
	ports     *starlark.Dict
}

func NewService(ipAddress starlark.String, ports *starlark.Dict) *Service {
	return &Service{
		ipAddress: ipAddress,
		ports:     ports,
	}
}

func (rv *Service) String() string {
	return fmt.Sprintf("%v: ip_address:'%v', ports:'%v'", serviceTypeName, rv.ipAddress, rv.ports)
}

func (rv *Service) Type() string {
	return serviceTypeName
}

func (rv *Service) Freeze() {
	// this is a no-op its already immutable
}

func (rv *Service) Truth() starlark.Bool {
	return rv.ipAddress != "" && rv.ports != nil
}

func (rv *Service) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", serviceTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (rv *Service) Attr(name string) (starlark.Value, error) {
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
func (rv *Service) AttrNames() []string {
	return []string{"ip_address", "ports"}
}
