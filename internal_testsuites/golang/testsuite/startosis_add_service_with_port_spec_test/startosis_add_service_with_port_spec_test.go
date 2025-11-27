package startosis_add_service_with_port_spec_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "add-service-with-port-spec1"

	serviceName = "docker-getting-started-success"

	starlarkScriptWithPortSpecSuccess = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_NAME = "` + serviceName + `"

spec = PortSpec(number = 5000, transport_protocol = "UDP", wait = None)

def run(plan):
    plan.add_service(
        name = SERVICE_NAME,
        config = ServiceConfig(
            image = DOCKER_GETTING_STARTED_IMAGE, 
            ports = {
                "port1": PortSpec(number = 3333, wait = None),
                "port2": spec,
                "port3": PortSpec(number = 1234, transport_protocol = "TCP", application_protocol = "http", wait = None),
            }
        )
    )
    plan.print("httpd has been added successfully")`
)

func TestAddServiceWithPortSpec_Success(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, starlarkScriptWithPortSpecSuccess)
	logrus.Infof("Test Output: %v", runResult)
	require.NoError(t, err, "Unexpected error executing starlark script")

	service, err := enclaveCtx.GetServiceContext(serviceName)
	require.NoError(t, err, "Unexpected error occurred while getting service id")
	require.NotNil(t, service, "Error occurred while fetching service with name: '%v'. This may occur if service was not created")

	ports := service.GetPrivatePorts()
	require.Equal(t, services.TransportProtocol_TCP, ports["port1"].GetTransportProtocol())
	require.Equal(t, uint16(3333), ports["port1"].GetNumber())

	require.Equal(t, services.TransportProtocol_UDP, ports["port2"].GetTransportProtocol())
	require.Equal(t, uint16(5000), ports["port2"].GetNumber())

	require.Equal(t, services.TransportProtocol_TCP, ports["port3"].GetTransportProtocol())
	require.Equal(t, uint16(1234), ports["port3"].GetNumber())
	require.Equal(t, "http", ports["port3"].GetMaybeApplicationProtocol())

}
