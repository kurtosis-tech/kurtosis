package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	complexRequestWaitAssertTestName        = "startosis_complex_request_test"
	complexRequestWaitAssertStartosisScript = `
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
	response = wait(get_recipe, "code", "==", 200, interval="10s", timeout="200s")
	assert(response["code"], "==", 200)
	assert("My test returned " + response["code"], "==", "My test returned 200")
	assert(response["code"], "!=", 500)
	assert(response["code"], ">=", 200)
	assert(response["code"], "<=", 200)
	assert(response["code"], "<", 300)
	assert(response["code"], ">", 100)
	assert(response["code"], "IN", [100, 200])
	assert(response["code"], "NOT_IN", [100, 300])
	assert(response["extract.exploded-slash"], "==", "bar")
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
	assert(post_response["extract.my-body"], "==", "post_output")
`
)

func TestStartosis_ComplexRequestWaitAssert(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, complexRequestWaitAssertTestName, complexRequestWaitAssertStartosisScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis script")
}
