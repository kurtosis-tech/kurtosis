package resolved_config

import (
	v1 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := "docker"
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{Type: &dockerType}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := "kubernetes"
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{Type: &kubernetesType}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "nonsense"
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{
			Type: &clusterType,
	}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}
