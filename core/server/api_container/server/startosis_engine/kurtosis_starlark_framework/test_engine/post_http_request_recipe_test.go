package test_engine

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	postHttpRequestRecipeResponseJson = `{"value": "Hello world!"}`
)

type postHttpRequestRecipeTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisTypeConstructorTestSuite) TestPostHttpRequestRecipe() {
	suite.serviceNetwork.EXPECT().HttpRequestServiceObject(
		mock.Anything,
		testService,
		testPrivatePortId,
		"POST",
		"application/json",
		"/test",
		"{}",
		map[string]string{"key": "value"},
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(postHttpRequestRecipeResponseJson)),
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

	suite.run(&postHttpRequestRecipeTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *postHttpRequestRecipeTestCase) GetStarlarkCode() string {
	extractors := `{"result": ".value"}`
	headers := `{"key": "value"}`
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%q, %s=%q, %s=%s, %s=%s)", recipe.PostHttpRecipeTypeName, recipe.PortIdAttr, testPrivatePortId, recipe.EndpointAttr, "/test", recipe.RequestBodyAttr, "{}", recipe.ContentTypeAttr, "application/json", recipe.ExtractAttr, extractors, recipe.HeadersAttr, headers)
}

func (t *postHttpRequestRecipeTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	postHttpRequestRecipe, ok := typeValue.(*recipe.PostHttpRequestRecipe)
	require.True(t, ok)

	_, err := postHttpRequestRecipe.Execute(context.Background(), t.serviceNetwork, t.runtimeValueStore, testService)
	require.NoError(t, err)

	returnValue, interpretationErr := postHttpRequestRecipe.CreateStarlarkReturnValue("result-fake-uuid")
	require.Nil(t, interpretationErr)
	expectedInterpretationResult := `{"body": "{{kurtosis:result-fake-uuid:body.runtime_value}}", "code": "{{kurtosis:result-fake-uuid:code.runtime_value}}", "extract.result": "{{kurtosis:result-fake-uuid:extract.result.runtime_value}}"}`
	require.Equal(t, expectedInterpretationResult, returnValue.String())
}
