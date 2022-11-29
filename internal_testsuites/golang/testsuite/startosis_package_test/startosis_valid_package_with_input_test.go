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
	validPackageWithInputTestName = "valid-module-with-input"
	validPackageWithInputRelPath  = "../../../starlark/valid-kurtosis-package-with-input"
)

func TestStartosisPackage_ValidPackageWithInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validPackageWithInputTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, validPackageWithInputRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Package...")

	logrus.Infof("Startosis package path: \n%v", packageDirpath)

	params := `{"greetings": "bonjour!"}`
	outputStream, _, err := enclaveCtx.RunStarlarkPackage(ctx, packageDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing starlark package")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `bonjour!
Hello World!
`
	require.Nil(t, interpretationError, "Unexpected interpretation error")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, scriptOutput)
	logrus.Info("Successfully ran Startosis module")
}

func TestStartosisPackage_ValidPackageWithInput_MissingKeyInParams(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validPackageWithInputTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validPackageWithInputRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Package...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	params := `{"hello": "world"}` // expecting key 'greetings' here
	outputStream, _, err := enclaveCtx.RunStarlarkPackage(ctx, moduleDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	require.NotNil(t, interpretationError, "Unexpected interpretation error")
	require.Contains(t, interpretationError.GetErrorMessage(), "Evaluation error: struct has no .greetings attribute")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Empty(t, scriptOutput)
	logrus.Info("Successfully ran Startosis module")
}
