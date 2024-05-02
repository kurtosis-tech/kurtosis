package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

type addServicesTestCase struct {
	*testing.T
	serviceNetwork               *service_network.MockServiceNetwork
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestAddServices() {
	suite.serviceNetwork.EXPECT().ExistServiceRegistration(testServiceName).Times(1).Return(false, nil)
	suite.serviceNetwork.EXPECT().ExistServiceRegistration(testServiceName2).Times(1).Return(false, nil)
	suite.serviceNetwork.EXPECT().UpdateServices(
		mock.Anything,
		map[service.ServiceName]*service.ServiceConfig{},
		mock.Anything,
	).Times(1).Return(
		map[service.ServiceName]*service.Service{},
		map[service.ServiceName]error{},
		nil,
	)
	suite.serviceNetwork.EXPECT().AddServices(
		mock.Anything,
		mock.MatchedBy(func(configs map[service.ServiceName]*service.ServiceConfig) bool {
			suite.Require().Len(configs, 2)
			suite.Require().Contains(configs, testServiceName)
			suite.Require().Contains(configs, testServiceName2)

			expectedServiceConfig1, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true)
			require.NoError(suite.T(), err)

			actualServiceConfig1 := configs[testServiceName]
			suite.Assert().Equal(expectedServiceConfig1, actualServiceConfig1)

			expectedServiceConfig2, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, testCpuAllocation, testMemoryAllocation, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true)
			require.NoError(suite.T(), err)

			actualServiceConfig2 := configs[testServiceName2]
			suite.Assert().Equal(expectedServiceConfig2, actualServiceConfig2)
			return true
		}),
		mock.Anything,
	).Times(1).Return(
		map[service.ServiceName]*service.Service{
			testServiceName:  service.NewService(service.NewServiceRegistration(testServiceName, testServiceUuid, testEnclaveUuid, nil, string(testServiceName)), nil, nil, nil, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil)),
			testServiceName2: service.NewService(service.NewServiceRegistration(testServiceName2, testServiceUuid2, testEnclaveUuid, nil, string(testServiceName2)), nil, nil, nil, container.NewContainer(container.ContainerStatus_Running, testContainerImageName, nil, nil, nil)),
		},
		map[service.ServiceName]error{},
		nil,
	)

	suite.serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(testServiceName),
		testReadyConditionsRecipePortId,
		testGetRequestMethod,
		"",
		testReadyConditionsRecipeEndpoint,
		"",
	).Times(1).Return(&http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Request: &http.Request{ //nolint:exhaustruct
			Method: testGetRequestMethod,
			URL:    &url.URL{}, //nolint:exhaustruct
		},
		Close:            true,
		ContentLength:    -1,
		Body:             io.NopCloser(strings.NewReader("{}")),
		Trailer:          nil,
		TransferEncoding: nil,
		Uncompressed:     true,
		TLS:              nil,
	}, nil)

	suite.serviceNetwork.EXPECT().HttpRequestService(
		mock.Anything,
		string(testServiceName2),
		testReadyConditions2RecipePortId,
		testGetRequestMethod,
		"",
		testReadyConditions2RecipeEndpoint,
		"",
	).Times(1).Return(&http.Response{
		Status:     "201 OK",
		StatusCode: 201,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Request: &http.Request{
			Method: testGetRequestMethod,
			URL: &url.URL{ //nolint:exhaustruct
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

	suite.run(&addServicesTestCase{
		T:                            suite.T(),
		serviceNetwork:               suite.serviceNetwork,
		runtimeValueStore:            suite.runtimeValueStore,
		packageContentProvider:       suite.packageContentProvider,
		interpretationTimeValueStore: suite.interpretationTimeValueStore,
	})
}

func (t *addServicesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return add_service.NewAddServices(
		t.serviceNetwork,
		t.runtimeValueStore,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		t.interpretationTimeValueStore,
		image_download_mode.ImageDownloadMode_Missing)
}

func (t *addServicesTestCase) GetStarlarkCode() string {
	service1ReadyConditionsScriptPart := getDefaultReadyConditionsScriptPart()
	service2ReadyConditionsScriptPart := getCustomReadyConditionsScripPart(
		testReadyConditions2RecipePortId,
		testReadyConditions2RecipeEndpoint,
		testReadyConditions2RecipeExtract,
		testReadyConditions2Field,
		testReadyConditions2Assertion,
		testReadyConditions2Target,
		testReadyConditions2Interval,
		testReadyConditions2Timeout,
	)
	serviceConfig1 := fmt.Sprintf("ServiceConfig(image=%q, ready_conditions=%s)", testContainerImageName, service1ReadyConditionsScriptPart)
	serviceConfig2 := fmt.Sprintf("ServiceConfig(image=%q, cpu_allocation=%d, memory_allocation=%d, ready_conditions=%s)", testContainerImageName, testCpuAllocation, testMemoryAllocation, service2ReadyConditionsScriptPart)
	return fmt.Sprintf(`%s(%s={%q: %s, %q: %s})`, add_service.AddServicesBuiltinName, add_service.ConfigsArgName, testServiceName, serviceConfig1, testServiceName2, serviceConfig2)
}

func (t *addServicesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServicesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	resultDict, ok := interpretationResult.(*starlark.Dict)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.Equal(t, resultDict.Len(), 2)
	require.Contains(t, resultDict.Keys(), starlark.String(testServiceName))
	require.Contains(t, resultDict.Keys(), starlark.String(testServiceName2))

	require.Contains(t, *executionResult, "Successfully added the following '2' services:")
	require.Contains(t, *executionResult, fmt.Sprintf("Service '%s' added with UUID '%s'", testServiceName, testServiceUuid))
	require.Contains(t, *executionResult, fmt.Sprintf("Service '%s' added with UUID '%s'", testServiceName2, testServiceUuid2))
}
