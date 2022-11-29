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
	validModuleNoTypeTestName = "valid-module-no-input"
	validModuleNoTypeRelPath  = "../../../startosis/valid-kurtosis-module-no-input"
)

func TestStartosisModule_ValidModuleNoInput(t *testing.T) {
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

	outputStream, _, err := enclaveCtx.ExecuteStarlarkPackage(ctx, moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Hello world!
`

	require.Nil(t, interpretationError)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Equal(t, expectedScriptOutput, scriptOutput)
}

func TestStartosisModule_ValidModuleNoInput_PassingParamsAlsoWorks(t *testing.T) {
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

	params := `{"greetings": "bonjour!"}`
	outputStream, _, err := enclaveCtx.ExecuteStarlarkPackage(ctx, moduleDirpath, params, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Hello world!
`

	require.Nil(t, interpretationError)
	require.Empty(t, validationErrors)
	require.Nil(t, executionError)
	require.Equal(t, expectedScriptOutput, scriptOutput)
}
