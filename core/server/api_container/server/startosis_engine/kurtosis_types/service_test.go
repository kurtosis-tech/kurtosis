package kurtosis_types

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	hostnameTestValue        = starlark.String("datastore-1")
	ipAddressTestValue       = starlark.String("{{kurtosis:service_name.ip_address}}")
	testInvalidAttr          = "invalid-test-attr"
	httpApplicationProtocol  = "http"
	emptyApplicationProtocol = ""
	serviceNameTestValue     = starlark.String("test-service")
	emptyPortSpecWaitTimeout = ""
	portSpecWaitTimeout      = "2s"
)

func TestService_StringRepresentation(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	expectedStr := `Service(hostname = "datastore-1", ip_address = "{{kurtosis:service_name.ip_address}}", name = "test-service", ports = {"grpc": PortSpec(number=123, transport_protocol="TCP")})`
	serviceStr := service.String()
	require.Equal(t, expectedStr, serviceStr)
}

func TestService_StringRepresentationWithApplicationProtocol(t *testing.T) {
	service, err := createTestServiceTypeWithApplicationProtocol()
	require.Nil(t, err)
	expectedStr := `Service(hostname = "datastore-1", ip_address = "{{kurtosis:service_name.ip_address}}", name = "test-service", ports = {"grpc": PortSpec(number=123, transport_protocol="TCP", application_protocol="http")})`
	require.Equal(t, expectedStr, service.String())
}

func TestService_StringRepresentationWithWait(t *testing.T) {
	service, err := createTestServiceTypeWithWait()
	require.Nil(t, err)
	expectedStr := `Service(hostname = "datastore-1", ip_address = "{{kurtosis:service_name.ip_address}}", name = "test-service", ports = {"grpc": PortSpec(number=123, transport_protocol="TCP", wait="2s")})`
	require.Equal(t, expectedStr, service.String())
}

func TestService_StringRepresentationWithApplicationProtocolAndWait(t *testing.T) {
	service, err := createTestServiceTypeWithApplicationProtocolAndWait()
	require.Nil(t, err)
	expectedStr := `Service(hostname = "datastore-1", ip_address = "{{kurtosis:service_name.ip_address}}", name = "test-service", ports = {"grpc": PortSpec(number=123, transport_protocol="TCP", application_protocol="http", wait="2s")})`
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
	// just checking it doesn't panic
	require.NotPanics(t, service.Freeze)
}

func TestService_TruthValidService(t *testing.T) {
	service, err := createTestServiceType()
	require.Nil(t, err)
	// starlarkstruct.Struct Truth() function always return true
	require.Equal(t, starlark.Bool(true), service.Truth())
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
	require.Equal(t, []string{hostnameAttr, ipAddressAttr, serviceNameAttr, portsAttr}, attrNames)
}

func createTestServiceType() (*Service, error) {
	ports := starlark.NewDict(1)
	portSpec, err := port_spec.CreatePortSpec(123, kurtosis_core_rpc_api_bindings.Port_TCP, emptyApplicationProtocol, emptyPortSpecWaitTimeout)
	if err != nil {
		return nil, err
	}
	if err := ports.SetKey(starlark.String("grpc"), portSpec); err != nil {
		return nil, err
	}
	service := NewService(serviceNameTestValue, hostnameTestValue, ipAddressTestValue, ports)
	return service, nil
}

func createTestServiceTypeWithApplicationProtocol() (*Service, error) {
	ports := starlark.NewDict(1)
	portSpec, err := port_spec.CreatePortSpec(123, kurtosis_core_rpc_api_bindings.Port_TCP, httpApplicationProtocol, emptyPortSpecWaitTimeout)
	if err != nil {
		return nil, err
	}
	if err := ports.SetKey(starlark.String("grpc"), portSpec); err != nil {
		return nil, err
	}
	service := NewService(serviceNameTestValue, hostnameTestValue, ipAddressTestValue, ports)
	return service, nil
}

func createTestServiceTypeWithWait() (*Service, error) {
	ports := starlark.NewDict(1)
	portSpec, err := port_spec.CreatePortSpec(123, kurtosis_core_rpc_api_bindings.Port_TCP, emptyApplicationProtocol, portSpecWaitTimeout)
	if err != nil {
		return nil, err
	}
	if err := ports.SetKey(starlark.String("grpc"), portSpec); err != nil {
		return nil, err
	}
	service := NewService(serviceNameTestValue, hostnameTestValue, ipAddressTestValue, ports)
	return service, nil
}

func createTestServiceTypeWithApplicationProtocolAndWait() (*Service, error) {
	ports := starlark.NewDict(1)
	portSpec, err := port_spec.CreatePortSpec(123, kurtosis_core_rpc_api_bindings.Port_TCP, httpApplicationProtocol, portSpecWaitTimeout)
	if err != nil {
		return nil, err
	}
	if err := ports.SetKey(starlark.String("grpc"), portSpec); err != nil {
		return nil, err
	}
	service := NewService(serviceNameTestValue, hostnameTestValue, ipAddressTestValue, ports)
	return service, nil
}
