package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	requestInvalidPortIDFailScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server-invalid-port-id-request-test", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "invalid-port-id",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.request(service_name = "web-server-invalid-port-id-request-test", recipe = get_recipe)
`
)

func (suite *StartosisRequestWaitAssertTestSuite) TestStarlark_InvalidPortIdRequest() {
	ctx := context.Background()
	runResult, _ := suite.RunScript(ctx, requestInvalidPortIDFailScript)

	t := suite.T()
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation errors")
	require.Len(t, runResult.ValidationErrors, 1)
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Request required port ID 'invalid-port-id' to exist on service 'web-server-invalid-port-id-request-test' but it doesn't")
}
