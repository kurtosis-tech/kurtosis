package test_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/recipe"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	getHttpRequestRecipeResponseJson = `{"value": "Hello world!"}`
)

type getHttpRequestRecipeTestCase struct {
	*testing.T
	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func newGetHttpRequestRecipeTestCase(t *testing.T) *getHttpRequestRecipeTestCase {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork := service_network.NewMockServiceNetwork(t)
	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(TestServiceName),
		TestPrivatePortId,
		"GET",
		"",
		"/test",
		"",
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(getHttpRequestRecipeResponseJson)),
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

	return &getHttpRequestRecipeTestCase{
		T:                 t,
		serviceNetwork:    serviceNetwork,
		runtimeValueStore: runtimeValueStore,
	}
}

func (t *getHttpRequestRecipeTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", recipe.GetHttpRecipeTypeName, "with_extractors")
}

func (t *getHttpRequestRecipeTestCase) GetStarlarkCode() string {
	extractors := `{"result": ".value"}`
	return fmt.Sprintf("%s(%s=%q, %s=%q, %s=%s)", recipe.GetHttpRecipeTypeName, recipe.PortIdAttr, TestPrivatePortId, recipe.EndpointAttr, "/test", recipe.ExtractAttr, extractors)
}

func (t *getHttpRequestRecipeTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	getHttpRequestRecipe, ok := typeValue.(*recipe.GetHttpRequestRecipe)
	require.True(t, ok)

	_, err := getHttpRequestRecipe.Execute(context.Background(), t.serviceNetwork, t.runtimeValueStore, TestServiceName)
	require.NoError(t, err)

	returnValue, interpretationErr := getHttpRequestRecipe.CreateStarlarkReturnValue("result-fake-uuid")
	require.Nil(t, interpretationErr)
	expectedInterpretationResult := `{"body": "{{kurtosis:result-fake-uuid:body.runtime_value}}", "code": "{{kurtosis:result-fake-uuid:code.runtime_value}}", "extract.result": "{{kurtosis:result-fake-uuid:extract.result.runtime_value}}"}`
	require.Equal(t, expectedInterpretationResult, returnValue.String())
}
