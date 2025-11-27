package enclave_manager

import (
	"testing"

	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/stretchr/testify/require"
)

func TestIsContainerRunningDeterminerCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, found := isContainerRunningDeterminer[containerStatus]
		require.True(t, found, "No is-container-running determination provided for container status '%v'", containerStatus.String())
	}
}

func TestGetEnclaveStatusFromEnclaveStatusCompleteness(t *testing.T) {
	for _, enclaveStatus := range enclave.EnclaveStatusValues() {
		_, err := getEnclaveStatusFromEnclaveStatus(enclaveStatus)
		require.NoError(t, err, "No EnclaveStatus provided for enclave status '%v'", enclaveStatus.String())
	}
}

func TestGetApiContainerStatusFromContainerStatusCompleteness(t *testing.T) {
	for _, containerStatus := range container.ContainerStatusValues() {
		_, err := getApiContainerStatusFromContainerStatus(containerStatus)
		require.NoError(t, err, "No ApiContainerStatus provided for container status '%v'", containerStatus.String())
	}
}
