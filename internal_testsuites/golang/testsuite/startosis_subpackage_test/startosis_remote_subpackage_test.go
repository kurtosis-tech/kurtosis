package startosis_subpackage_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	remoteTestName = "subpackage-remote"
	remotePackage  = "github.com/kurtosis-tech/examples/quickstart"
	emptyParams    = "{}"
)

func TestStarlarkRemotePackage(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP remoteTestName
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, remoteTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.Nil(t, err, "Unexpected Error Occurred")
	}()
	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Starlark Package: '%v'", remotePackage)

	runResult, err := enclaveCtx.RunStarlarkRemotePackageBlocking(ctx, remotePackage, emptyParams, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, "Hello world!\n", string(runResult.RunOutput))

	logrus.Infof("")
}
