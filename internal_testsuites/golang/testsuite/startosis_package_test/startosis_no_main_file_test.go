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
	invalidCaseMainStarMissingTestName = "invalid-package-no-main-file"
	packageWithNoMainStarRelPath       = "../../../starlark/no-main-star"
)

func TestStartosisPackage_NoMainFile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, invalidCaseMainStarMissingTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.Nil(t, err, "Unexpected Error Occurred")
	}()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	packageDirpath := path.Join(currentWorkingDirectory, packageWithNoMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Starlark Package...")

	logrus.Infof("Starlark package path: \n%v", packageDirpath)

	expectedErrorContents := `An error occurred while verifying that 'main.star' exists in the package 'github.com/sample/sample-kurtosis-package' at '/kurtosis-data/startosis-packages/sample/sample-kurtosis-package/main.star'
	Caused by: stat /kurtosis-data/startosis-packages/sample/sample-kurtosis-package/main.star: no such file or directory`
	runResult, err := enclaveCtx.RunStarlarkPackageBlocking(ctx, packageDirpath, emptyRunParams, defaultDryRun, defaultParallelism)
	require.Nil(t, err, "Unexpected error executing package")
	require.NotNil(t, runResult.InterpretationError)
	require.Equal(t, runResult.InterpretationError.GetErrorMessage(), expectedErrorContents)
	require.Empty(t, runResult.ValidationErrors)
	require.Nil(t, runResult.ExecutionError)
	require.Empty(t, string(runResult.RunOutput))
}
