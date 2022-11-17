package startosis_remote_module_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "module"
	isPartitioningEnabled = false
	defaultDryRun         = false
	remoteModule          = "github.com/kurtosis-tech/datastore-army-module"
	executeParams         = `{"num_datastores": "2"}`
	dataStoreService0     = "datastore-0"
	dataStoreService1     = "datastore-1"
	datastorePortId       = "grpc"
)

func TestStartosisRemoteModule(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Startosis module: \n%v", remoteModule)

	executionResult, err := enclaveCtx.ExecuteStartosisRemoteModule(remoteModule, executeParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")

	expectedScriptOutput := `Deploying module datastore_army_module with args:
ModuleInput(num_datastores=2)
Adding service datastore-0
Adding service datastore-1
Module datastore_army_module deployed successfully.
ModuleOutput(created_service_ids_to_port_ids=[ServiceIdPortId(service_id="datastore-0", port_id="grpc"), ServiceIdPortId(service_id="datastore-1", port_id="grpc")])
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
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService0, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService0,
	)
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService1, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService1,
	)
	logrus.Infof("All services added via the module work as expected")
}
