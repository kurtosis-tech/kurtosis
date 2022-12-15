package add_service_with_application_protocol_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName               = "add-service-with-port-spec"
	isPartitioningEnabled  = false
	serviceIdStr           = "service-with-port-spec"
	portWithAppProtocol    = "port1"
	portWithOutAppProtocol = "port2"
	dockerImage            = "docker/getting-started:latest"
)

func TestAddServiceWithApplicationProtocol(t *testing.T) {
	ctx := context.Background()
	serviceId := services.ServiceID(serviceIdStr)
	// ENGINE SETUP
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	defer destroyEnclaveFunc()
	require.NoError(t, err, "An error occurred creating an enclave")

	portSpecMap := map[string]*services.PortSpec{
		portWithAppProtocol:    services.NewPortSpec(4444, services.TransportProtocol_UDP, "http"),
		portWithOutAppProtocol: services.NewPortSpec(3333, services.TransportProtocol_TCP, ""),
	}

	containerConfig := services.NewContainerConfigBuilder(dockerImage).WithUsedPorts(portSpecMap)
	containerConfigs := map[services.ServiceID]*services.ContainerConfig{
		serviceId: containerConfig.Build(),
	}

	successServices, failureServices, err := enclaveCtx.AddServices(containerConfigs)
	require.NoError(t, err, "An unexpected error occurred while adding serivce")
	require.Empty(t, failureServices)

	service, found := successServices[serviceId]
	require.True(t, found)
	ports := service.GetPrivatePorts()

	portSpecWithAppProtocol := ports[portWithAppProtocol]
	portSpecWithoutAppProtocol := ports[portWithOutAppProtocol]

	require.Equal(t, uint16(4444), portSpecWithAppProtocol.GetNumber())
	require.Equal(t, "http", portSpecWithAppProtocol.GetMaybeApplicationProtocol())
	require.Equal(t, services.TransportProtocol_UDP, portSpecWithAppProtocol.GetTransportProtocol())

	require.Equal(t, uint16(3333), portSpecWithoutAppProtocol.GetNumber())
	require.Equal(t, "", portSpecWithoutAppProtocol.GetMaybeApplicationProtocol())
	require.Equal(t, services.TransportProtocol_TCP, portSpecWithoutAppProtocol.GetTransportProtocol())
}
