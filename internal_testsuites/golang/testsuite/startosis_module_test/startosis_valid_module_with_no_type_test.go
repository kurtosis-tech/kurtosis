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

func TestStartosisModule_ValidModuleWithNoType(t *testing.T) {
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

	outputStream, _, err := enclaveCtx.ExecuteKurtosisModule(ctx, moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Hello World!
`
	require.Nil(t, interpretationError, "Unexpected interpretation error")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, test_helpers.GenerateScriptOutput(instructions))
	logrus.Info("Successfully ran Startosis module")
}
