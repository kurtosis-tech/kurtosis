package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

const (
	requestTestCaseServiceName = service.ServiceName("web-server")
	requestPortId              = "port_id"
	requestMethod              = "GET"
	requestContentType         = ""
	requestEndpoint            = "/"
	requestBody                = ""

	requestResponseBody = `{"value": "Hello World!"}`
)

// type requestWithNamedArgsTestCase struct {
// 	*testing.T
// 	serviceNetwork    *service_network.MockServiceNetwork
// 	runtimeValueStore *runtime_value_store.RuntimeValueStore
// }

// func (suite *KurtosisPlanInstructionTestSuite) TestRequestWithNamedArgs() {
// 	suite.serviceNetwork.EXPECT().GetService(
// 		mock.Anything,
// 		string(requestTestCaseServiceName),
// 	).Times(1).Return(
// 		suite.serviceNetwork.GetService(mock.Anything, string(requestTestCaseServiceName)),
// 		nil,
// 	)

// 	suite.serviceNetwork.EXPECT().HttpRequestService(
// 		mock.Anything,
// 		string(requestTestCaseServiceName),
// 		requestPortId,
// 		requestMethod,
// 		requestContentType,
// 		requestEndpoint,
// 		requestBody,
// 		testEmptyHeaders,
// 	).Times(1).Return(
// 		&http.Response{
// 			Status:           "200 OK",
// 			StatusCode:       200,
// 			Proto:            "HTTP/1.0",
// 			ProtoMajor:       1,
// 			ProtoMinor:       0,
// 			Header:           nil,
// 			Body:             io.NopCloser(strings.NewReader(requestResponseBody)),
// 			ContentLength:    -1,
// 			TransferEncoding: nil,
// 			Close:            false,
// 			Uncompressed:     false,
// 			Trailer:          nil,
// 			Request:          nil,
// 			TLS:              nil,
// 		},
// 		nil,
// 	)

// 	suite.run(&requestWithNamedArgsTestCase{
// 		T:                 suite.T(),
// 		serviceNetwork:    suite.serviceNetwork,
// 		runtimeValueStore: suite.runtimeValueStore,
// 	})
// }

// func (t *requestWithNamedArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
// 	return request.NewRequest(t.serviceNetwork, t.runtimeValueStore)
// }

// func (t *requestWithNamedArgsTestCase) GetStarlarkCode() string {
// 	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, endpoint=%q, extract={"key": ".value"})`, requestPortId, requestEndpoint)
// 	return fmt.Sprintf("%s(%s=%q, %s=%s)", request.RequestBuiltinName, request.ServiceNameArgName, requestTestCaseServiceName, request.RecipeArgName, recipe)
// }

// func (t *requestWithNamedArgsTestCase) GetStarlarkCodeForAssertion() string {
// 	return ""
// }

// func (t *requestWithNamedArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
// 	expectedInterpretationResultMap := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
// 	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

// 	expectedExecutionResult := `Request had response code '200' and body "{\"value\": \"Hello World!\"}", with extracted fields:
// 'extract.key': "Hello World!"`
// 	require.Equal(t, expectedExecutionResult, *executionResult)
// }
