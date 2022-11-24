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
	invalidCaseNoMainInMainStarTestName = "invalid-module-missing-main"
	moduleWithNoMainInMainStarRelPath   = "../../../startosis/no-main-in-main-star"
)

func TestStartosisModule_NoMainInMainStar(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseNoMainInMainStarTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithNoMainInMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedInterpretationErr := "Evaluation error: module has no .main field or method\n\tat [3:12]: <toplevel>"
	outputStream, _, err := enclaveCtx.ExecuteKurtosisModule(ctx, moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)
	require.NotNil(t, interpretationError)
	require.Contains(t, interpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Empty(t, test_helpers.GenerateScriptOutput(instructions))
}
