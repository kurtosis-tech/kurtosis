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
	invalidModuleNoTypeButInputArgsTestName = "invalid-module-no-type-input-args"
	invalidModuleNoTypeButInputArgsRelPath  = "../../../startosis/invalid-no-type-but-input-args"
)

func TestStartosisModule_InvalidModuleNoTypesButInputArgsTestName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidModuleNoTypeButInputArgsTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, invalidModuleNoTypeButInputArgsRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	outputStream, _, err := enclaveCtx.ExecuteKurtosisModule(ctx, moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)
	require.NotNil(t, interpretationError)
	expectedInterpretationErr := "Evaluation error: function run missing 1 argument (input_args)"
	require.Contains(t, interpretationError.GetErrorMessage(), expectedInterpretationErr)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Empty(t, test_helpers.GenerateScriptOutput(instructions))
}
