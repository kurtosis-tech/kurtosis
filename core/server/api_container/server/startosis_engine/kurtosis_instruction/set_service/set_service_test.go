package set_service

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
)

func TestUpsertServiceConfigsCopiesPrivilegedFields(t *testing.T) {
	currServiceConfig := newEmptyServiceConfigForTest(t)
	serviceConfigOverride := newEmptyServiceConfigForTest(t)
	expectedBindMounts := map[string]string{"/var/run/docker.sock": "/var/run/docker.sock"}
	serviceConfigOverride.SetPrivileged(true)
	serviceConfigOverride.SetBindMounts(expectedBindMounts)
	serviceConfigOverride.SetHostPIDNamespace(true)

	updatedServiceConfig, err := upsertServiceConfigs(currServiceConfig, serviceConfigOverride)

	require.NoError(t, err)
	require.True(t, updatedServiceConfig.GetPrivileged())
	require.Equal(t, expectedBindMounts, updatedServiceConfig.GetBindMounts())
	require.True(t, updatedServiceConfig.GetHostPIDNamespace())
}

func newEmptyServiceConfigForTest(t *testing.T) *service.ServiceConfig {
	serviceConfig, err := service.CreateServiceConfig(
		"",
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{},
		[]string{},
		map[string]string{},
		nil,
		nil,
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		nil,
		nil,
		image_download_mode.ImageDownloadMode_Always,
		false,
		false,
		nil,
		false,
		service.NewGpuConfig(0, nil, 0, nil, "", ""),
	)
	require.NoError(t, err)
	return serviceConfig
}
