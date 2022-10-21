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
	testName               = "startosis_module_test"
	isPartitioningEnabled  = false
	sampleModuleRelDirpath = "../../static_files/sample-kurtosis-module"
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, sampleModuleRelDirpath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis Module...")

	logrus.Debugf("Startosis module path: \n%v", moduleDirpath)

	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath)
	require.NoError(t, err, "Unexpected error executing startosis module")

	expectedScriptOutput := `Hello World!
`
	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error")
	require.Lenf(t, executionResult.ValidationErrors, 0, "Unexpected validation error")
	require.Empty(t, executionResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Infof("Successfully ran Startosis module")
}
