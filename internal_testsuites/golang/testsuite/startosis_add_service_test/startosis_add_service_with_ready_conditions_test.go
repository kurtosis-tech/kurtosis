package startosis_add_service_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServiceWithReadyConditionsTestName = "service-test"

	addServiceWithReadyConditionsScript = `
def run(plan):
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=foo/bar",
	)

	ready_conditions = ReadyCondition(
        recipe=get_recipe,
		field="code",
		assertion="==",
		target_value=%v,
		interval="1s",
		timeout="3s"
    )

	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		},
        ready_conditions = ready_conditions
	)

	plan.add_service(name = "web-server", config = service_config)
`

	okStatusCode          = 200
	serverErrorStatusCode = 500
)

func TestStartosis_AddServiceWithReadyConditionsCheck(t *testing.T) {
	ctx := context.Background()

	script := fmt.Sprintf(addServiceWithReadyConditionsScript, okStatusCode)

	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, addServiceWithReadyConditionsTestName, script)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
}

func TestStartosis_AddServiceWithReadyConditionsCheckFail(t *testing.T) {
	ctx := context.Background()

	expectedLastAssertionErrorStr := fmt.Sprintf("Assertion failed '%v' '==' '%v'", okStatusCode, serverErrorStatusCode)

	script := fmt.Sprintf(addServiceWithReadyConditionsScript, serverErrorStatusCode)

	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, addServiceWithReadyConditionsTestName, script)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from failed ready conditions")
	require.Contains(t, runResult.ExecutionError.ErrorMessage, expectedLastAssertionErrorStr)
}
