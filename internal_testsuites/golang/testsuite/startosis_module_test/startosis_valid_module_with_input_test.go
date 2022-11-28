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
	validModuleWithInputTestName = "valid-module-with-input"
	validModuleWithInputRelPath  = "../../../startosis/valid-kurtosis-module-with-input"
)

func TestStartosisModule_ValidModuleWithInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleWithInputTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleWithInputRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	params := `{"greetings": "bonjour!"}`
	outputStream, _, err := enclaveCtx.ExecuteKurtosisModule(ctx, moduleDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `bonjour!
Hello World!
`
	require.Nil(t, interpretationError, "Unexpected interpretation error")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, test_helpers.GenerateScriptOutput(instructions))
	logrus.Info("Successfully ran Startosis module")
}

func TestStartosisModule_ValidModuleWithInput_MissingKeyInParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleWithInputTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleWithInputRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	params := `{"hello": "world"}` // expecting key 'greetings' here
	outputStream, _, err := enclaveCtx.ExecuteKurtosisModule(ctx, moduleDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)

	require.NotNil(t, interpretationError, "Unexpected interpretation error")
	require.Contains(t, interpretationError.GetErrorMessage(), "Evaluation error: struct has no .greetings attribute")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Empty(t, test_helpers.GenerateScriptOutput(instructions))
	logrus.Info("Successfully ran Startosis module")
}
