package test_engine

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

var waitRecipeTestCaseServicePositional *service.Service = service.NewService(
	service.NewServiceRegistration(
		waitRecipeTestCaseServiceName,
		service.ServiceUUID(""),
		enclave.EnclaveUUID(""),
		net.IP{},
		"",
	),
	map[string]*port_spec.PortSpec{},
	net.IP{},
	map[string]*port_spec.PortSpec{},
	container.NewContainer(
		container.ContainerStatus_Running,
		"",
		[]string{},
		[]string{},
		map[string]string{},
	),
)

// This test case is for testing positional arguments retro-compatibility for those script
// that are using the recipe value as the first positional argument
type waitWithPositionalArgsTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestWaitWithPositionalArgs() {
	suite.serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(waitRecipeTestCaseServiceName),
	).Times(1).Return(
		waitRecipeTestCaseServicePositional,
		nil,
	)

	suite.serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		waitRecipeTestCaseServicePositional,
		waitRecipePortId,
		waitRecipeMethod,
		waitRecipeContentType,
		waitRecipeEndpoint,
		waitRecipeBody,
		testEmptyHeaders,
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(waitRecipeResponseBody)),
			ContentLength:    -1,
			TransferEncoding: nil,
			Close:            false,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          nil,
			TLS:              nil,
		},
		nil,
	)

	suite.run(&waitWithPositionalArgsTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *waitWithPositionalArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return wait.NewWait(t.serviceNetwork, t.runtimeValueStore)
}

func (t *waitWithPositionalArgsTestCase) GetStarlarkCode() string {
	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
	return fmt.Sprintf("%s(%q, %s, %q, %q, %s, %q, %q)", wait.WaitBuiltinName, waitRecipeTestCaseServiceName, recipeStr, waitValueField, waitAssertion, waitTargetValue, waitInterval, waitTimeout)
}

func (t *waitWithPositionalArgsTestCase) GetStarlarkCodeForAssertion() string {
	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
	return fmt.Sprintf("%s(%s=%q, %s=%s, %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)", wait.WaitBuiltinName, wait.ServiceNameArgName, waitRecipeTestCaseServiceName, wait.RecipeArgName, recipeStr, wait.ValueFieldArgName, waitValueField, wait.AssertionArgName, waitAssertion, wait.TargetArgName, waitTargetValue, wait.IntervalArgName, waitInterval, wait.TimeoutArgName, waitTimeout)
}

func (t *waitWithPositionalArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResult := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResult, interpretationResult.String())

	expectedExecutionResult := `Assertion passed with following:
Request had response code '200' and body "{\"value\": \"Hello world!\"}", with extracted fields:
'extract.key': "Hello world!"`

	require.Contains(t, *executionResult, expectedExecutionResult)
}
