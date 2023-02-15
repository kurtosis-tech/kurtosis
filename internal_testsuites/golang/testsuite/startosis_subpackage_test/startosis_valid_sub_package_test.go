package startosis_subpackage_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

const (
	testName                    = "subpackage-module-test"
	relPathToKurtosisSubpackage = "../../../starlark/kurtosis-sub-package/subpackage-module"
)

func TestStarlarkValidLocalSubPackage(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.Nil(t, err, "Unexpected Error Occurred")
	}()
	
	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Debugf("Executing Starlark Package")

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, relPathToKurtosisSubpackage)

	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, emptyRunParams, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, "Successfully ran kurtosis package from a subfolder\n", string(runResult.RunOutput))
}
