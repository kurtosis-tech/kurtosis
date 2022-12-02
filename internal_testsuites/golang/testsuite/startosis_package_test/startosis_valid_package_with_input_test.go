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
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing starlark package")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `bonjour!
Hello World!
`
	require.Equal(t, expectedScriptOutput, runResult.RunOutput)
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
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, moduleDirpath, params, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")

	require.NotNil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Contains(t, runResult.InterpretationError.GetErrorMessage(), "Evaluation error: struct has no .greetings attribute")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Empty(t, runResult.RunOutput)
	logrus.Info("Successfully ran Startosis module")
}
