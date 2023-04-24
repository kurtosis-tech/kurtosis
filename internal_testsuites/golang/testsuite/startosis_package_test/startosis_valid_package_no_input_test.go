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
	validPackageNoTypeTestName = "valid-package-no-input"
	validPackageNoTypeRelPath  = "../../../starlark/valid-kurtosis-package-no-input"
)

func TestStartosisPackage_ValidPackageNoInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validPackageNoTypeTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, validPackageNoTypeRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, emptyRunParams, defaultDryRun, defaultParallelism)
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutput := `Hello world!
{
	"message": "Hello world!"
}
`
	require.Equal(t, expectedScriptOutput, string(runResult.RunOutput))
	require.Len(t, runResult.Instructions, 1)
}

func TestStartosisPackage_ValidPackageNoInput_PassingParamsAlsoWorks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validPackageNoTypeTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, validPackageNoTypeRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	params := `{"greetings": "bonjour!"}`
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, params, defaultDryRun, defaultParallelism)
	require.Nil(t, err, "Unexpected error executing Starlark package")

	require.Nil(t, runResult.InterpretationError)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)

	expectedScriptOutput := `Hello world!
{
	"message": "Hello world!"
}
`

	require.Equal(t, expectedScriptOutput, string(runResult.RunOutput))
	require.Len(t, runResult.Instructions, 1)
}
