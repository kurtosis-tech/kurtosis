package kurtosis_types

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	ipAddressTestValue = starlark.String("{{kurtosis:service_id.ip_address}}")
	testInvalidAttr    = "invalid-test-attr"
)

func TestService_StringRepresentation(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	expectedStr := `service(ip_address="{{kurtosis:service_id.ip_address}}", ports={"grpc": port_spec(number=123, protocol="TCP")})`
	require.Equal(t, expectedStr, service.String())
}

func TestService_ServiceType(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	require.Equal(t, serviceTypeName, service.Type())
}

func TestService_Freeze(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	// just checking it doesn't panic as it's a no-op
	require.NotPanics(t, service.Freeze)
}

func TestService_TruthValidService(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	// just checking it doesn't panic as it's a no-op
	require.Equal(t, starlark.Bool(true), service.Truth())
}

func TestService_TruthFalsyService(t *testing.T) {
	service := Service{}
	require.Equal(t, starlark.Bool(false), service.Truth())
}

func TestService_HashThrowsError(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	hash, err := service.Hash()
	require.NotNil(t, err)
	require.Equal(t, uint32(0), hash)
}

func TestService_TestValidAttr(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	attrValue, err := service.Attr(ipAddressAttr)
	require.Nil(t, err)
	require.Equal(t, ipAddressTestValue, attrValue)
}

func TestService_TestInvalidAttr(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	attrValue, err := service.Attr(testInvalidAttr)
	require.NotNil(t, err)
	require.Nil(t, attrValue)
}

func TestService_TestAttrNames(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	attrNames := service.AttrNames()
	require.Equal(t, []string{ipAddressAttr, portsAttr}, attrNames)
}

func createTestServiceType() (*Service, error) {
	ports := starlark.NewDict(1)
	err := ports.SetKey(starlark.String("grpc"), NewPortSpec(starlark.MakeInt(123), "TCP"))
	if err != nil {
		return nil, err
	}
	service := NewService(ipAddressTestValue, ports)
	return service, nil
}
