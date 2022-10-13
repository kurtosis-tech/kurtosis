package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
	"strings"
)

const (
	serviceTypeName = "service"
	ipAddressAttr   = "ip_address"
	portsAttr       = "ports"
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

// String the starlark.Value interface
func (rv *Service) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(serviceTypeName + "(")
	buffer.WriteString(ipAddressAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v, ", rv.ipAddress))
	buffer.WriteString(portsAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", rv.ports.String()))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (rv *Service) Type() string {
	return serviceTypeName
}

// Freeze implements the starlark.Value interface
func (rv *Service) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (rv *Service) Truth() starlark.Bool {
	return rv.ipAddress != "" && rv.ports != nil
}

// Hash implements the starlark.Value interface
// TODO maybe implement this, otherwise this can't be used as a key to a dictionary
func (rv *Service) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%v'", serviceTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (rv *Service) Attr(name string) (starlark.Value, error) {
	switch name {
	case ipAddressAttr:
		return rv.ipAddress, nil
	case portsAttr:
		return rv.ports, nil
	default:
		return nil, fmt.Errorf("'%v' has no attribute '%v", serviceTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (rv *Service) AttrNames() []string {
	return []string{ipAddressAttr, portsAttr}
}
