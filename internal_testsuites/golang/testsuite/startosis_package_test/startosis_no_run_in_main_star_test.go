package startosis_package_test

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
	invalidCaseNoMainInMainStarTestName = "invalid-package-missing-main"
	packageWithNoMainInMainStarRelPath  = "../../../starlark/no-run-in-main-star"
)

func TestStartosisPackage_NoMainInMainStar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseNoMainInMainStarTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, packageWithNoMainInMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	expectedInterpretationErr := "No 'run' function found in file 'github.com/sample/sample-kurtosis-package/main.star'; a 'run' entrypoint function with the signature `run(args)` or `run()` is required in the main.star file of any Kurtosis package"
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, emptyRunParams, defaultDryRun, defaultParallelism)
	require.Nil(t, err, "Unexpected error executing Starlark package")
	require.NotNil(t, runResult.InterpretationError)
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
