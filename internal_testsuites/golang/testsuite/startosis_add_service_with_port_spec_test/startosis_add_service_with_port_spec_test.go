package startosis_add_service_with_port_spec_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "add-service-with-port-spec1"
	isPartitioningEnabled = false
	defaultDryRun         = false

	serviceId = "docker-getting-started-success"
	emptyArgs = "{}"

	starlarkScriptWithPortSpec_Success = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "` + serviceId + `"

spec = PortSpec(number = 5000, transport_protocol = "UDP")

def run(plan):
    plan.add_service(
        service_id = SERVICE_ID, 
        config = struct(
            image = DOCKER_GETTING_STARTED_IMAGE, 
            ports = {
                "port1": PortSpec(number = 3333),
                "port2": spec
            }
        )
    )
    plan.print("httpd has been added successfully")`
)

func TestAddServiceWithPortSpec_Success(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScriptWithPortSpec_Success, emptyArgs, defaultDryRun)
	logrus.Infof("Test Output: %v", runResult)
	require.NoError(t, err, "Unexpected error executing starlark script")

	service, err := enclaveCtx.GetServiceContext(serviceId)
	require.NoError(t, err, "Unexpected error occurred while getting service id")
	require.NotNil(t, service, "Error occurred while fetching service with ID: '%v'. This may occur if service was not created")

	ports := service.GetPrivatePorts()
	require.Equal(t, services.TransportProtocol_TCP, ports["port1"].GetTransportProtocol())
	require.Equal(t, uint16(3333), ports["port1"].GetNumber())

	require.Equal(t, services.TransportProtocol_UDP, ports["port2"].GetTransportProtocol())
	require.Equal(t, uint16(5000), ports["port2"].GetNumber())
}
