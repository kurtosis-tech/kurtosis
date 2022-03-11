package docker

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/partition"
)

func (backend *DockerKurtosisBackend) CreateEnclave(
	ctx context.Context,
	enclaveId string,
) (
	*enclave.Enclave,
	error,
) {
	panic("Implement me")
}

// Gets enclaves matching the given filters
func (backend *DockerKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[string]*enclave.Enclave,
	error,
) {
	panic("Implement me")
}

// TODO MAYYYYYYYBE DumpEnclaves?

// Stops enclaves matching the given filters
func (backend *DockerKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[string]bool,
	erroredEnclaveIds map[string]error,
	resultErr error,
) {
	panic("Implement me")
}

// Destroys enclaves matching the given filters
func (backend *DockerKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[string]bool,
	erroredEnclaveIds map[string]error,
	resultErr error,
) {
	panic("Implement me")
}

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
