package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/update_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	updateService_enclaveUuid = enclave.EnclaveUUID("enclave-uuid")
	updateService_serviceName = service.ServiceName("test-service")
	updateService_serviceUuid = service.ServiceUUID("test-service-uuid")
	updateService_subnetwork  = "test-subnetwork"
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
	return update_service.UpdateServiceBuiltinName
}

func (t *updateServiceTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().GetService(
		mock.Anything,
		string(updateService_serviceName),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(updateService_serviceName, updateService_serviceUuid, updateService_enclaveUuid, nil, string(updateService_serviceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		nil,
	)

	serviceNetwork.EXPECT().UpdateService(
		mock.Anything,
		map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig{
			updateService_serviceName: binding_constructors.NewUpdateServiceConfig(updateService_subnetwork),
		},
	).Times(1).Return(
		map[service.ServiceName]bool{
			updateService_serviceName: true,
		},
		map[service.ServiceName]error{},
		nil,
	)

	return update_service.NewUpdateService(serviceNetwork)
}

func (t *updateServiceTestCase) GetStarlarkCode() string {
	updateServiceConfig := fmt.Sprintf("UpdateServiceConfig(subnetwork=%q)", updateService_subnetwork)
	return fmt.Sprintf("%s(%s=%q, %s=%s)", update_service.UpdateServiceBuiltinName, update_service.ServiceNameArgName, updateService_serviceName, update_service.UpdateServiceConfigArgName, updateServiceConfig)
}

func (t *updateServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *updateServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Service '%s' with UUID '%s' updated", updateService_serviceName, updateService_serviceUuid)
	require.Regexp(t, expectedExecutionResult, *executionResult)
}
