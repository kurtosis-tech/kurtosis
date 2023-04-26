package startosis_request_wait_assert_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	assertSuccessTestName = "startosis_ports_wait_sucess_test"
	assertSuccessScript   = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP", wait="2s"),
		}
	)

	plan.add_service(name = "web-server", config = service_config)
`

	assertFailTestName = "startosis_ports_wait_fail_test"
	assertFailScript   = `
def run(plan):
	service_config = ServiceConfig(
		image = "mendhak/http-https-echo:26",
		ports = {
			"http-port": PortSpec(number = 8080, transport_protocol = "TCP"),
			"not-opened-port": PortSpec(number = 1234, wait="2s")
		}
	)

	plan.add_service(name = "web-server", config = service_config)
`

	assertFailTestName2 = "startosis_ports_wait_fail_2_test"
	assertFailScript2   = `
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

func TestStartosis_AssertSuccessPortChecks(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, assertSuccessTestName, assertSuccessScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
}

func TestStartosis_AssertFailBecausePortIsNotOpen(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, assertFailTestName, assertFailScript)

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.NotEmpty(t, runResult.ExecutionError, "Expected execution error coming from assert fail")
}

func TestStartosis_AssertFailBecauseEmptyStringIsNotValid(t *testing.T) {
	ctx := context.Background()
	runResult := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, assertFailTestName2, assertFailScript2)

	require.NotNil(t, runResult.InterpretationError, "Expected interpretation error coming from wait validation")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Empty(t, runResult.ExecutionError, "Unexpected execution error")
}
