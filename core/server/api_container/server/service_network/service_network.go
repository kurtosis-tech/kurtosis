package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
)

type ServiceNetwork interface {
	Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
		newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
		newDefaultConnection partition_topology.PartitionConnection,
	) error

	StartServices(
		ctx context.Context,
		serviceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
		partitionID service_network_types.PartitionID,
	) (
		map[service.ServiceID]*service.Service,
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

	GetService(ctx context.Context, serviceId service.ServiceID) (
		*service.Service,
		error,
	)

	CopyFilesFromService(
		ctx context.Context,
		serviceId service.ServiceID,
		srcPath string,
	) (
		enclave_data_directory.FilesArtifactUUID,
		error,
	)
}
