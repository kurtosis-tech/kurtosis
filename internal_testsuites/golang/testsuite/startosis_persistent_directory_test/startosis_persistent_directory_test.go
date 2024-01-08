package startosis_persistent_directory_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "persist-data-test"

	addServiceWriteLineToFile = `
IMAGE = "docker/getting-started"
SERVICE_NAME = "test-service"

def run(plan):
	service_config = ServiceConfig(
		image=IMAGE,
		files={
			"/data": Directory(
				persistent_key="persistent-data",
			),
		},
		min_cpu=%d,
	)
	service = plan.add_service(name=SERVICE_NAME, config=service_config)

	plan.exec(
        service_name=SERVICE_NAME,
        recipe=ExecRecipe([
            "/bin/sh",
            "-c",
            "echo 'Hello world !' >> /data/test.log",
        ])
    )
	
	plan.exec(
        service_name=SERVICE_NAME,
        recipe=ExecRecipe([
            "/bin/sh",
            "-c",
            "wc -l /data/test.log",
        ])
    )
`
)

func TestAddServiceAndPersistentFileToDirectory(t *testing.T) {

	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	firstScript := fmt.Sprintf(addServiceWriteLineToFile, 100)
	firstRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, firstScript)
	logrus.Infof("Test Output: %v", firstRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, firstRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, firstRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, firstRunResult.ExecutionError, "Unexpected execution error")
	// This checks the output of the `wc -l` at the end. For the first run, the file should contain one line only
	require.Contains(t, firstRunResult.RunOutput, "\n1 /data/test.log\n")

	secondScript := fmt.Sprintf(addServiceWriteLineToFile, 150) // we slightly change the service config so that the service gets restarted and the execs re-run
	secondRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, secondScript)
	logrus.Infof("Test Output: %v", secondRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, secondRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, secondRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, secondRunResult.ExecutionError, "Unexpected execution error")
	// For the second run, it should contain 2 lines as the file persisted between the 2 runs
	require.Contains(t, secondRunResult.RunOutput, "\n2 /data/test.log\n")
}
