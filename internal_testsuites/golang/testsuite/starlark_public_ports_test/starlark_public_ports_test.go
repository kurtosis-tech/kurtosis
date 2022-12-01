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

	serviceId           = "example-datastore-server-1"
	portId              = "grpc"
	publicPortNumberStr = "11323"
	publicPortNumber    = uint16(11323)

	starlarkScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PUBLIC_PORT_NUMBER = ` + publicPortNumberStr + `
DATASTORE_PORT_PROTOCOL = "TCP"

def run(args):
	print("Adding service " + DATASTORE_SERVICE_ID + ".")
	
	config = struct(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
		},
		public_ports = {
			DATASTORE_PORT_ID: struct(number = DATASTORE_PUBLIC_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	add_service(service_id = DATASTORE_SERVICE_ID, config = config)
	print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")
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

	outputStream, _, err := enclaveCtx.RunStarlarkScript(ctx, starlarkScript, emptyArgs, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing Starlark script")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with internal ID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
`
	require.Nil(t, interpretationError, "Unexpected interpretation error")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, scriptOutput)
	logrus.Infof("Successfully ran Starlark script")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that service is healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is healthy",
		serviceId,
	)
	logrus.Infof("All services added via the module work as expected")

	// Check that the right port got exposed
	logrus.Infof("Checking the right port got exposed on " + serviceId)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exposedPort, found := serviceCtx.GetPublicPorts()[portId]
	require.True(t, found)
	require.Equal(t, publicPortNumber, exposedPort.GetNumber())
}
