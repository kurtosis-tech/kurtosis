package docker_kurtosis_backend

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db/free_ip_addr_tracker"
	"github.com/stretchr/testify/require"
)

func TestDockerKurtosisBackend(t *testing.T) {
	dockerManager, err := docker_manager.CreateDockerManager([]client.Opt{})
	require.NoError(t, err)
	enclaveFreeIpAddrTrackers := map[enclave.EnclaveUUID]*free_ip_addr_tracker.FreeIpAddrTracker{}
	productionMode := false

	backend := NewDockerKurtosisBackend(dockerManager, enclaveFreeIpAddrTrackers, nil, productionMode)

	err = backend.SnapshotEnclave(context.Background(), enclave.EnclaveUUID("tedi"), "./")
	require.NoError(t, err)
}
