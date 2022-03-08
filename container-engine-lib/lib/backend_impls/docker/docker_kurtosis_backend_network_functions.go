package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/partition"
)

func (backendCore *DockerKurtosisBackend) CreateRepartition(
	ctx context.Context,
	partitions []*partition.Partition,
	newPartitionConnections map[partition.PartitionConnectionID]partition.PartitionConnection,
	newDefaultConnection partition.PartitionConnection,
)(
	resultErr error,
) {
	panic("Implement me")
}
