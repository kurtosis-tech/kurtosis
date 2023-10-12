package resolved_config

import (
	v3 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   nil,
		Config: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   &dockerType,
		Config: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   &kubernetesType,
		Config: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   &clusterType,
		Config: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesPartialConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesPartialConfig := v3.KubernetesClusterConfigV3{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           nil,
		EnclaveSizeInMegabytes: nil,
	}
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   &kubernetesType,
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
	kubernetesFullConfig := v3.KubernetesClusterConfigV3{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
	}
	kurtosisClusterConfigOverrides := v3.KurtosisClusterConfigV3{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}
