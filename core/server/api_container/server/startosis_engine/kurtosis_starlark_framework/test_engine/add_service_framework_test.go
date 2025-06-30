package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"testing"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

type addServiceTestCase struct {
	*testing.T
	serviceNetwork               *service_network.MockServiceNetwork
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestAddService() {
	suite.serviceNetwork.EXPECT().ExistServiceRegistration(testServiceName).Times(1).Return(false, nil)
	suite.serviceNetwork.EXPECT().AddService(
		mock.Anything,
		testServiceName,
		mock.MatchedBy(func(serviceConfig *service.ServiceConfig) bool {
			expectedServiceConfig, err := service.CreateServiceConfig(testContainerImageName, nil, nil, nil, map[string]*port_spec.PortSpec{}, map[string]*port_spec.PortSpec{}, nil, nil, map[string]string{}, nil, nil, 0, 0, service_config.DefaultPrivateIPAddrPlaceholder, 0, 0, map[string]string{}, nil, nil, map[string]string{}, image_download_mode.ImageDownloadMode_Missing, true, false)
			require.NoError(suite.T(), err)

			actualServiceConfig := serviceConfig
			suite.Assert().Equal(expectedServiceConfig, actualServiceConfig)
			return true
		}),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(testServiceName, testServiceUuid, testEnclaveUuid, nil, string(testServiceName)), nil, nil, nil, container.NewContainer(container.ContainerStatus_Running, "", nil, nil, nil)),
		nil,
	)

	suite.run(&addServiceTestCase{
		T:                            suite.T(),
		serviceNetwork:               suite.serviceNetwork,
		runtimeValueStore:            suite.runtimeValueStore,
		packageContentProvider:       suite.packageContentProvider,
		interpretationTimeValueStore: suite.interpretationTimeValueStore,
	})
}

func (t *addServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return add_service.NewAddService(
		t.serviceNetwork,
		t.runtimeValueStore,
		testModulePackageId,
		t.packageContentProvider,
		testNoPackageReplaceOptions,
		t.interpretationTimeValueStore,
		image_download_mode.ImageDownloadMode_Missing)
}

func (t *addServiceTestCase) GetStarlarkCode() string {
	serviceConfig := fmt.Sprintf("ServiceConfig(image=%q)", testContainerImageName)
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, add_service.AddServiceBuiltinName, add_service.ServiceNameArgName, testServiceName, add_service.ServiceConfigArgName, serviceConfig)
}

func (t *addServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	serviceObj, ok := interpretationResult.(*kurtosis_types.Service)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.NotNil(t, serviceObj)
	expectedServiceObj := fmt.Sprintf(`Service\(name="%v", hostname="{{kurtosis:[0-9a-f]{32}:hostname.runtime_value}}", ip_address="{{kurtosis:[0-9a-f]{32}:ip_address.runtime_value}}", ports={}\)`, testServiceName)
	require.Regexp(t, expectedServiceObj, serviceObj.String())

	expectedExecutionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", testServiceName, testServiceUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
