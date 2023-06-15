package startosis_request_wait_assert_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	assertSuccessScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP", wait="10s"),
		}
	)

	plan.add_service(name = "web-server", config = service_config)
`

	assertFailScript = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP"),
			"not-opened-port": PortSpec(number = 1234, wait="10s")
		}
	)

	plan.add_service(name = "web-server", config = service_config)
`

	assertFailScript2 = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP", wait=""),
		}
	)

	plan.add_service(name = "web-server", config = service_config)
`
)

func (suite *StartosisPortsWaitTestSuite) TestStartosis_AssertSuccessPortChecks() {
	ctx := context.Background()
	_, err := suite.RunScript(ctx, assertSuccessScript)

	t := suite.T()

	require.Nil(t, err)
}

func (suite *StartosisPortsWaitTestSuite) TestStartosis_AssertFailBecausePortIsNotOpen() {
	ctx := context.Background()
	runResult, _ := suite.RunScript(ctx, assertFailScript)

	t := suite.T()

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from assert fail")
}

func (suite *StartosisPortsWaitTestSuite) TestStartosis_AssertFailBecauseEmptyStringIsNotValid() {
	ctx := context.Background()
	runResult, _ := suite.RunScript(ctx, assertFailScript2)

	t := suite.T()

	require.NotNil(t, runResult.InterpretationError, "Expected interpretation error coming from wait validation")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
}
