package startosis_add_service_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServicesWithReadyConditionsTestName = "services-test"

	addServicesWithReadyConditionsScript = `
HTTP_ECHO_IMAGE = "mendhak/http-https-echo:26"
SERVICE_NAME_PREFIX = "service-"
NUM_SERVICES = 4

def run(plan):
    plan.print("Adding {0} services to enclave".format(NUM_SERVICES))

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

    config = ServiceConfig(
        image = HTTP_ECHO_IMAGE,
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		},
		ready_conditions = ready_conditions,
    )
    configs = {}
    for index in range(NUM_SERVICES):
        service_name = SERVICE_NAME_PREFIX + str(index)
        configs[service_name] = config
    
    plan.add_services(configs)
`
)

func TestStartosis_AddServicesWithReadyConditionsCheck(t *testing.T) {
	ctx := context.Background()

	script := fmt.Sprintf(addServicesWithReadyConditionsScript, okStatusCode)

	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, addServicesWithReadyConditionsTestName, script)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
}

func TestStartosis_AddServicesWithReadyConditionsCheckFail(t *testing.T) {
	ctx := context.Background()

	expectedLastAssertionErrorStr := fmt.Sprintf("Assertion failed '%v' '==' '%v'", okStatusCode, serverErrorStatusCode)

	script := fmt.Sprintf(addServicesWithReadyConditionsScript, serverErrorStatusCode)

	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, addServicesWithReadyConditionsTestName, script)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from failed ready conditions")
	require.Contains(t, runResult.ExecutionError.ErrorMessage, expectedLastAssertionErrorStr)
}
