package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{Type: &dockerType}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{Type: &kubernetesType}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{
			Type: &clusterType,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesPartialConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesPartialConfig := v2.KubernetesClusterConfigV2{
		KubernetesClusterName: &kubernetesClusterName,
	}
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{
		Type: &kubernetesType,
		Config: &kubernetesPartialConfig,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesFullConfig := v2.KubernetesClusterConfigV2{
		KubernetesClusterName: &kubernetesClusterName,
		StorageClass: &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
	}
	kurtosisClusterConfigOverrides := v2.KurtosisClusterConfigV2{
		Type: &kubernetesType,
		Config: &kubernetesFullConfig,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}