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

	startosisScript = `
service_config = struct(
    image = "httpd:latest",
    ports = {
        "http_port": struct(number = 80, protocol = "TCP")
    }
)
add_service(service_id = "web-server", config = service_config)
recipe = struct(
    service_id = "web-server",
    port_id = "http_port",
    endpoint = "/",
    method = "GET",
)
response = get_value(recipe)
assert(response.code, "==", 200)
assert(response.code, "!=", 500)
assert(response.code, ">=", 200)
assert(response.code, "<=", 200)
assert(response.code, "<", 300)
assert(response.code, ">", 100)
assert(response.code, "IN", [100, 200])
assert(response.code, "NOT_IN", [100, 300])
assert(response.body, "==", "<html><body><h1>It works!</h1></body></html>\n")
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

	executionResult, err := enclaveCtx.ExecuteStartosisScript(startosisScript, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis script")

	require.Nil(t, executionResult.GetInterpretationError(), "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, executionResult.GetValidationErrors().GetErrors(), "Unexpected validation error")
	require.Nil(t, executionResult.GetExecutionError(), "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis script")

}
