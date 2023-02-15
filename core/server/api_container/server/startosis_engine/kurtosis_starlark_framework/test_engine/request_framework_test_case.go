package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
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

const (
	request_serviceName = service.ServiceName("web-server")
	request_portId      = "port_id"
	request_method      = "GET"
	request_contentType = ""
	request_endpoint    = "/"
	request_body        = ""

	request_responseBody = `{"value": "Hello World!"}`
)

type requestTestCase struct {
	*testing.T
}

func newRequestTestCase(t *testing.T) *requestTestCase {
	return &requestTestCase{
		T: t,
	}
}

func (t *requestTestCase) GetId() string {
	return request.RequestBuiltinName
}

func (t *requestTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(request_serviceName),
		request_portId,
		request_method,
		request_contentType,
		request_endpoint,
		request_body,
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(request_responseBody)),
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

func (t *requestTestCase) GetStarlarkCode() string {
	recipe := fmt.Sprintf(`GetHttpRequestRecipe(port_id=%q, service_name=%q, endpoint=%q, extract={"key": ".value"})`, request_portId, request_serviceName, request_endpoint)
	return fmt.Sprintf("%s(%s=%s)", request.RequestBuiltinName, request.RecipeArgName, recipe)
}

func (t *requestTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResultMap := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResultMap, interpretationResult.String())

	expectedExecutionResult := `Request had response code '200' and body "{\"value\": \"Hello World!\"}", with extracted fields:
'extract.key': "Hello World!"`
	require.Equal(t, expectedExecutionResult, *executionResult)
}
