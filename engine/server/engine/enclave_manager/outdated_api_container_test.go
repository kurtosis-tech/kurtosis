package enclave_manager

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	"github.com/stretchr/testify/require"
)

func newAPIContainerWithImage(imageName string) *api_container.APIContainer {
	return api_container.NewAPIContainer("enclave-uuid", container.ContainerStatus_Running, nil, nil, nil, nil, nil, false, imageName)
}

func TestIsOutdatedAPIContainer(t *testing.T) {
	require.True(t, isOutdatedAPIContainer(newAPIContainerWithImage("kurtosistech/core:0.0.1")))
	require.False(t, isOutdatedAPIContainer(newAPIContainerWithImage(api_container_launcher.GetDefaultContainerImageAndTag())))
	require.False(t, isOutdatedAPIContainer(newAPIContainerWithImage("")))
}
