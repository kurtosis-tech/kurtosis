package test_engine

// type getHttpRequestRecipeNoExtractorTestCase struct {
// 	*testing.T
// 	serviceNetwork    *service_network.MockServiceNetwork
// 	runtimeValueStore *runtime_value_store.RuntimeValueStore
// }

// func (suite *KurtosisTypeConstructorTestSuite) TestGetHttpRequestRecipeNoExtractor() {
// 	suite.serviceNetwork.EXPECT().HttpRequestService(
// 		mock.Anything,
// 		string(testServiceName),
// 		testPrivatePortId,
// 		"GET",
// 		"",
// 		"/test",
// 		"",
// 		testEmptyHeaders,
// 	).Times(1).Return(
// 		&http.Response{
// 			Status:           "200 OK",
// 			StatusCode:       200,
// 			Proto:            "HTTP/1.0",
// 			ProtoMajor:       1,
// 			ProtoMinor:       0,
// 			Header:           nil,
// 			Body:             io.NopCloser(strings.NewReader("")),
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

// 	suite.run(&getHttpRequestRecipeNoExtractorTestCase{
// 		T:                 suite.T(),
// 		serviceNetwork:    suite.serviceNetwork,
// 		runtimeValueStore: suite.runtimeValueStore,
// 	})
// }

// func (t *getHttpRequestRecipeNoExtractorTestCase) GetStarlarkCode() string {
// 	return fmt.Sprintf("%s(%s=%q, %s=%q)", recipe.GetHttpRecipeTypeName, recipe.PortIdAttr, testPrivatePortId, recipe.EndpointAttr, "/test")
// }

// func (t *getHttpRequestRecipeNoExtractorTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
// 	getHttpRequestRecipe, ok := typeValue.(*recipe.GetHttpRequestRecipe)
// 	require.True(t, ok)

// 	_, err := getHttpRequestRecipe.Execute(context.Background(), t.serviceNetwork, t.runtimeValueStore, testServiceName)
// 	require.NoError(t, err)

// 	returnValue, interpretationErr := getHttpRequestRecipe.CreateStarlarkReturnValue("result-fake-uuid")
// 	require.Nil(t, interpretationErr)
// 	expectedInterpretationResult := `{"body": "{{kurtosis:result-fake-uuid:body.runtime_value}}", "code": "{{kurtosis:result-fake-uuid:code.runtime_value}}"}`
// 	require.Equal(t, expectedInterpretationResult, returnValue.String())
// }
