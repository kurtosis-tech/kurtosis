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
	starlarkScript         = `
def run(plan):
	plan.add_service(
		service_id = "service-with-port-spec",
		config = struct(
			image = "docker/getting-started:latest",
			ports = {
				"port1": PortSpec(
					number = 4444,
					transport_protocol = "UDP",
					application_protocol = "http"
				),
				"port2": PortSpec(
					number = 3333,
					transport_protocol = "TCP"
				)
			},
		)
	)
`
)

func TestAddServiceWithApplicationProtocol(t *testing.T) {
	ctx := context.Background()
	serviceId := services.ServiceID(serviceIdStr)

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	defer func() {
		err := destroyEnclaveFunc()
		require.Nil(t, err)
	}()
	require.NoError(t, err, "An error occurred creating an enclave")

	// -------------------------------------- SCRIPT RUN -----------------------------------------------
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, "", false)
	require.NoError(t, err, "An unexpected error occurred while running Starlark script")
	require.Empty(t, runResult.InterpretationError, "An unexpected error occurred while interpreting Starlark script")
	require.Empty(t, runResult.ValidationErrors, "An unexpected error occurred while validating Starlark script")
	require.Empty(t, runResult.ExecutionError, "An unexpected error occurred while executing Starlark script")

	// ------------------------------------ TEST ASSERTIONS ---------------------------------------------
	service, err := enclaveCtx.GetServiceContext(serviceId)
	require.NoError(t, err, "An unexpected error occurred while getting service")
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
