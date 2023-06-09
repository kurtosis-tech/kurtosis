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

	expectedInterpretationErr := "No 'run' function found in the main file of package 'github.com/sample/sample-kurtosis-package'; a 'run' entrypoint function with the signature `run(plan, args)` or `run()` is required in the main file of any Kurtosis package"
	runResult, _ := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, useDefaultMainFile, useDefaultFunctionName, emptyRunParams, defaultDryRun, defaultParallelism)
	require.NotNil(t, runResult.InterpretationError)
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
