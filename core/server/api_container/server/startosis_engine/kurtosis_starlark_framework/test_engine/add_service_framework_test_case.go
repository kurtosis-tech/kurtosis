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

const (
	addService_enclaveUuid = "test-enclave"
	addService_serviceName = service.ServiceName("service-1")
	addService_serviceUuid = service.ServiceUUID("service-1-uuid")
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
		service1,
		mock.MatchedBy(func(serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) bool {
			expectedServiceConfig := services.NewServiceConfigBuilder(
				"kurtosistech/example-datastore-server",
			).WithPrivatePorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": binding_constructors.NewPort(1234, kurtosis_core_rpc_api_bindings.Port_TCP, "http"),
			}).WithPublicPorts(map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": binding_constructors.NewPort(80, kurtosis_core_rpc_api_bindings.Port_TCP, "http"),
			}).WithFilesArtifactMountDirpaths(map[string]string{
				"path/to/file/1": "file_1",
				"path/to/file/2": "file_2",
			}).WithCmdArgs(
				[]string{"bash", "-c", "/apps/main.py", "1234"},
			).WithEntryPointArgs(
				[]string{"127.0.0.0", "1234"},
			).WithEnvVars(map[string]string{
				"VAR_1": "VALUE_1",
				"VAR_2": "VALUE_2",
			}).WithPrivateIPAddressPlaceholder(
				"<IP_ADDRESS>",
			).WithSubnetwork(
				"subnetwork_1",
			).WithCpuAllocationMillicpus(
				2000,
			).WithMemoryAllocationMegabytes(
				1024,
			).Build()
			actualServiceConfig := services.NewServiceConfigBuilderFromServiceConfig(serviceConfig).Build()
			assert.Equal(t, expectedServiceConfig, actualServiceConfig)
			return true
		}),
	).Times(1).Return(
		service.NewService(service.NewServiceRegistration(addService_serviceName, addService_serviceUuid, addService_enclaveUuid, nil, string(addService_serviceName)), container_status.ContainerStatus_Running, nil, nil, nil),
		nil,
	)

	return add_service.NewAddService(serviceNetwork, runtimeValueStore)
}

func (t *addServiceTestCase) GetStarlarkCode() string {
	serviceConfig := `ServiceConfig(image="kurtosistech/example-datastore-server", ports={"grpc": PortSpec(number=1234, transport_protocol="TCP", application_protocol="http")}, public_ports={"grpc": PortSpec(number=80, transport_protocol="TCP", application_protocol="http")}, files={"path/to/file/1": "file_1", "path/to/file/2": "file_2"}, entrypoint=["127.0.0.0", "1234"], cmd=["bash", "-c", "/apps/main.py", "1234"], env_vars={"VAR_1": "VALUE_1", "VAR_2": "VALUE_2"}, private_ip_address_placeholder="<IP_ADDRESS>", subnetwork="subnetwork_1", cpu_allocation=2000, memory_allocation=1024)`
	return fmt.Sprintf(`%s(%s=%q, %s=%s)`, add_service.AddServiceBuiltinName, add_service.ServiceNameArgName, addService_serviceName, add_service.ServiceConfigArgName, serviceConfig)
}

func (t *addServiceTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	serviceObj, ok := interpretationResult.(*kurtosis_types.Service)
	require.True(t, ok, "interpretation result should be a dictionary")
	require.NotNil(t, serviceObj)
	expectedServiceObj := `Service(hostname = "{{kurtosis:service-1.hostname}}", ip_address = "{{kurtosis:service-1.ip_address}}", ports = {"grpc": PortSpec(number=1234, transport_protocol="TCP", application_protocol="http")})`
	require.Equal(t, expectedServiceObj, serviceObj.String())

	require.Equal(t, *executionResult, "Service 'service-1' added with service UUID 'service-1-uuid'")
}
