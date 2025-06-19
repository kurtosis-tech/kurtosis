package startosis_add_service_test

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/require"
)

const (
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
		timeout="40s"
    )

	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP")
		},
        ready_conditions = ready_conditions
	)

	plan.add_service(name = "ws-ready-conditions-%v", config = service_config)
`

	okStatusCode          = 200
	serverErrorStatusCode = 500
)

func (suite *StartosisAddServiceTestSuite) TestStartosis_AddServiceWithReadyConditionsCheck() {
	ctx := context.Background()
	script := fmt.Sprintf(addServiceWithReadyConditionsScript, okStatusCode, okStatusCode)
	_, err := suite.RunScript(ctx, script)

	t := suite.T()

	require.Nil(t, err)
}

func (suite *StartosisAddServiceTestSuite) TestStartosis_AddServiceWithReadyConditionsCheckFail() {
	ctx := context.Background()
	script := fmt.Sprintf(addServiceWithReadyConditionsScript, serverErrorStatusCode, serverErrorStatusCode)
	runResult, _ := suite.RunScript(ctx, script)

	t := suite.T()
	expectedLastAssertionErrorStr := fmt.Sprintf("Verification failed '%v' '==' '%v'", okStatusCode, serverErrorStatusCode)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from failed ready conditions")
	require.Contains(t, runResult.ExecutionError.ErrorMessage, expectedLastAssertionErrorStr)
}
