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
	isPartitioningEnabled               = false
	validCaseTestName                   = "startosis-module-valid"
	validKurtosisModuleRelPath          = "../../../startosis/valid-kurtosis-module"
	invalidCaseModFileTestName          = "startosis-module-invalid-mod-file"
	moduleWithInvalidKurtosisModRelPath = "../../../startosis/invalid-mod-file"
	invalidCaseMainStarMissingTestName  = "startosis-module-missing-main-star"
	moduleWithNoMainStarRelPath         = "../../../startosis/no-main-star"
	invalidCaseNoMainInMainStarTestName = "startosis-module-missing-main"
	moduleWithNoMainInMainStarRelPath   = "../../../startosis/no-main-in-main-star"
)

func TestStartosisModule_SimpleValidCase(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validCaseTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validKurtosisModuleRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

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

func TestStartosisModule_InvalidModFile(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseModFileTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithInvalidKurtosisModRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "Field module.name in kurtosis.mod needs to be set and cannot be empty"
	_, err = enclaveCtx.ExecuteStartosisModule(moduleDirpath)
	require.NotNil(t, err, "Unexpected error executing startosis module")
	require.Contains(t, err.Error(), expectedErrorContents)
}

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
	logrus.Infof("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "An error occurred while verifying that 'main.star' exists on root of module"
	_, err = enclaveCtx.ExecuteStartosisModule(moduleDirpath)
	require.NotNil(t, err, "Unexpected error executing startosis module")
	require.Contains(t, err.Error(), expectedErrorContents)
}

func TestStartosisModule_NoMainInMainStar(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidCaseNoMainInMainStarTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, moduleWithNoMainInMainStarRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedInterpretationErr := "Evaluation error: load: name main not found in module github.com/sample/sample-kurtosis-module/main.star"
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.InterpretationError)
	require.Contains(t, executionResult.InterpretationError, expectedInterpretationErr)
	require.Nil(t, executionResult.ValidationErrors)
	require.Empty(t, executionResult.ExecutionError)
	require.Empty(t, executionResult.SerializedScriptOutput)
}
