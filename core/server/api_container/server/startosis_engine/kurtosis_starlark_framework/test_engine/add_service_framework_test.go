package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
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

	serviceNetwork.EXPECT().StartService(
		mock.Anything,
		TestServiceName,
		mock.MatchedBy(func(serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) bool {
			expectedServiceConfig := services.NewServiceConfigBuilder(
				TestContainerImageName,
			).WithPrivatePorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
				TestPrivatePortId: binding_constructors.NewPort(TestPrivatePortNumber, TestPrivatePortProtocol, TestPublicApplicationProtocol),
			}).WithPublicPorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
				TestPublicPortId: binding_constructors.NewPort(TestPublicPortNumber, TestPublicPortProtocol, TestPublicApplicationProtocol),
			}).WithFilesArtifactMountDirpaths(map[string]string{
				TestFilesArtifactPath1: TestFilesArtifactName1,
			}).WithCmdArgs(
				TestCmdSlice,
			).WithEntryPointArgs(
				TestEntryPointSlice,
			).WithEnvVars(map[string]string{
				TestEnvVarName1: TestEnvVarValue1,
			}).WithPrivateIPAddressPlaceholder(
				TestPrivateIPAddressPlaceholder,
			).WithSubnetwork(
				string(TestSubnetwork),
			).WithCpuAllocationMillicpus(
				TestCpuAllocation,
			).WithMemoryAllocationMegabytes(
				TestMemoryAllocation,
			).Build()
			actualServiceConfig := services.NewServiceConfigBuilderFromServiceConfig(serviceConfig).Build()
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
	serviceConfigStarlarkStrTemplate := "ServiceConfig(" +
		"image=%q, " +
		"ports={%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}, " +
		"public_ports={%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)}, " +
		"files={%q: %q}, " +
		"entrypoint=[%q, %q], " +
		"cmd=[%q, %q, %q], " +
		"env_vars={%q: %q}, " +
		"private_ip_address_placeholder=%q, " +
		"subnetwork=%q, " +
		"cpu_allocation=%d, " +
		"memory_allocation=%d)"
	serviceConfig := fmt.Sprintf(serviceConfigStarlarkStrTemplate,
		TestContainerImageName,
		TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol,
		TestPublicPortId, TestPublicPortNumber, TestPublicPortProtocolStr, TestPublicApplicationProtocol,
		TestFilesArtifactPath1, TestFilesArtifactName1,
		TestEntryPointSlice[0], TestEntryPointSlice[1],
		TestCmdSlice[0], TestCmdSlice[1], TestCmdSlice[2],
		TestEnvVarName1, TestEnvVarValue1,
		TestPrivateIPAddressPlaceholder,
		TestSubnetwork,
		TestCpuAllocation,
		TestMemoryAllocation)
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, add_service.AddServiceBuiltinName, add_service.ServiceNameArgName, TestServiceName, add_service.ServiceConfigArgName, serviceConfig)
}

func (t *addServiceTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *addServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	serviceObj, ok := interpretationResult.(*kurtosis_types.Service)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.NotNil(t, serviceObj)
	expectedServiceObj := fmt.Sprintf(`Service(hostname = "{{kurtosis:%s.hostname}}", ip_address = "{{kurtosis:%s.ip_address}}", ports = {%q: PortSpec(number=%d, transport_protocol=%q, application_protocol=%q)})`, TestServiceName, TestServiceName, TestPrivatePortId, TestPrivatePortNumber, TestPrivatePortProtocolStr, TestPrivateApplicationProtocol)
	require.Equal(t, expectedServiceObj, serviceObj.String())

	expectedExecutionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", TestServiceName, TestServiceUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
