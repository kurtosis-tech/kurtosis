package privileged_mode

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/stretchr/testify/require"
)

func TestValidateServiceConfigAllowsPlainServiceConfig(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()

	err := ValidateServiceConfig(serviceConfig, false, args.KurtosisBackendType_Kubernetes)

	require.Nil(t, err)
}

func TestValidateServiceConfigRejectsPrivilegedWithoutOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetPrivileged(true)

	err := ValidateServiceConfig(serviceConfig, false, args.KurtosisBackendType_Docker)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "did not opt in")
}

func TestValidateServiceConfigRejectsBindMountsWithoutOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetBindMounts(map[string]string{"/var/run/docker.sock": "/var/run/docker.sock"})

	err := ValidateServiceConfig(serviceConfig, false, args.KurtosisBackendType_Docker)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "did not opt in")
}

func TestValidateServiceConfigRejectsHostPIDNamespaceWithoutOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetHostPIDNamespace(true)

	err := ValidateServiceConfig(serviceConfig, false, args.KurtosisBackendType_Docker)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "did not opt in")
}

func TestValidateServiceConfigAllowsPrivilegedOnDockerWithOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetPrivileged(true)

	err := ValidateServiceConfig(serviceConfig, true, args.KurtosisBackendType_Docker)

	require.Nil(t, err)
}

func TestValidateServiceConfigAllowsHostPIDNamespaceOnDockerWithOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetHostPIDNamespace(true)

	err := ValidateServiceConfig(serviceConfig, true, args.KurtosisBackendType_Docker)

	require.Nil(t, err)
}

func TestValidateServiceConfigRejectsHostCgroupNamespaceWithoutOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetHostCgroupNamespace(true)

	err := ValidateServiceConfig(serviceConfig, false, args.KurtosisBackendType_Docker)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "did not opt in")
}

func TestValidateServiceConfigAllowsHostCgroupNamespaceOnDockerWithOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetHostCgroupNamespace(true)

	err := ValidateServiceConfig(serviceConfig, true, args.KurtosisBackendType_Docker)

	require.Nil(t, err)
}

func TestValidateServiceConfigRejectsPrivilegedOnKubernetesEvenWithOptIn(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()
	serviceConfig.SetPrivileged(true)

	err := ValidateServiceConfig(serviceConfig, true, args.KurtosisBackendType_Kubernetes)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Docker-only")
}
