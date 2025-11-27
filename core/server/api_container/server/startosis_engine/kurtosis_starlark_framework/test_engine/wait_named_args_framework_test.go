package test_engine

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

const (
	waitRecipeTestCaseServiceName = service.ServiceName("web-server")

	waitAssertion   = "=="
	waitInterval    = "1s"
	waitTargetValue = "200"
	waitTimeout     = "5s"
	waitValueField  = "code"

	waitRecipePortId       = "http-port"
	waitRecipeMethod       = "POST"
	waitRecipeEndpoint     = "/"
	waitRecipeBody         = "{}"
	waitRecipeContentType  = "application/json"
	waitRecipeResponseBody = `{"value": "Hello world!"}`
)

type waitWithNamedArgsTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

var waitRecipeTestCaseService *service.Service = getService(waitRecipeTestCaseServiceName)

func (suite *KurtosisPlanInstructionTestSuite) TestWaitWithNamedArgs() {
	suite.serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(waitRecipeTestCaseServiceName),
	).Times(1).Return(
		waitRecipeTestCaseService,
		nil,
	)

	suite.serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		waitRecipeTestCaseService,
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

	suite.run(&waitWithNamedArgsTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *waitWithNamedArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return wait.NewWait(t.serviceNetwork, t.runtimeValueStore)
}

func (t *waitWithNamedArgsTestCase) GetStarlarkCode() string {
	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
	return fmt.Sprintf("%s(%s=%q, %s=%s, %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)", wait.WaitBuiltinName, wait.ServiceNameArgName, waitRecipeTestCaseServiceName, wait.RecipeArgName, recipeStr, wait.ValueFieldArgName, waitValueField, wait.AssertionArgName, waitAssertion, wait.TargetArgName, waitTargetValue, wait.IntervalArgName, waitInterval, wait.TimeoutArgName, waitTimeout)
}

func (t *waitWithNamedArgsTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *waitWithNamedArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResult := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResult, interpretationResult.String())

	expectedExecutionResult := `Assertion passed with following:
Request had response code '200' and body "{\"value\": \"Hello world!\"}", with extracted fields:
'extract.key': "Hello world!"`

	require.Contains(t, *executionResult, expectedExecutionResult)
}
