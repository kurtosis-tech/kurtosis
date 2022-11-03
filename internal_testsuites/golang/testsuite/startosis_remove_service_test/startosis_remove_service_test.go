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

	serviceId = "example-datastore-server-1"
	portId    = "grpc"

	startosisScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

service_config = struct(
    container_image_name = DATASTORE_IMAGE,
    used_ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    }
)

add_service(service_id = DATASTORE_SERVICE_ID, service_config = service_config)
print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")
`
	// We remove the service we created through the script above with a different script
	removeScript = `
DATASTORE_SERVICE_ID = "` + serviceId + `"
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
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", startosisScript)

	executionResult, err := enclaveCtx.ExecuteStartosisScript(startosisScript)
	require.NoError(t, err, "Unexpected error executing startosis script")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service example-datastore-server-1 deployed successfully.
Deployed example-datastore-server-2 successfully
`
	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Lenf(t, executionResult.ValidationErrors, 0, "Unexpected validation error")
	require.Empty(t, executionResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Infof("Successfully ran Startosis script")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is healthy",
		serviceId,
	)

	logrus.Infof("All services added via the module work as expected")

	// we run the remove script and see if things still work
	executionResult, err = enclaveCtx.ExecuteStartosisScript(removeScript)
	require.NoError(t, err, "Unexpected error executing remove script")
	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error")
	require.Lenf(t, executionResult.ValidationErrors, 0, "Unexpected validation error")
	require.Empty(t, executionResult.ExecutionError, "Unexpected execution error")

	require.Error(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is not healthy",
		serviceId,
	)
}
