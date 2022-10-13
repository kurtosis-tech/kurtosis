package kurtosis_types

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestService_StringRepresentation(t *testing.T) {
	ipAddressStr := "{{kurtosis:service_id.ip_address}}"
	ports := starlark.NewDict(1)
	err := ports.SetKey(starlark.String("grpc"), NewPortSpec(starlark.MakeInt(123), "TCP"))
	require.Nil(t, err)
	service := NewService(starlark.String(ipAddressStr), ports)
	expectedStr := `service(ip_address="{{kurtosis:service_id.ip_address}}", ports={"grpc": port_spec(number=123, protocol="TCP")})`
	require.Equal(t, expectedStr, service.String())
}
