package kurtosis_types

import (
	"fmt"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"strings"
)

const (
	serviceTypeName = "Service"

	hostnameAttr    = "hostname"
	ipAddressAttr   = "ip_address"
	portsAttr       = "ports"
	serviceNameAttr = "name"
)

// Service is just a wrapper around a regular starlarkstruct.Struct
// It naturally inherits all its function making it a valid starlark.Value
type Service struct {
	*starlarkstruct.Struct
}

func NewService(serviceName starlark.String, hostname starlark.String, ipAddress starlark.String, ports *starlark.Dict) *Service {
	structDict := starlark.StringDict{
		serviceNameAttr: serviceName,
		hostnameAttr:    hostname,
		ipAddressAttr:   ipAddress,
		portsAttr:       ports,
	}
	return &Service{
		Struct: starlarkstruct.FromStringDict(starlark.String(serviceTypeName), structDict),
	}
}

// String manually overrides the default starlarkstruct.Struct String() function because it is wrong when
// we provide a custom constructor, which we do here
//
// See https://github.com/google/starlark-go/issues/448 for more details
func (service *Service) String() string {
	oldInvalid := fmt.Sprintf("\"%s\"(", serviceTypeName)
	newValid := fmt.Sprintf("%s(", serviceTypeName)
	return strings.Replace(service.Struct.String(), oldInvalid, newValid, 1)
}

// Type Needs to be overridden as the default for starlarkstruct.Struct always return "struct", which is dumb
func (service *Service) Type() string {
	return serviceTypeName
}
