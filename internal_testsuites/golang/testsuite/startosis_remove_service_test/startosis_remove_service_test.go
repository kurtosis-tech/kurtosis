package startosis_remove_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "startosis_remove_service_test"
	isPartitioningEnabled = false
	defaultDryRun         = false
	emptyArgs             = "{}"

	serviceId = "example-datastore-server-1"
	portId    = "grpc"

	starlarkScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

def run(args):
	print("Adding service " + DATASTORE_SERVICE_ID + ".")
	
	config = struct(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	add_service(service_id = DATASTORE_SERVICE_ID, config = config)
	print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")
`
	// We remove the service we created through the script above with a different script
	removeScript = `
DATASTORE_SERVICE_ID = "` + serviceId + `"
def run(args):
	remove_service(DATASTORE_SERVICE_ID)
`
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Starlark script to first add the datastore service...")
	logrus.Debugf("Starlark script content: \n%v", starlarkScript)

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, emptyArgs, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with service GUID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
`
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script to add datastore service")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is healthy",
		serviceId,
	)

	logrus.Infof("Validated that all services are healthy")

	// we run the remove script and see if things still work
	runResult, err = enclaveCtx.RunStarlarkScriptBlocking(ctx, removeScript, emptyArgs, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing remove script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput = `Service 'example-datastore-server-1' with service GUID '[a-z-0-9]+' removed
`
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))

	require.Error(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is not healthy",
		serviceId,
	)

	// Ensure that service listing is empty too
	serviceInfos, err := enclaveCtx.GetServices()
	require.Nil(t, err)
	require.Empty(t, serviceInfos)
}
