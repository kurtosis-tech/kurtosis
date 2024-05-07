package startosis_request_wait_assert_test

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	complexRequestWaitAssertStartosisScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-complex-request-wait-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval="10s", timeout="200s", service_name="web-server-complex-request-wait-test")
	plan.verify(response["code"], "==", 200)
	plan.verify("My test returned " + response["code"], "==", "My test returned 200")
	plan.verify(response["code"], "!=", 500)
	plan.verify(response["code"], ">=", 200)
	plan.verify(response["code"], "<=", 200)
	plan.verify(response["code"], "<", 300)
	plan.verify(response["code"], ">", 100)
	plan.verify(response["code"], "IN", [100, 200])
	plan.verify(response["code"], "NOT_IN", [100, 300])
	plan.verify(response["extract.exploded-slash"], "==", "bar")
	post_recipe = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		content_type="text/plain",
		headers={"fizz":"buzz"},
		body=response["extract.exploded-slash"],
		extract = {
			"my-body": ".body",
			"my-content-type": '.headers["content-type"]',
			"my-headers-fizz": '.headers["fizz"]'
		}
	)
	plan.wait(recipe=post_recipe, field="code", assertion="==", target_value=200, service_name="web-server-complex-request-wait-test")
	post_response = plan.request(recipe=post_recipe, service_name="web-server-complex-request-wait-test")
	plan.verify(post_response["code"], "==", 200)
	plan.verify(post_response["extract.my-content-type"], "==", "text/plain")
	plan.verify(post_response["extract.my-headers-fizz"], "==", "buzz")
	plan.verify(post_response["extract.my-body"], "==", "bar")
	post_recipe_no_body = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		body = "0",
		extract = {
			"my-content-type": '.headers["content-type"]'
		}
	)
	post_recipe_no_body_output = plan.wait(recipe=post_recipe_no_body, field="code", assertion="==", target_value=200, service_name = "web-server-complex-request-wait-test")
	plan.verify(post_recipe_no_body_output["extract.my-content-type"], "==", "application/json")
	exec_recipe = ExecRecipe(
		command = ["echo", "hello", post_response["extract.my-body"]]
	)
	exec_result = plan.wait(recipe=exec_recipe, field="code", assertion="==", target_value=0, service_name="web-server-complex-request-wait-test")
	plan.verify(exec_result["output"], "==", "hello bar\n")

	# content_type default to application/json
	post_json = PostHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "/",
		body='{"a":"b"}',
		extract = {
			"my-json": ".body"
		}
	)
	post_json_response = plan.request(recipe = post_json, service_name = "web-server-complex-request-wait-test")
	plan.verify(post_json_response["extract.my-json"], "==", '{"a":"b"}')
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStartosis_ComplexRequestWaitAssert() {
	ctx := context.Background()
	t := suite.T()
	_, err := suite.RunScript(ctx, complexRequestWaitAssertStartosisScript)

	require.Nil(t, err)
	logrus.Infof("Successfully ran Startosis script")
}
