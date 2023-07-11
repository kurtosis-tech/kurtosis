package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_update_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type updateServiceTestCase struct {
	*testing.T
}

func newUpdateServiceTestCase(t *testing.T) *updateServiceTestCase {
	return &updateServiceTestCase{
		T: t,
	}
}

func (t *updateServiceTestCase) GetId() string {
	return add_update_service.UpdateServiceBuiltinName
}

func (t *updateServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(TestServiceName),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(TestServiceName, TestServiceUuid, TestEnclaveUuid, nil, string(TestServiceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		nil,
	)

	serviceNetwork.EXPECT().UpdateServices(
		mock.Anything,
		map[service.ServiceName]*service.ServiceConfig{
			TestServiceName: service.NewServiceConfig(
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
			),
		},
		1,
	).Times(1).Return(
		map[service.ServiceName]*service.Service{
			TestServiceName: service.NewService(service.NewServiceRegistration(TestServiceName, TestServiceUuid, TestEnclaveUuid, nil, string(TestServiceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		},
		map[service.ServiceName]error{},
		nil,
	)

	return add_update_service.NewUpdateService(serviceNetwork)
}

func (t *updateServiceTestCase) GetStarlarkCode() string {
	updateServiceConfig := fmt.Sprintf("ServiceConfig(image=%q)", TestContainerImageName)
	return fmt.Sprintf("%s(%s=%q, %s=%s)", add_update_service.UpdateServiceBuiltinName, add_update_service.ServiceNameArgName, TestServiceName, add_update_service.ServiceConfigArgName, updateServiceConfig)
}

func (t *updateServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *updateServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' with UUID '%s' updated", TestServiceName, TestServiceUuid)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
