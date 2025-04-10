package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v5"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/stretchr/testify/require"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           nil,
		Config:         nil,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                         &dockerType,
		Config:                       nil,
		LogsAggregator:               nil,
		GraflokiConfig:               nil,
		ShouldTurnOffDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &kubernetesType,
		Config:         nil,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &clusterType,
		Config:         nil,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesPartialConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesPartialConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           nil,
		EnclaveSizeInMegabytes: nil,
		EngineNodeName:         nil,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &kubernetesType,
		Config:         &kubernetesPartialConfig,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &kubernetesType,
		Config:         &kubernetesFullConfig,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigLogsAggregatorNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &kubernetesType,
		Config:         &kubernetesFullConfig,
		LogsAggregator: nil,
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigLogsAggregatorReservedSinkId(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v5.LogsAggregatorConfigV5{
			Sinks: map[string]map[string]interface{}{
				logs_aggregator.DefaultSinkId: {
					"type": "elasticsearch",
				},
			},
		},
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigLogsAggregatorFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v5.LogsAggregatorConfigV5{
			Sinks: map[string]map[string]interface{}{
				"elasticsearch": {
					"type": "elasticsearch",
				},
			},
		},
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigGraflokiNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v5.LogsAggregatorConfigV5{
			Sinks: map[string]map[string]interface{}{
				"elasticsearch": {
					"type": "elasticsearch",
				},
			},
		},
		GraflokiConfig: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigGraflokiFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:           &kubernetesType,
		Config:         &kubernetesFullConfig,
		LogsAggregator: nil,
		GraflokiConfig: &v5.GraflokiConfig{
			ShouldStartBeforeEngine: false,
			GrafanaImage:            "grafana:1.32",
			LokiImage:               "loki:1.32",
		},
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}
