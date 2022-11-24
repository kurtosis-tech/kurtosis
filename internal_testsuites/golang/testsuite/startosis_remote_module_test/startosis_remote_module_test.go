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
	dataStoreService0Id   = "datastore-0"
	dataStoreService1Id   = "datastore-1"
	datastorePortId       = "grpc"
)

func TestStartosisRemoteModule(t *testing.T) {
	// TODO(gb): Re-activate post merge
	t.Skip("Skipping as the remote module still uses import_types")
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Startosis module: '%v'", remoteModule)

	outputStream, _, err := enclaveCtx.ExecuteKurtosisRemoteModule(ctx, remoteModule, executeParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, _ := test_helpers.ReadStreamContentUntilClosed(outputStream)

	require.Nil(t, interpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Empty(t, executionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis Module")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService0Id, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService0Id,
	)
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService1Id, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService1Id,
	)
	logrus.Infof("All services added via the module work as expected")
}
