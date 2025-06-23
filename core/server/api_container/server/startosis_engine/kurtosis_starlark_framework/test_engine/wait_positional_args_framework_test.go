package test_engine

// // This test case is for testing positional arguments retro-compatibility for those script
// // that are using the recipe value as the first positional argument
// type waitWithPositionalArgsTestCase struct {
// 	*testing.T
// 	serviceNetwork    *service_network.MockServiceNetwork
// 	runtimeValueStore *runtime_value_store.RuntimeValueStore
// }

// func (suite *KurtosisPlanInstructionTestSuite) TestWaitWithPositionalArgs() {
// 	suite.serviceNetwork.EXPECT().HttpRequestService(
// 		mock.Anything,
// 		string(waitRecipeTestCaseServiceName),
// 		waitRecipePortId,
// 		waitRecipeMethod,
// 		waitRecipeContentType,
// 		waitRecipeEndpoint,
// 		waitRecipeBody,
// 		testEmptyHeaders,
// 	).Times(1).Return(
// 		&http.Response{
// 			Status:           "200 OK",
// 			StatusCode:       200,
// 			Proto:            "HTTP/1.0",
// 			ProtoMajor:       1,
// 			ProtoMinor:       0,
// 			Header:           nil,
// 			Body:             io.NopCloser(strings.NewReader(waitRecipeResponseBody)),
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

// 	suite.run(&waitWithPositionalArgsTestCase{
// 		T:                 suite.T(),
// 		serviceNetwork:    suite.serviceNetwork,
// 		runtimeValueStore: suite.runtimeValueStore,
// 	})
// }

// func (t *waitWithPositionalArgsTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
// 	return wait.NewWait(t.serviceNetwork, t.runtimeValueStore)
// }

// func (t *waitWithPositionalArgsTestCase) GetStarlarkCode() string {
// 	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
// 	return fmt.Sprintf("%s(%q, %s, %q, %q, %s, %q, %q)", wait.WaitBuiltinName, waitRecipeTestCaseServiceName, recipeStr, waitValueField, waitAssertion, waitTargetValue, waitInterval, waitTimeout)
// }

// func (t *waitWithPositionalArgsTestCase) GetStarlarkCodeForAssertion() string {
// 	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, waitRecipePortId, waitRecipeEndpoint, waitRecipeBody, waitRecipeContentType)
// 	return fmt.Sprintf("%s(%s=%q, %s=%s, %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)", wait.WaitBuiltinName, wait.ServiceNameArgName, waitRecipeTestCaseServiceName, wait.RecipeArgName, recipeStr, wait.ValueFieldArgName, waitValueField, wait.AssertionArgName, waitAssertion, wait.TargetArgName, waitTargetValue, wait.IntervalArgName, waitInterval, wait.TimeoutArgName, waitTimeout)
// }

// func (t *waitWithPositionalArgsTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
// 	expectedInterpretationResult := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
// 	require.Regexp(t, expectedInterpretationResult, interpretationResult.String())

// 	expectedExecutionResult := `Assertion passed with following:
// Request had response code '200' and body "{\"value\": \"Hello world!\"}", with extracted fields:
// 'extract.key': "Hello world!"`

// 	require.Contains(t, *executionResult, expectedExecutionResult)
// }
