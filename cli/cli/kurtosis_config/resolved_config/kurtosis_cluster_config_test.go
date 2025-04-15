package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v5"
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/stretchr/testify/require"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                        nil,
		Config:                      nil,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                        &dockerType,
		Config:                      nil,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                        &kubernetesType,
		Config:                      nil,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                        &clusterType,
		Config:                      nil,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		Type:                        &kubernetesType,
		Config:                      &kubernetesPartialConfig,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
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
	grafanaImage := "grafana:1.32"
	lokiImage := "loki:1.32"
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
		GrafanaLokiConfig: &v5.GrafanaLokiConfig{
			ShouldStartBeforeEngine: false,
			GrafanaImage:            grafanaImage,
			LokiImage:               lokiImage,
		},
		ShouldEnableDefaultLogsSink: nil,
	}
	actualKurtosisClusterConfig, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NotNil(t, actualKurtosisClusterConfig.graflokiConfig)
	require.Equal(t, actualKurtosisClusterConfig.graflokiConfig.GrafanaImage, grafanaImage)
	require.Equal(t, actualKurtosisClusterConfig.graflokiConfig.LokiImage, lokiImage)
	require.False(t, actualKurtosisClusterConfig.graflokiConfig.ShouldStartBeforeEngine)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigShouldEnableDefaultLogsSink(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	ShouldEnableDefaultLogsSink := true
	kubernetesFullConfig := v5.KubernetesClusterConfigV5{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v5.KurtosisClusterConfigV5{
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: &ShouldEnableDefaultLogsSink,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}
