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
	isPartitioningEnabled = false
	emptyExecuteParams    = "{}"
	defaultDryRun         = false

	validModuleWithTypesTestName = "valid-module-with-types"
	validModuleWithTypesRelPath  = "../../../startosis/valid-kurtosis-module-with-types"

	validModuleNoTypeTestName = "valid-module-no-type"
	validModuleNoTypeRelPath  = "../../../startosis/valid-kurtosis-module-no-type"

	validModuleNoModuleInputTypeTestName = "valid-module-no-input-type"
	validModuleNoModuleInputTypeRelPath  = "../../../startosis/valid-kurtosis-module-no-module-input-type"

	invalidTypesFileTestName = "invalid-types-file"
	invalidTypesFileRelPath  = "../../../startosis/invalid-types-file"

	invalidModuleNoTypeButInputArgsTestName = "invalid-module-no-type-input-args"
	invalidModuleNoTypeButInputArgsRelPath  = "../../../startosis/invalid-no-type-but-input-args"

	invalidCaseModFileTestName          = "invalid-module-invalid-mod-file"
	moduleWithInvalidKurtosisModRelPath = "../../../startosis/invalid-mod-file"

	invalidCaseMainStarMissingTestName = "invalid-module-no-main-file"
	moduleWithNoMainStarRelPath        = "../../../startosis/no-main-star"

	invalidCaseNoMainInMainStarTestName = "invalid-module-missing-main"
	moduleWithNoMainInMainStarRelPath   = "../../../startosis/no-main-in-main-star"
)

func TestStartosisModule_ValidModuleWithType(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleWithTypesTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleWithTypesRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	serializedParams := `{"greetings": "Bonjour!"}`
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, serializedParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")

	expectedScriptOutput := `Bonjour!
Hello World!
`
	require.Nil(t, executionResult.GetInterpretationError(), "Unexpected interpretation error")
	require.Nil(t, executionResult.GetValidationErrors(), "Unexpected validation error")
	require.Nil(t, executionResult.GetExecutionError(), "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Info("Successfully ran Startosis module")
}

func TestStartosisModule_ValidModuleWithNoType(t *testing.T) {
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

	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")

	expectedScriptOutput := `Hello World!
`
	require.Nil(t, executionResult.GetInterpretationError(), "Unexpected interpretation error")
	require.Nil(t, executionResult.GetValidationErrors(), 0, "Unexpected validation error")
	require.Nil(t, executionResult.GetExecutionError(), "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Info("Successfully ran Startosis module")
}

func TestStartosisModule_ValidModuleWithNoTypesFile_FailureCalledWithParams(t *testing.T) {
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
	serializedParams := `{"greetings": "Bonjour!"}`
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, serializedParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	expectedInterpretationErr := "A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid)."
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
}

func TestStartosisModule_ValidModuleNoModulInputTypeTestName(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleNoModuleInputTypeTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleNoModuleInputTypeRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis module")

	expectedScriptOutput := `Hello world!
`
	require.Nil(t, executionResult.GetInterpretationError(), "Unexpected interpretation error")
	require.Nil(t, executionResult.GetValidationErrors(), "Unexpected validation error")
	require.Nil(t, executionResult.GetExecutionError(), "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Info("Successfully ran Startosis module")
}

func TestStartosisModule_ValidModuleNoModulInputTypeTestName_FailureCalledWithParams(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, validModuleNoModuleInputTypeTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, validModuleNoModuleInputTypeRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	serializedParams := `{"greetings": "Bonjour!"}`
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, serializedParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	expectedInterpretationErr := "A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but 'ModuleInput' type is not defined in the module's 'types.proto' file."
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
}

func TestStartosisModule_InvalidTypesFileTestName(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidTypesFileTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, invalidTypesFileRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	serializedParams := `{"greetings": "Bonjour!"}`
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, serializedParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	expectedInterpretationErr := "A non empty parameter was passed to the module 'github.com/sample/sample-kurtosis-module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid)."
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
}

func TestStartosisModule_InvalidModuleNoTypesButInputArgsTestName(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, invalidModuleNoTypeButInputArgsTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	currentWorkingDirectory, err := os.Getwd()
	require.Nil(t, err)
	moduleDirpath := path.Join(currentWorkingDirectory, invalidModuleNoTypeButInputArgsRelPath)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	expectedInterpretationErr := "Evaluation error: function main missing 1 argument (input_args)"
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
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
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "Field module.name in kurtosis.mod needs to be set and cannot be empty"
	_, err = enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
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
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedErrorContents := "An error occurred while verifying that 'main.star' exists on root of module"
	_, err = enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
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
	logrus.Info("Executing Startosis Module...")

	logrus.Infof("Startosis module path: \n%v", moduleDirpath)

	expectedInterpretationErr := "Evaluation error: module has no .main field or method\n\tat [3:12]: <toplevel>"
	executionResult, err := enclaveCtx.ExecuteStartosisModule(moduleDirpath, emptyExecuteParams, defaultDryRun)
	require.Nil(t, err, "Unexpected error executing startosis module")
	require.NotNil(t, executionResult.GetInterpretationError())
	require.Contains(t, executionResult.GetInterpretationError().GetErrorMessage(), expectedInterpretationErr)
	require.Nil(t, executionResult.GetValidationErrors())
	require.Nil(t, executionResult.GetExecutionError())
	require.Empty(t, executionResult.SerializedScriptOutput)
}
