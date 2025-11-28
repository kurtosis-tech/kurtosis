package container_status_calculator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsContainerRunningCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, err := IsContainerRunning(containerStatus)
		require.Nil(t, err, "No branch in IsContainerRunning for container status %v", containerStatus.String())
	}
}
