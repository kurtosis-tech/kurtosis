package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/request"
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

//This test case is for testing positional arguments
type requestTestCase2 struct {
	*testing.T
}

func newRequestTestCase2(t *testing.T) *requestTestCase2 {
	return &requestTestCase2{
		T: t,
	}
}

func (t *requestTestCase2) GetId() string {
	return request.RequestBuiltinName
}

func (t *requestTestCase2) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(requestTestCaseServiceName),
		requestPortId,
		requestMethod,
		requestContentType,
		requestEndpoint,
		requestBody,
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(requestResponseBody)),
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

	return request.NewRequest(serviceNetwork, runtimeValueStore)
}

func (t *requestTestCase2) GetStarlarkCode() string {
	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, endpoint=%q, extract={"key": ".value"})`, requestPortId, requestEndpoint)
	return fmt.Sprintf("%s(%q, %s)", request.RequestBuiltinName, requestTestCaseServiceName, recipe)
}

func (t *requestTestCase2) GetStarlarkCodeForAssertion() string {
	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, endpoint=%q, extract={"key": ".value"})`, requestPortId, requestEndpoint)
	return fmt.Sprintf("%s(%s=%q, %s=%s)", request.RequestBuiltinName, request.ServiceNameArgName, requestTestCaseServiceName, request.RecipeArgName, recipe)
}

func (t *requestTestCase2) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	expectedExecutionResult := `Request had response code '200' and body "{\"value\": \"Hello World!\"}", with extracted fields:
'extract.key': "Hello World!"`
	require.Equal(t, expectedExecutionResult, *executionResult)
}
