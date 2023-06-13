package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	waitInvalidPortIDFailScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-invalid-port-id-wait-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "invalid-port-id",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(service_name = "web-server-invalid-port-id-wait-test", recipe = get_recipe, field = "code", assertion = "==", target_value = 200)
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStarlark_InvalidPortIdWait() {
	ctx := context.Background()
	t := suite.T()
	runResult, _ := suite.RunScript(ctx, waitInvalidPortIDFailScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation error")
	require.Len(t, runResult.ValidationErrors, 1)
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Request required port ID 'invalid-port-id' to exist on service 'web-server-invalid-port-id-wait-test' but it doesn't")
}
