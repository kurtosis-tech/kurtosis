package startosis_recipe_get_value_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "startosis_recipe_get_value_test"
	isPartitioningEnabled = false
	defaultDryRun         = false

	emptyParams     = "{}"
	startosisScript = `
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
		endpoint = "?input=output",
		method = "GET",
		extract = {
			"input": ".query.input"
		}
	)
	response = wait(get_recipe, "code", "==", 200, backoff="10s", retry=3)
	assert(response["code"], "==", 200)
	assert("My test returned " + response["code"], "==", "My test returned 200")
	assert(response["code"], "!=", 500)
	assert(response["code"], ">=", 200)
	assert(response["code"], "<=", 200)
	assert(response["code"], "<", 300)
	assert(response["code"], ">", 100)
	assert(response["code"], "IN", [100, 200])
	assert(response["code"], "NOT_IN", [100, 300])
	assert(response["input"], "==", "output")
	post_recipe = struct(
		service_id = "web-server",
		port_id = "http-port",
		endpoint = "/",
		method = "POST",
		content_type="text/plain",
		body="post_output",
		extract = {
			"my-body": ".body"
		}
	)
	wait(post_recipe, "code", "==", 200)
	post_response = request(post_recipe)
	assert(post_response["code"], "==", 200)
	assert(post_response["my-body"], "==", "post_output")
`
)

func TestStartosis(t *testing.T) {
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
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis script")

}
