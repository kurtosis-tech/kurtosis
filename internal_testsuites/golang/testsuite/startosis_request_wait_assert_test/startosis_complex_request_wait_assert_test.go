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
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval="10s", timeout="200s", service_name="web-server")
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
	post_recipe = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		content_type="text/plain",
		body=response["extract.exploded-slash"],
		extract = {
			"my-body": ".body",
			"my-content-type": '.headers["content-type"]'
		}
	)
	plan.wait(recipe=post_recipe, field="code", assertion="==", target_value=200, service_name="web-server")
	post_response = plan.request(recipe=post_recipe, service_name="web-server")
	plan.assert(post_response["code"], "==", 200)
	plan.assert(post_response["extract.my-content-type"], "==", "text/plain")
	plan.assert(post_response["extract.my-body"], "==", "bar")
	post_recipe_no_body = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		body = "0",
		extract = {
			"my-content-type": '.headers["content-type"]'
		}
	)
	post_recipe_no_body_output = plan.wait(recipe=post_recipe_no_body, field="code", assertion="==", target_value=200, service_name = "web-server")
	plan.assert(post_recipe_no_body_output["extract.my-content-type"], "==", "application/json")
	exec_recipe = ExecRecipe(
		command = ["echo", "hello", post_response["extract.my-body"]]
	)
	exec_result = plan.wait(recipe=exec_recipe, field="code", assertion="==", target_value=0, service_name="web-server")
	plan.assert(exec_result["output"], "==", "hello bar\n")

	# content_type default to application/json
	post_json = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		body='{"a":"b"}',
		extract = {
			"my-json": ".body"
		}
	)
	post_json_response = plan.request(recipe = post_json, service_name = "web-server")
	plan.assert(post_json_response["extract.my-json"], "==", '{"a":"b"}')
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
