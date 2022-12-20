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
def run(plan):
	service_config = struct(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)
	plan.add_service(service_id = "web-server", config = service_config)
	exec_recipe = struct(
		service_id = "web-server",
		command = ["echo", "hello", "world"]
	)
	response = plan.exec(exec_recipe)
	plan.assert(response["code"], "==", 0)
	plan.assert(response["output"], "==", "hello world\n")
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

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, emptyParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Starlark script")

}
