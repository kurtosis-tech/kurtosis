package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"net/http"
)

type ServiceNetwork interface {
	Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types.PartitionID]map[service.ServiceName]bool,
		newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
		newDefaultConnection partition_topology.PartitionConnection,
	) error

	SetConnection(
		ctx context.Context,
		partition1 service_network_types.PartitionID,
		partition2 service_network_types.PartitionID,
		connection partition_topology.PartitionConnection,
	) error

	UnsetConnection(
		ctx context.Context,
		partition1 service_network_types.PartitionID,
		partition2 service_network_types.PartitionID,
	) error

	SetDefaultConnection(
		ctx context.Context,
		connection partition_topology.PartitionConnection,
	) error

	StartService(
		ctx context.Context,
		serviceName service.ServiceName,
		serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig,
	) (
		*service.Service,
		error,
	)

	StartServices(
		ctx context.Context,
		serviceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig,
		batchSize int,
	) (
		map[service.ServiceName]*service.Service,
		map[service.ServiceName]error,
		error,
	)

	UpdateService(
		ctx context.Context,
		updateServiceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig,
	) (
		map[service.ServiceName]bool,
		map[service.ServiceName]error,
		error,
	)

	RemoveService(ctx context.Context, serviceIdentifier string) (service.ServiceUUID, error)

	PauseService(ctx context.Context, serviceIdentifier string) error

	UnpauseService(ctx context.Context, serviceIdentifier string) error

	ExecCommand(ctx context.Context, serviceIdentifier string, command []string) (int32, string, error)

	HttpRequestService(ctx context.Context, serviceIdentifier string, portId string, method string, contentType string, endpoint string, body string) (*http.Response, error)

	GetService(ctx context.Context, serviceIdentifier string) (*service.Service, error)

	CopyFilesFromService(ctx context.Context, serviceIdentifier string, srcPath string, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	GetServiceNames() map[service.ServiceName]bool

	GetServiceNameToPrivatePortIdsMap() (map[service.ServiceName][]string, error)

	GetExistingAndHistoricalServiceIdentifiers() []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers

	GetServiceRegistration(serviceName service.ServiceName) (*service.ServiceRegistration, bool)

	RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	UploadFilesArtifact(data []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	IsNetworkPartitioningEnabled() bool

	GetUniqueNameForFileArtifact() (string, error)
}
