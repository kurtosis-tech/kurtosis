package test_engine

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/request"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

type requestWithPositionalArgsTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestRequestWithPositionalArgs() {
	suite.serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(requestTestCaseServiceName),
	).Times(1).Return(
		requestTestCaseService,
		nil,
	)

	suite.serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		requestTestCaseService,
		requestPortId,
		requestMethod,
		requestContentType,
		requestEndpoint,
		requestBody,
		testEmptyHeaders,
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

	suite.run(&requestWithPositionalArgsTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *requestWithPositionalArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return request.NewRequest(t.serviceNetwork, t.runtimeValueStore)
}

func (t *requestWithPositionalArgsTestCase) GetStarlarkCode() string {
	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, endpoint=%q, extract={"key": ".value"})`, requestPortId, requestEndpoint)
	return fmt.Sprintf("%s(%q, %s)", request.RequestBuiltinName, requestTestCaseServiceName, recipe)
}

func (t *requestWithPositionalArgsTestCase) GetStarlarkCodeForAssertion() string {
	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, endpoint=%q, extract={"key": ".value"})`, requestPortId, requestEndpoint)
	return fmt.Sprintf("%s(%s=%q, %s=%s)", request.RequestBuiltinName, request.ServiceNameArgName, requestTestCaseServiceName, request.RecipeArgName, recipe)
}

func (t *requestWithPositionalArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	expectedExecutionResult := `Request had response code '200' and body "{\"value\": \"Hello World!\"}", with extracted fields:
'extract.key': "Hello World!"`
	require.Equal(t, expectedExecutionResult, *executionResult)
}
