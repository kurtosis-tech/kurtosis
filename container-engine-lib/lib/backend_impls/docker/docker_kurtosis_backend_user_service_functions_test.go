package docker

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUserServiceStatusDeterminerCompleteness(t *testing.T) {
	for _, containerStatus := range types.ContainerStatusValues() {
		_, found := userServiceStatusDeterminer[containerStatus]
		require.True(t, found, "No user service status designation set for container status '%v'", containerStatus.String())
	}
}
