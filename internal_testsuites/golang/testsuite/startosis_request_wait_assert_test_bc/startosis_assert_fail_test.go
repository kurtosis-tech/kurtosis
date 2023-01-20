package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const assertFailScriptWithStruct = `
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
	response = plan.wait(get_recipe, "code", "==", 200, interval="100ms", timeout="30s")
	plan.assert(response["code"], "!=", 200)

	# dumb test to validate we can pass 2 runtime values here
	plan.assert(response["code"], "==", response["code"])
`

func TestStartosis_AssertFailWithStruct(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, "starlarkUnderTest", assertFailScriptWithStruct)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from assert fail")
}
