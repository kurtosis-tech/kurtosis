package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	waitInvalidServiceTestScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-invalid-service-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(service_name = "invalid-service", recipe = get_recipe, field = "code", assertion = "==", target_value = 200)
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStarlark_InvalidServiceWait() {
	ctx := context.Background()
	t := suite.T()
	runResult, _ := suite.RunScript(ctx, waitInvalidServiceTestScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation error")
	require.Len(t, runResult.ValidationErrors, 1)
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Tried creating a wait for service 'invalid-service' which doesn't exist")
}
