package startosis_module_test

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
	invalidCaseMainStarMissingTestName = "invalid-module-no-main-file"
	moduleWithNoMainStarRelPath        = "../../../startosis/no-main-star"
)

func TestStartosisModule_NoMainFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseMainStarMissingTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, moduleWithNoMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	expectedErrorContents := "An error occurred while verifying that 'main.star' exists on root of package"
	outputStream, _, err := enclaveCtx.RunStarlarkPackage(ctx, packageDirpath, emptyRunParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing package")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)
	require.NotNil(t, interpretationError)
	require.Contains(t, interpretationError.GetErrorMessage(), expectedErrorContents)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Empty(t, scriptOutput)
}
