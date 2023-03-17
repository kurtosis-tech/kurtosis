package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type addServicesTestCase struct {
	*testing.T
}

func newAddServicesTestCase(t *testing.T) *addServicesTestCase {
	return &addServicesTestCase{
		T: t,
	}
}

func (t *addServicesTestCase) GetId() string {
	return add_service.AddServicesBuiltinName
}

func (t *addServicesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().StartServices(
		mock.Anything,
		mock.MatchedBy(func(configs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig) bool {
			require.Len(t, configs, 2)
			require.Contains(t, configs, TestServiceName)
			require.Contains(t, configs, TestServiceName2)

			expectedServiceConfig1 := services.NewServiceConfigBuilder(TestContainerImageName).WithSubnetwork(string(TestSubnetwork)).Build()
			actualServiceConfig1 := services.NewServiceConfigBuilderFromServiceConfig(configs[TestServiceName]).Build()
			assert.Equal(t, expectedServiceConfig1, actualServiceConfig1)
			actualServiceConfig2 := services.NewServiceConfigBuilderFromServiceConfig(configs[TestServiceName2]).Build()
			expectedServiceConfig2 := services.NewServiceConfigBuilder(TestContainerImageName).WithCpuAllocationMillicpus(TestCpuAllocation).WithMemoryAllocationMegabytes(TestMemoryAllocation).Build()
			assert.Equal(t, expectedServiceConfig2, actualServiceConfig2)
			return true
		}),
		mock.Anything,
	).Times(1).Return(
		map[service.ServiceName]*service.Service{
			TestServiceName:  service.NewService(service.NewServiceRegistration(TestServiceName, TestServiceUuid, TestEnclaveUuid, nil, string(TestServiceName)), container_status.ContainerStatus_Running, nil, nil, nil),
			TestServiceName2: service.NewService(service.NewServiceRegistration(TestServiceName2, TestServiceUuid2, TestEnclaveUuid, nil, string(TestServiceName2)), container_status.ContainerStatus_Running, nil, nil, nil),
		},
		map[service.ServiceName]error{},
		nil,
	)

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(TestServiceName),
		TestReadyConditionsRecipePortId,
		TestGetRequestMethod,
		"",
		TestReadyConditionsRecipeEndpoint,
		"",
	).Times(1).Return(&http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Request: &http.Request{
			Method: TestGetRequestMethod,
			URL:    &url.URL{},
		},
		Close:            true,
		ContentLength:    -1,
		Body:             io.NopCloser(strings.NewReader("{}")),
		Trailer:          nil,
		TransferEncoding: nil,
		Uncompressed:     true,
		TLS:              nil,
	}, nil)

	serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(TestServiceName2),
		TestReadyConditions2RecipePortId,
		TestGetRequestMethod,
		"",
		TestReadyConditions2RecipeEndpoint,
		"",
	).Times(1).Return(&http.Response{
		Status:     "201 OK",
		StatusCode: 201,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Request: &http.Request{
			Method: TestGetRequestMethod,
			URL: &url.URL{
				Path:        "",
				Scheme:      "",
				Opaque:      "",
				User:        nil,
				Host:        "",
				RawPath:     "",
				ForceQuery:  false,
				RawQuery:    "",
				Fragment:    "",
				RawFragment: "",
			},
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           http.Header{},
			Body:             nil,
			GetBody:          nil,
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            false,
			Host:             "",
			Form:             nil,
			PostForm:         nil,
			MultipartForm:    nil,
			Trailer:          nil,
			RemoteAddr:       "",
			RequestURI:       "",
			TLS:              nil,
			Cancel:           nil,
			Response:         nil,
		},
		Close:            true,
		ContentLength:    -1,
		Body:             io.NopCloser(strings.NewReader("{}")),
		Trailer:          nil,
		TransferEncoding: nil,
		Uncompressed:     true,
		TLS:              nil,
	}, nil)

	return add_service.NewAddServices(serviceNetwork, runtimeValueStore)
}

func (t *addServicesTestCase) GetStarlarkCode() string {
	service1ReadyConditionsScriptPart := getDefaultReadyConditionsScriptPart()
	service2ReadyConditionsScriptPart := getCustomReadyConditionsScripPart(
		TestReadyConditions2RecipePortId,
		TestReadyConditions2RecipeEndpoint,
		TestReadyConditions2RecipeExtract,
		TestReadyConditions2Field,
		TestReadyConditions2Assertion,
		TestReadyConditions2Target,
		TestReadyConditions2Interval,
		TestReadyConditions2Timeout,
	)
	serviceConfig1 := fmt.Sprintf("ServiceConfig(image=%q, subnetwork=%q, ready_conditions=%s)", TestContainerImageName, TestSubnetwork, service1ReadyConditionsScriptPart)
	serviceConfig2 := fmt.Sprintf("ServiceConfig(image=%q, cpu_allocation=%d, memory_allocation=%d, ready_conditions=%s)", TestContainerImageName, TestCpuAllocation, TestMemoryAllocation, service2ReadyConditionsScriptPart)
	return fmt.Sprintf(`%s(%s={%q: %s, %q: %s})`, add_service.AddServicesBuiltinName, add_service.ConfigsArgName, TestServiceName, serviceConfig1, TestServiceName2, serviceConfig2)
}

func (t *addServicesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServicesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	resultDict, ok := interpretationResult.(*starlark.Dict)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.Equal(t, resultDict.Len(), 2)
	require.Contains(t, resultDict.Keys(), starlark.String(TestServiceName))
	require.Contains(t, resultDict.Keys(), starlark.String(TestServiceName2))

	require.Contains(t, *executionResult, "Successfully added the following '2' services:")
	require.Contains(t, *executionResult, fmt.Sprintf("Service '%s' added with UUID '%s'", TestServiceName, TestServiceUuid))
	require.Contains(t, *executionResult, fmt.Sprintf("Service '%s' added with UUID '%s'", TestServiceName2, TestServiceUuid2))
}
