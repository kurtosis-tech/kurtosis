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
	invalidCaseNoMainInMainStarTestName = "invalid-module-missing-main"
	packageWithNoMainInMainStarRelPath  = "../../../startosis/no-run-in-main-star"
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

	expectedInterpretationErr := "No 'run' function found in file 'github.com/sample/sample-kurtosis-module/main.star'; a 'run' entrypoint function is required in the main.star file of any Kurtosis package"
	outputStream, _, err := enclaveCtx.RunStarlarkPackage(ctx, packageDirpath, emptyRunParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing Starlark package")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)
	require.NotNil(t, interpretationError)
	require.Contains(t, interpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Empty(t, scriptOutput)
}
