package startosis_request_wait_assert

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	waitInvalidPortIDTest       = "starlark-wait-invalid-portid"
	waitInvalidPortIDFailScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		}
	)

	plan.add_service(name = "web-server", config = service_config)
	get_recipe = GetHttpRequestRecipe(
		port_id = "invalid-port-id",
		endpoint = "?input=foo/bar",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = plan.wait(service_name = "web-server", recipe = get_recipe, field = "code", assertion = "==", target_value = 200)
`
)

func TestStarlark_InvalidPortIdWait(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, waitInvalidPortIDTest, waitInvalidPortIDFailScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation error")
	require.Len(t, runResult.ValidationErrors, 1)
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Request required port ID 'invalid-port-id' to exist on service 'web-server' but it doesn't")
}
