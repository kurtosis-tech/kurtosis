package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	timeoutWaitTestName        = "startosis_timeout_wait_test"
	timeoutWaitStartosisScript = `
def run(args):
	service_config = struct(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, protocol = "TCP")
		}
	)

	add_service(service_id = "web-server", config = service_config)
	get_recipe = struct(
		service_id = "web-server",
		port_id = "http-port",
		endpoint = "?input=foo/bar",
		method = "GET",
		extract = {
			"exploded-slash": ".query.input | split(\"/\") | .[1]"
		}
	)
	response = wait(get_recipe, "code", "<", 0, interval="100ms", timeout="10s")
`
)

func TestStartosis_TimeoutWait(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, timeoutWaitTestName, timeoutWaitStartosisScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from wait timeout")
}
