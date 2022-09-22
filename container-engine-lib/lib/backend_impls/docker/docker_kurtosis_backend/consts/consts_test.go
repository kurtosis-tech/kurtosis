package consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsContainerRunningDeterminerCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, found := IsContainerRunningDeterminer[containerStatus]
		require.True(t, found, "No is-running designation set for container status '%v'", containerStatus.String())
	}
}
