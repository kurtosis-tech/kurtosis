package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	waitServiceName = service.ServiceName("web-server")
)

//For a short period (until we deprecate recipe.service_name) the wait instruction will have a
//dynamic first parameter which will accept the current 'recipe' argument and a new 'service_name' argument
//In the waitTestCase1 we test the current behaviour, it means receiving an 'recipe' as the first argument
//In this test case we test that 'service_name' is also accepted as the first parameter, and it is used
//in the serviceNetwork.HttpRequestService call
type waitTestCase2 struct {
	*testing.T
}

func newWaitTestCase2(t *testing.T) *waitTestCase2 {
	return &waitTestCase2{
		T: t,
	}
}

func (t *waitTestCase2) GetId() string {
	return wait.WaitBuiltinName
}

func (t *waitTestCase2) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(waitServiceName),
		waitRecipePortId,
		waitRecipeMethod,
		waitRecipeContentType,
		waitRecipeEndpoint,
		waitRecipeBody,
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

	return wait.NewWait(serviceNetwork, runtimeValueStore)
}

func (t *waitTestCase2) GetStarlarkCode() string {
	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
	return fmt.Sprintf("%s(%s=%q, %s=%s, %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)", wait.WaitBuiltinName, wait.ServiceNameArgName, waitServiceName, wait.RecipeArgName, recipeStr, wait.ValueFieldArgName, waitValueField, wait.AssertionArgName, waitAssertion, wait.TargetArgName, waitTargetValue, wait.IntervalArgName, waitInterval, wait.TimeoutArgName, waitTimeout)
}

func (t *waitTestCase2) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResult := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResult, interpretationResult.String())

	expectedExecutionResult := `Assertion passed with following:
Request had response code '200' and body "{\"value\": \"Hello world!\"}", with extracted fields:
'extract.key': "Hello world!"`

	require.Contains(t, *executionResult, expectedExecutionResult)
}
