package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	requestInvalidServiceName       = "starlark-request-invalid-service"
	requestInvalidServiceNameScript = `
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
	response = plan.request(service_name = "invalid-service", recipe = get_recipe)
`
)

func TestStarlark_InvalidServiceRequest(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, requestInvalidServiceName, requestInvalidServiceNameScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation error")
	require.Len(t, runResult.ValidationErrors, 1)
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Tried creating a request for service 'invalid-service' which doesn't exist")
}
