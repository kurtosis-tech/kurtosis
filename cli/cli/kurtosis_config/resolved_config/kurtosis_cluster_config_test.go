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
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{Type: &dockerType}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{Type: &kubernetesType}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{
			Type: &clusterType,
	}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesPartialConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesPartialConfig := v1.KubernetesClusterConfigV1{
		KubernetesClusterName: &kubernetesClusterName,
	}
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{
		Type: &kubernetesType,
		Config: &kubernetesPartialConfig,
	}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesFullConfig := v1.KubernetesClusterConfigV1{
		KubernetesClusterName: &kubernetesClusterName,
		StorageClass: &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
	}
	kurtosisClusterConfigOverrides := v1.KurtosisClusterConfigV1{
		Type: &kubernetesType,
		Config: &kubernetesFullConfig,
	}
	_, err := NewKurtosisClusterConfigFromOverrides(&kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}