package startosis_recipe_get_value_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	timeoutTestName        = "startosis_timeout_wait_test"
	timeoutStartosisScript = `
def run(args):
	service_config = struct(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": struct(number = 8080, protocol = "TCP")
		}
	)

	add_service(service_id = "web-server", config = service_config)
	get_recipe = struct(
		service_id = "web-server",
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		method = "GET",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = wait(get_recipe, "code", "==", 200, interval="100ms", timeout="30s")
`
)

func TestStartosis_TimeoutWait(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", startosisScript)

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, startosisScript, emptyParams, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis script")

}
