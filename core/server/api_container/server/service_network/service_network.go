package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"net"
	"net/http"
)

type ServiceNetwork interface {
	Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
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

	StartServices(
		ctx context.Context,
		serviceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	) (
		map[service.ServiceID]*service.Service,
		map[service.ServiceID]error,
		error,
	)

	UpdateService(
		ctx context.Context,
		updateServiceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.UpdateServiceConfig,
	) (
		map[service.ServiceID]bool,
		map[service.ServiceID]error,
		error,
	)

	RemoveService(
		ctx context.Context,
		serviceId service.ServiceID,
	) (service.ServiceGUID, error)

	PauseService(
		ctx context.Context,
		serviceId service.ServiceID,
	) error

	UnpauseService(
		ctx context.Context,
		serviceId service.ServiceID,
	) error

	ExecCommand(
		ctx context.Context,
		serviceId service.ServiceID,
		command []string,
	) (int32, string, error)

	HttpRequestService(
		ctx context.Context,
		serviceId service.ServiceID,
		portId string,
		method string,
		contentType string,
		endpoint string,
		body string,
	) (*http.Response, error)

	GetService(ctx context.Context, serviceId service.ServiceID) (
		*service.Service,
		error,
	)

	CopyFilesFromService(
		ctx context.Context,
		serviceId service.ServiceID,
		srcPath string,
		artifactName string,
	) (
		enclave_data_directory.FilesArtifactUUID,
		error,
	)

	GetServiceIDs() map[service.ServiceID]bool

	GetIPAddressForService(serviceID service.ServiceID) (net.IP, bool)

	RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	UploadFilesArtifact(data []byte, artifactName string) (enclave_data_directory.FilesArtifactUUID, error)

	IsNetworkPartitioningEnabled() bool
}
