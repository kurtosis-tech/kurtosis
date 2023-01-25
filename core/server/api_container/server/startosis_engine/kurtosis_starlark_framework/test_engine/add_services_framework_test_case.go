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
	"testing"
)

const (
	enclaveUuid  = "test-enclave"
	service1     = service.ServiceName("service-1")
	serviceUuid1 = service.ServiceUUID("service-1-uuid")
	service2     = service.ServiceName("service-2")
	serviceUuid2 = service.ServiceUUID("service-2-uuid")
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
			require.Contains(t, configs, service1)
			require.Contains(t, configs, service2)

			expectedServiceConfig1 := services.NewServiceConfigBuilder("test-image-1").WithSubnetwork("my-subnetwork").Build()
			actualServiceConfig1 := services.NewServiceConfigBuilderFromServiceConfig(configs[service1]).Build()
			assert.Equal(t, expectedServiceConfig1, actualServiceConfig1)
			actualServiceConfig2 := services.NewServiceConfigBuilderFromServiceConfig(configs[service2]).Build()
			expectedServiceConfig2 := services.NewServiceConfigBuilder("test-image-2").WithCpuAllocationMillicpus(1000).WithMemoryAllocationMegabytes(2048).Build()
			assert.Equal(t, expectedServiceConfig2, actualServiceConfig2)
			return true
		}),
	).Times(1).Return(
		map[service.ServiceName]*service.Service{
			service1: service.NewService(service.NewServiceRegistration(service1, serviceUuid1, enclaveUuid, nil), container_status.ContainerStatus_Running, nil, nil, nil),
			service2: service.NewService(service.NewServiceRegistration(service2, serviceUuid2, enclaveUuid, nil), container_status.ContainerStatus_Running, nil, nil, nil),
		},
		map[service.ServiceName]error{},
		nil,
	)

	return add_service.NewAddServices(serviceNetwork, runtimeValueStore)
}

func (t *addServicesTestCase) GetStarlarkCode() string {
	serviceConfig1 := `ServiceConfig(image="test-image-1", subnetwork="my-subnetwork")`
	serviceConfig2 := `ServiceConfig(image="test-image-2", cpu_allocation=1000, memory_allocation=2048)`
	return fmt.Sprintf(`%s(%s={"service-1": %s, "service-2": %s})`, add_service.AddServicesBuiltinName, add_service.ConfigsArgName, serviceConfig1, serviceConfig2)
}

func (t *addServicesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	resultDict, ok := interpretationResult.(*starlark.Dict)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.Equal(t, resultDict.Len(), 2)
	require.Contains(t, resultDict.Keys(), starlark.String(service1))
	require.Contains(t, resultDict.Keys(), starlark.String(service2))

	require.Contains(t, *executionResult, "Successfully added the following '2' services:")
	require.Contains(t, *executionResult, "Service 'service-1' added with UUID 'service-1-uuid'")
	require.Contains(t, *executionResult, "Service 'service-2' added with UUID 'service-2-uuid'")
}
