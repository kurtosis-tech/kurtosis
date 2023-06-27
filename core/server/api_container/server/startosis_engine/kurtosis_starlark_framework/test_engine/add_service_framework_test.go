package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type addServiceTestCase struct {
	*testing.T
}

func newAddServiceTestCase(t *testing.T) *addServiceTestCase {
	return &addServiceTestCase{
		T: t,
	}
}

func (t *addServiceTestCase) GetId() string {
	return add_service.AddServiceBuiltinName
}

func (t *addServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()

	serviceNetwork.EXPECT().AddService(
		mock.Anything,
		TestServiceName,
		mock.MatchedBy(func(serviceConfig *service.ServiceConfig) bool {
			expectedServiceConfig := service.NewServiceConfig(
				TestContainerImageName,
				map[string]*port_spec.PortSpec{},
				map[string]*port_spec.PortSpec{},
				nil,
				nil,
				map[string]string{},
				nil,
				0,
				0,
				service_config.DefaultPrivateIPAddrPlaceholder,
				0,
				0,
				service_config.DefaultSubnetwork,
			)

			actualServiceConfig := serviceConfig
			assert.Equal(t, expectedServiceConfig, actualServiceConfig)
			return true
		}),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(TestServiceName, TestServiceUuid, TestEnclaveUuid, nil, string(TestServiceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		nil,
	)

	return add_service.NewAddService(serviceNetwork, runtimeValueStore)
}

func (t *addServiceTestCase) GetStarlarkCode() string {
	serviceConfig := fmt.Sprintf("ServiceConfig(image=%q)", TestContainerImageName)
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, add_service.AddServiceBuiltinName, add_service.ServiceNameArgName, TestServiceName, add_service.ServiceConfigArgName, serviceConfig)
}

func (t *addServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	serviceObj, ok := interpretationResult.(*kurtosis_types.Service)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.NotNil(t, serviceObj)
	expectedServiceObj := fmt.Sprintf(`Service\(hostname = "{{kurtosis:[0-9a-f]{32}:hostname.runtime_value}}", ip_address = "{{kurtosis:[0-9a-f]{32}:ip_address.runtime_value}}", name = "%v", ports = {}\)`, TestServiceName)
	require.Regexp(t, expectedServiceObj, serviceObj.String())

	expectedExecutionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", TestServiceName, TestServiceUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
