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
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseMainStarMissingTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithNoMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "An error occurred while verifying that 'main.star' exists on root of module"
	_, err = enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.NotNil(t, err, "Unexpected error executing startosis module")
	require.Contains(t, err.Error(), expectedErrorContents)
}
