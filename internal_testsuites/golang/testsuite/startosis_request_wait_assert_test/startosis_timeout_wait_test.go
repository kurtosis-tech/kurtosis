package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	timeoutWaitStartosisScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-timeout-wait-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(recipe=get_recipe, field="code",  assertion="<", target_value=0, interval="100ms", timeout="10s", service_name="web-server-timeout-wait-test")
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStartosis_TimeoutWait() {
	ctx := context.Background()
	t := suite.T()
	runResult, _ := suite.RunScript(ctx, timeoutWaitStartosisScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from wait timeout")
}
