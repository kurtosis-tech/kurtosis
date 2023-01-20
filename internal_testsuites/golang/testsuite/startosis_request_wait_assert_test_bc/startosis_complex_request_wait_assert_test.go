package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

// TODO: remove everything below this once we stop supporting this functionality
const complexRequestWaitAssertStartosisScriptWithStruct = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(service_name = "web-server", config = service_config)
	get_recipe = struct(
		service_name = "web-server",
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		method = "GET",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(get_recipe, "code", "==", 200, interval="10s", timeout="200s")
	plan.assert(response["code"], "==", 200)
	plan.assert("My test returned " + response["code"], "==", "My test returned 200")
	plan.assert(response["code"], "!=", 500)
	plan.assert(response["code"], ">=", 200)
	plan.assert(response["code"], "<=", 200)
	plan.assert(response["code"], "<", 300)
	plan.assert(response["code"], ">", 100)
	plan.assert(response["code"], "IN", [100, 200])
	plan.assert(response["code"], "NOT_IN", [100, 300])
	plan.assert(response["extract.exploded-slash"], "==", "bar")
	post_recipe = struct(
		service_name = "web-server",
		port_id = "http-port",
		endpoint = "/",
		method = "POST",
		content_type="text/plain",
		body=response["extract.exploded-slash"],
		extract = {
			"my-body": ".body"
		}
	)
	plan.wait(post_recipe, "code", "==", 200)
	post_response = plan.request(post_recipe)
	plan.assert(post_response["code"], "==", 200)
	plan.assert(post_response["extract.my-body"], "==", "bar")
	exec_recipe = struct(
		service_name = "web-server",
		command = ["echo", "hello", post_response["extract.my-body"]]
	)
	exec_result = plan.wait(exec_recipe, "code", "==", 0)
	plan.assert(exec_result["output"], "==", "hello bar\n")
`

func TestStartosis_ComplexRequestWaitAssertWithStruct(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, "starlarkUnderTest", complexRequestWaitAssertStartosisScriptWithStruct)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	logrus.Infof("Successfully ran Startosis script")
}
