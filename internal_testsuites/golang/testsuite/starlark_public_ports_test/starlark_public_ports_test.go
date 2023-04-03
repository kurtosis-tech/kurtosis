//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because public ports aren't supported in Kubernetes backend
// TODO remove this test once we have the Kurtosis Portal, and public_ports isn't a thing

package starlark_public_ports_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "starlark_public_ports_test"
	isPartitioningEnabled = false
	defaultDryRun         = false
	emptyArgs             = "{}"

	serviceName         = "example-datastore-server-1"
	portId              = "grpc"
	publicPortNumberStr = "11323"
	publicPortNumber    = uint16(11323)

	starlarkScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_NAME = "` + serviceName + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PUBLIC_PORT_NUMBER = ` + publicPortNumberStr + `
DATASTORE_PORT_PROTOCOL = "TCP"

def run(plan):
	plan.print("Adding service " + DATASTORE_SERVICE_NAME + ".")
	
	config = ServiceConfig(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		},
		public_ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PUBLIC_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	plan.add_service(name = DATASTORE_SERVICE_NAME, config = config)
	plan.print("Service " + DATASTORE_SERVICE_NAME + " deployed successfully.")
`
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", starlarkScript)

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, starlarkScript)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with service UUID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
`
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that service is healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceName, portId),
		"Error validating datastore server '%s' is healthy",
		serviceName,
	)
	logrus.Infof("All services added via the module work as expected")

	// Check that the right port got exposed
	logrus.Infof("Checking the right port got exposed on " + serviceName)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceName)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exposedPort, found := serviceCtx.GetPublicPorts()[portId]
	require.True(t, found)
	require.Equal(t, publicPortNumber, exposedPort.GetNumber())
}
