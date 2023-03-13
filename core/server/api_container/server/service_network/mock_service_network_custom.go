package service_network

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"net"
	"net/http"
)

const (
	mockEnclaveUuid   = "enclave-uuid"
	serviceUuidSuffix = "uuid"

	mockFileArtifactName = "mock-artifact-id"
	unimplementedMsg     = "Method is unimplemented!!!"
)

// MockServiceNetworkCustom is a manual mock for ServiceNetwork interface
// TODO: migrate to use the mockery-generated mock MockServiceNetwork
type MockServiceNetworkCustom struct {
	serviceRegistrations map[service.ServiceName]*service.ServiceRegistration
}

func NewMockServiceNetworkCustom(ipAddresses map[service.ServiceName]net.IP) *MockServiceNetworkCustom {
	serviceRegistrations := map[service.ServiceName]*service.ServiceRegistration{}
	for serviceName, ipAddress := range ipAddresses {
		serviceRegistrations[serviceName] = generateMockServiceRegistration(serviceName, ipAddress)
	}
	return &MockServiceNetworkCustom{
		serviceRegistrations: serviceRegistrations,
	}
}

func NewEmptyMockServiceNetwork() *MockServiceNetworkCustom {
	return &MockServiceNetworkCustom{
		serviceRegistrations: nil,
	}
}

func (m *MockServiceNetworkCustom) Repartition(ctx context.Context, newPartitionServices map[service_network_types.PartitionID]map[service.ServiceName]bool, newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection, newDefaultConnection partition_topology.PartitionConnection) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) SetConnection(ctx context.Context, partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID, connection partition_topology.PartitionConnection) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) UnsetConnection(ctx context.Context, partition1 service_network_types.PartitionID, partition2 service_network_types.PartitionID) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) SetDefaultConnection(ctx context.Context, connection partition_topology.PartitionConnection) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) StartService(
	ctx context.Context,
	serviceName service.ServiceName,
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig,
	serviceReadinessCheckFunc ServiceReadinessCheckFunc,
) (
	*service.Service,
	error,
) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) StartServices(
	ctx context.Context,
	serviceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	batchSize int,
	serviceReadinessCheckFuncs map[service.ServiceName]ServiceReadinessCheckFunc,
) (
	map[service.ServiceName]*service.Service,
	map[service.ServiceName]error,
	error,
) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) UpdateService(ctx context.Context, updateServiceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig) (map[service.ServiceName]bool, map[service.ServiceName]error, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) RemoveService(ctx context.Context, serviceIdentifier string) (service.ServiceUUID, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) PauseService(ctx context.Context, serviceIdentifier string) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) UnpauseService(ctx context.Context, serviceIdentifier string) error {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) ExecCommand(ctx context.Context, serviceIdentifier string, command []string) (int32, string, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string) (*http.Response, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error) {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) GetServiceNames() map[service.ServiceName]bool {
	//TODO implement me
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) GetServiceRegistration(serviceName service.ServiceName) (*service.ServiceRegistration, bool) {
	serviceRegistration, found := m.serviceRegistrations[serviceName]
	return serviceRegistration, found
}

func (m *MockServiceNetworkCustom) RenderTemplates(_ map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, _ string) (enclave_data_directory.FilesArtifactUUID, error) {
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) UploadFilesArtifact(_ []byte, _ string) (enclave_data_directory.FilesArtifactUUID, error) {
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) IsNetworkPartitioningEnabled() bool {
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) GetExistingAndHistoricalServiceIdentifiers() []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers {
	panic(unimplementedMsg)
}

func (m *MockServiceNetworkCustom) GetUniqueNameForFileArtifact() (string, error) {
	return mockFileArtifactName, nil
}

func generateMockServiceRegistration(serviceName service.ServiceName, ipAddress net.IP) *service.ServiceRegistration {
	return service.NewServiceRegistration(
		serviceName,
		service.ServiceUUID(fmt.Sprintf("%s-%s", serviceName, serviceUuidSuffix)),
		mockEnclaveUuid,
		ipAddress,
		string(serviceName),
	)
}
