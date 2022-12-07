package starlark_exec_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "starlark_exec_test"
	isPartitioningEnabled = false
	defaultDryRun         = false

	emptyParams    = "{}"
	starlarkScript = `
def run(args):
	service_config = struct(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": struct(number = 8080, protocol = "TCP")
		}
	)

	add_service(service_id = "web-server", config = service_config)
	response = exec("web-server", ["echo", "hello", "world"])
	assert(response.code, "==", 0)
	assert(response.output, "==", "hello world\n")
`
)

func TestStarlarkExec(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Startosis script content: \n%v", starlarkScript)

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, emptyParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Starlark script")

}
