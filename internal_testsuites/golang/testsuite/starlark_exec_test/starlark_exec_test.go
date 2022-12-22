package starlark_exec_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	testName              = "starlark_exec_test"
	isPartitioningEnabled = false
	defaultDryRun         = false
	emptyParams           = "{}"
)

const (
	noExecOutput              = ""
	expectedLogOutput         = "hello\\n"
	expectedAdvancedLogOutput = "hello && hello\\n"

	successExitCode int32 = 0
	failureExitCode int32 = 1

	setupStarlarkScript = `
def run(plan, args):
	service_config = struct(
		image = "alpine:3.12.4",
		entrypoint = ["sleep"],
		cmd = ["30"]
	)
	plan.add_service(service_id = "test", config = service_config)
`
	testStarlarkScriptTemplate = `
def run(plan, args):
	exec_recipe = struct(
		service_id = "test",
		command = %v,
	)
	exec_result = plan.exec(exec_recipe)
	plan.assert(exec_result["code"], "==", %d)
	plan.assert(exec_result["output"], "==", "%s")
`
)

var (
	execCommandThatShouldWork          = []string{"true"}
	execCommandThatShouldFail          = []string{"false"}
	execCommandThatShouldHaveLogOutput = []string{"echo", "hello"}

	// This command tests to ensure that the commands the user is running get passed exactly as-is to the Docker
	// container. If Kurtosis code is magically wrapping the code with "sh -c", this will fail.
	execCommandThatWillFailIfShWrapped = []string{"echo", "hello && hello"}

	testStarlarkScripts = []string{
		fmt.Sprintf(testStarlarkScriptTemplate, sliceToStarlarkString(execCommandThatShouldWork), successExitCode, noExecOutput),
		fmt.Sprintf(testStarlarkScriptTemplate, sliceToStarlarkString(execCommandThatShouldFail), failureExitCode, noExecOutput),
		fmt.Sprintf(testStarlarkScriptTemplate, sliceToStarlarkString(execCommandThatShouldHaveLogOutput), successExitCode, expectedLogOutput),
		fmt.Sprintf(testStarlarkScriptTemplate, sliceToStarlarkString(execCommandThatWillFailIfShWrapped), successExitCode, expectedAdvancedLogOutput),
	}
)

func TestStarlarkExec(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// -------------------------------------- TEST SETUP -----------------------------------------------
	logrus.Infof("Executing Starlark setup script...")

	setupRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, setupStarlarkScript, emptyParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing Starlark script")
	require.Nil(t, setupRunResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, setupRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, setupRunResult.ExecutionError, "Unexpected execution error")

	// ------------------------------------- TEST RUN ----------------------------------------------

	for testIndex, testStarlarkScript := range testStarlarkScripts {
		logrus.Infof("Executing Starlark test script %d:\n%s", testIndex, testStarlarkScript)
		runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, testStarlarkScript, emptyParams, defaultDryRun)
		require.NoError(t, err, "Unexpected error executing Starlark script")
		require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
		require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
		require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
		logrus.Infof("Successfully ran Starlark test script")
	}

}

func sliceToStarlarkString(slice []string) string {
	quotedSlice := []string{}
	for _, s := range slice {
		quotedSlice = append(quotedSlice, fmt.Sprintf(`"%s"`, s))
	}
	return fmt.Sprintf("[%v]", strings.Join(quotedSlice, ","))
}
