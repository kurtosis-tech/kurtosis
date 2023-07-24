package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	assertFailScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-assert-fail-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval="100ms", timeout="30s", service_name="web-server-assert-fail-test")
	plan.assert(response["code"], "!=", 200)

	# dumb test to validate we can pass 2 runtime values here
	plan.assert(response["code"], "==", response["code"])
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStartosis_AssertFail() {
	ctx := context.Background()
	t := suite.T()
	runResult, _ := suite.RunScript(ctx, assertFailScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from assert fail")
}
