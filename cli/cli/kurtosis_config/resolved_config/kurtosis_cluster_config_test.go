package resolved_config

import (
	"testing"

	v6 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v6"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/stretchr/testify/require"
)

func TestNewKurtosisClusterConfigEmptyOverrides(t *testing.T) {
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        nil,
		Config:                      nil,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigDockerType(t *testing.T) {
	dockerType := KurtosisClusterType_Docker.String()
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &dockerType,
		Config:                      nil,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigKubernetesNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      nil,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigNonsenseType(t *testing.T) {
	clusterType := "gdsfgsdfvsf"
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &clusterType,
		Config:                      nil,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.Error(t, err)
}

func TestNewKurtosisClusterConfigKubernetesPartialConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesPartialConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           nil,
		EnclaveSizeInMegabytes: nil,
		EngineNodeName:         nil,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      &kubernetesPartialConfig,
		LogsAggregator:              nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v6.LogsAggregatorConfigV6{
			Sinks: map[string]map[string]interface{}{
				logs_aggregator.DefaultSinkId: {
					"type": "elasticsearch",
				},
			},
		},
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v6.LogsAggregatorConfigV6{
			Sinks: map[string]map[string]interface{}{
				"elasticsearch": {
					"type": "elasticsearch",
				},
			},
		},
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v6.LogsAggregatorConfigV6{
			Sinks: map[string]map[string]interface{}{
				"elasticsearch": {
					"type": "elasticsearch",
				},
			},
		},
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:           &kubernetesType,
		Config:         &kubernetesFullConfig,
		LogsAggregator: nil,
		GrafanaLokiConfig: &v6.GrafanaLokiConfigV6{
			ShouldStartBeforeEngine: false,
			GrafanaImage:            grafanaImage,
			LokiImage:               lokiImage,
		},
		ShouldEnableDefaultLogsSink: nil,
		LogsCollector:               nil,
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
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: &ShouldEnableDefaultLogsSink,
	}
	_, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
}

func TestNewKurtosisClusterConfigLogsCollectorNoConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:                        &kubernetesType,
		Config:                      &kubernetesFullConfig,
		LogsAggregator:              nil,
		LogsCollector:               nil,
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	actualKurtosisClusterConfig, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
	require.Nil(t, actualKurtosisClusterConfig.logsCollector.Filters)
	require.Nil(t, actualKurtosisClusterConfig.logsCollector.Parsers)
}

func TestNewKurtosisClusterConfigLogsCollectorFullConfig(t *testing.T) {
	kubernetesType := KurtosisClusterType_Kubernetes.String()
	kubernetesClusterName := "some-name"
	kubernetesStorageClass := "some-storage-class"
	kubernetesEnclaveSizeInMB := uint(5)
	kubernetesEngineNodeName := "some-node-name"
	kubernetesFullConfig := v6.KubernetesClusterConfigV6{
		KubernetesClusterName:  &kubernetesClusterName,
		StorageClass:           &kubernetesStorageClass,
		EnclaveSizeInMegabytes: &kubernetesEnclaveSizeInMB,
		EngineNodeName:         &kubernetesEngineNodeName,
	}
	kurtosisClusterConfigOverrides := v6.KurtosisClusterConfigV6{
		Type:   &kubernetesType,
		Config: &kubernetesFullConfig,
		LogsAggregator: &v6.LogsAggregatorConfigV6{
			Sinks: map[string]map[string]interface{}{
				"elasticsearch": {
					"type": "elasticsearch",
				},
			},
		},
		LogsCollector: &v6.LogsCollectorConfigV6{
			Filters: []logs_collector.Filter{
				{
					Name:  "grep",
					Match: "*",
					Params: []logs_collector.FilterParam{
						{Key: "exclude", Value: "*"},
						{Key: "logical_op", Value: "&"},
					},
				},
				{
					Name:  "lua",
					Match: "*",
					Params: []logs_collector.FilterParam{
						{Key: "script", Value: "print smth"},
						{Key: "call", Value: "frontend"},
					},
				},
			},
			Parsers: []logs_collector.Parser{
				{
					"name":        "json",
					"format":      "json",
					"time_key":    "time",
					"time_format": "%Y-%m-%dT%H:%M:%S.%L",
				},
				{
					"name":   "regex",
					"format": "regex",
					"regex":  "^\\[(?<time>[^\\]]*)\\] (?<level>\\w+) (?<message>.*)$",
				},
			},
		},
		GrafanaLokiConfig:           nil,
		ShouldEnableDefaultLogsSink: nil,
	}
	actualKurtosisClusterConfig, err := NewKurtosisClusterConfigFromOverrides("test", &kurtosisClusterConfigOverrides)
	require.NoError(t, err)
	require.NotNil(t, actualKurtosisClusterConfig.logsCollector)
	require.NotNil(t, actualKurtosisClusterConfig.logsCollector.Filters)
	require.NotNil(t, actualKurtosisClusterConfig.logsCollector.Parsers)
	require.Equal(t, 2, len(actualKurtosisClusterConfig.logsCollector.Filters))
	require.Equal(t, 2, len(actualKurtosisClusterConfig.logsCollector.Parsers))
	require.Equal(t, "json", actualKurtosisClusterConfig.logsCollector.Parsers[0]["name"])
	require.Equal(t, "regex", actualKurtosisClusterConfig.logsCollector.Parsers[1]["name"])
	require.Equal(t, "grep", actualKurtosisClusterConfig.logsCollector.Filters[0].Name)
	require.Equal(t, "lua", actualKurtosisClusterConfig.logsCollector.Filters[1].Name)
}
