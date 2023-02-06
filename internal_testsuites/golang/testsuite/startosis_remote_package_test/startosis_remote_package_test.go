package startosis_remote_package_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "package"
	isPartitioningEnabled = false
	defaultDryRun         = false
	remotePackage         = "github.com/kurtosis-tech/datastore-army-package"
	executeParams         = `{"num_datastores": 2}`
	dataStoreService0Name = "datastore-0"
	dataStoreService1Name = "datastore-1"
	datastorePortId       = "grpc"
	defaultParallelism    = 4
)

func TestStartosisRemotePackage(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Starlark Package: '%v'", remotePackage)

	runResult, err := enclaveCtx.RunStarlarkRemotePackageBlocking(ctx, remotePackage, executeParams, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Starlark Package")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService0Name, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService0Name,
	)
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, dataStoreService1Name, datastorePortId),
		"Error validating datastore server '%s' is healthy",
		dataStoreService1Name,
	)
	logrus.Infof("All services added via the package work as expected")
}
