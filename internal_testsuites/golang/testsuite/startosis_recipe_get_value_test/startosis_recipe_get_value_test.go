package startosis_recipe_get_value_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

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

	serviceId = "example-datastore-server-1"
	portId    = "grpc"

	startosisScript = `
value = get_value()
print(value)
add_service()
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

	executionResult, err := enclaveCtx.ExecuteStartosisScript(startosisScript, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis script")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service example-datastore-server-1 deployed successfully.
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
	executionResult, err = enclaveCtx.ExecuteStartosisScript(removeScript, defaultDryRun)
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

	// Ensure that service listing is empty too
	serviceIds, err := enclaveCtx.GetServices()
	require.Nil(t, err)
	require.Empty(t, serviceIds)
}
