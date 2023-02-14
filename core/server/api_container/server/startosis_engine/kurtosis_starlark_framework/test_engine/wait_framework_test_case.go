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
	wait_assertion   = "=="
	wait_interval    = "1s"
	wait_targetValue = "200"
	wait_timeout     = "5s"
	wait_valueField  = "code"

	wait_recipe_portId      = "http-port"
	wait_recipe_serviceName = service.ServiceName("web-server")
	wait_recipe_method      = "POST"
	wait_recipe_endpoint    = "/"
	wait_recipe_body        = "{}"
	wait_recipe_contentType = "application/json"

	wait_recipe_responseBody = `{"value": "Hello world!"}`
)

type waitTestCase struct {
	*testing.T
}

func newWaitTestCase(t *testing.T) *waitTestCase {
	return &waitTestCase{
		T: t,
	}
}

func (t *waitTestCase) GetId() string {
	return wait.WaitBuiltinName
}

func (t *waitTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(wait_recipe_serviceName),
		wait_recipe_portId,
		wait_recipe_method,
		wait_recipe_contentType,
		wait_recipe_endpoint,
		wait_recipe_body,
	).Times(1).Return(
		&http.Response{
			Status:           "200 OK",
			StatusCode:       200,
			Proto:            "HTTP/1.0",
			ProtoMajor:       1,
			ProtoMinor:       0,
			Header:           nil,
			Body:             io.NopCloser(strings.NewReader(wait_recipe_responseBody)),
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

func (t *waitTestCase) GetStarlarkCode() string {
	recipeStr := fmt.Sprintf(`PostHttpRequestRecipe(port_id=%q, service_name=%q, endpoint=%q, body=%q, content_type=%q, extract={"key": ".value"})`, wait_recipe_portId, wait_recipe_serviceName, wait_recipe_endpoint, wait_recipe_body, wait_recipe_contentType)
	return fmt.Sprintf("%s(%s=%s, %s=%q, %s=%q, %s=%s, %s=%q, %s=%q)", wait.WaitBuiltinName, wait.RecipeArgName, recipeStr, wait.ValueFieldArgName, wait_valueField, wait.AssertionArgName, wait_assertion, wait.TargetArgName, wait_targetValue, wait.IntervalArgName, wait_interval, wait.TimeoutArgName, wait_timeout)
}

func (t *waitTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	expectedInterpretationResult := `{"body": "{{kurtosis:[0-9a-f]{32}:body.runtime_value}}", "code": "{{kurtosis:[0-9a-f]{32}:code.runtime_value}}", "extract.key": "{{kurtosis:[0-9a-f]{32}:extract.key.runtime_value}}"}`
	require.Regexp(t, expectedInterpretationResult, interpretationResult.String())

	expectedExecutionResult := `Assertion passed with following:
Request had response code '200' and body "{\"value\": \"Hello world!\"}", with extracted fields:
'extract.key': "Hello world!"`

	require.Contains(t, *executionResult, expectedExecutionResult)
}
