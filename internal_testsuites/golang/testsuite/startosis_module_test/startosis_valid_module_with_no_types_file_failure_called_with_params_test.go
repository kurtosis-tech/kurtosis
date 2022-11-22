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

func TestStartosisModule_ValidModuleWithNoTypesFile_FailureCalledWithParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleNoTypeTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleNoTypeRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)
	serializedParams := `{"greetings": "Bonjour!"}`
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, serializedParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	expectedInterpretationErr := "A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid)."
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
}
