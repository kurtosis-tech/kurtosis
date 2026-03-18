package overrides_migrators

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	v0 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v0"
	v1 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v1"
	v2 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v2"
	v3 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	v4 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v4"
	v5 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v5"
	v6 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v6"
	v7 "github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v7"
	"github.com/kurtosis-tech/stacktrace"
)

/*
This file contains functions that will migrate version N of the config overrides to version N+1
*/

// Takes a version of the config, casts it, migrates it to the N+1 version, and returns it
type configOverridesMigrator = func(uncastedOldConfig interface{}) (interface{}, error)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
// Adding a new version here is:
//  1. creating a new migrateFromVX function, where X = the latest-1 config version
//  2. adding an entry for the latest-1 config version with your new function
//
// We keep these sorted in REVERSE chronological order so you don't need to scroll
// to the bottom each time
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
var AllConfigOverridesMigrators = map[config_version.ConfigVersion]configOverridesMigrator{
	config_version.ConfigVersion_v6: migrateFromV6,
	config_version.ConfigVersion_v5: migrateFromV5,
	config_version.ConfigVersion_v4: migrateFromV4,
	config_version.ConfigVersion_v3: migrateFromV3,
	config_version.ConfigVersion_v2: migrateFromV2,
	config_version.ConfigVersion_v1: migrateFromV1,
	config_version.ConfigVersion_v0: migrateFromV0,
}

// vvvvvvvvvvvvvvvvvvvvvvv REVERSE chronological order so you don't have to scroll forever vvvvvvvvvvvvvvvvvvvv
func migrateFromV6(uncastedConfig interface{}) (interface{}, error) {
	castedOldConfig, ok := uncastedConfig.(*v6.KurtosisConfigV6)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	var newClusters map[string]*v7.KurtosisClusterConfigV7
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = make(map[string]*v7.KurtosisClusterConfigV7, len(castedOldConfig.KurtosisClusters))
		for oldClusterName, oldClusterConfig := range castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config
			oldLogsAggregatorConfig := oldClusterConfig.LogsAggregator
			oldLogsCollectorConfig := oldClusterConfig.LogsCollector
			oldGraflokiConfig := oldClusterConfig.GrafanaLokiConfig

			var newKubernetesConfig *v7.KubernetesClusterConfigV7
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v7.KubernetesClusterConfigV7{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
					EngineNodeName:         oldKubernetesConfig.EngineNodeName,
					NodeSelectors:          nil,
					Tolerations:            nil,
				}
			}

			var newLogsAggregatorConfig *v7.LogsAggregatorConfigV7
			if oldLogsAggregatorConfig != nil {
				newLogsAggregatorConfig = &v7.LogsAggregatorConfigV7{
					Sinks: oldLogsAggregatorConfig.Sinks,
				}
			}

			var newLogsCollectorConfig *v7.LogsCollectorConfigV7
			if oldLogsCollectorConfig != nil {
				newLogsCollectorConfig = &v7.LogsCollectorConfigV7{
					Parsers: oldLogsCollectorConfig.Parsers,
					Filters: oldLogsCollectorConfig.Filters,
				}
			}

			var newGraflokiConfig *v7.GrafanaLokiConfigV7
			if oldGraflokiConfig != nil {
				newGraflokiConfig = &v7.GrafanaLokiConfigV7{
					ShouldStartBeforeEngine: oldGraflokiConfig.ShouldStartBeforeEngine,
					GrafanaImage:            oldGraflokiConfig.GrafanaImage,
					LokiImage:               oldGraflokiConfig.LokiImage,
				}
			}

			newClusterConfig := &v7.KurtosisClusterConfigV7{
				Type:                        oldClusterConfig.Type,
				Config:                      newKubernetesConfig,
				LogsAggregator:              newLogsAggregatorConfig,
				LogsCollector:               newLogsCollectorConfig,
				GrafanaLokiConfig:           newGraflokiConfig,
				ShouldEnableDefaultLogsSink: oldClusterConfig.ShouldEnableDefaultLogsSink,
			}

			newClusters[oldClusterName] = newClusterConfig
		}
	}

	var newCloudConfig *v7.KurtosisCloudConfigV7
	if castedOldConfig.CloudConfig != nil {
		newCloudConfig = &v7.KurtosisCloudConfigV7{
			ApiUrl:           castedOldConfig.CloudConfig.ApiUrl,
			Port:             castedOldConfig.CloudConfig.Port,
			CertificateChain: castedOldConfig.CloudConfig.CertificateChain,
		}
	}

	newConfig := &v7.KurtosisConfigV7{
		ConfigVersion:     config_version.ConfigVersion_v7,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       newCloudConfig,
	}

	return newConfig, nil
}

func migrateFromV5(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v5.KurtosisConfigV5)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	var newClusters map[string]*v6.KurtosisClusterConfigV6
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = map[string]*v6.KurtosisClusterConfigV6{}
		for oldClusterName, oldClusterConfig := range castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config
			oldLogsAggregatorConfig := oldClusterConfig.LogsAggregator
			oldGraflokiConfig := oldClusterConfig.GrafanaLokiConfig

			var newKubernetesConfig *v6.KubernetesClusterConfigV6
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v6.KubernetesClusterConfigV6{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
					EngineNodeName:         oldKubernetesConfig.EngineNodeName,
				}
			}

			var newLogsAggregatorConfig *v6.LogsAggregatorConfigV6
			if oldLogsAggregatorConfig != nil {
				newLogsAggregatorConfig = &v6.LogsAggregatorConfigV6{
					Sinks: oldLogsAggregatorConfig.Sinks,
				}
			}

			var newGraflokiConfig *v6.GrafanaLokiConfigV6
			if oldGraflokiConfig != nil {
				newGraflokiConfig = &v6.GrafanaLokiConfigV6{
					ShouldStartBeforeEngine: oldGraflokiConfig.ShouldStartBeforeEngine,
					GrafanaImage:            oldGraflokiConfig.GrafanaImage,
					LokiImage:               oldGraflokiConfig.LokiImage,
				}
			}

			newClusterConfig := &v6.KurtosisClusterConfigV6{
				Type:                        oldClusterConfig.Type,
				Config:                      newKubernetesConfig,
				LogsAggregator:              newLogsAggregatorConfig,
				LogsCollector:               nil, // New field, initialize as nil
				GrafanaLokiConfig:           newGraflokiConfig,
				ShouldEnableDefaultLogsSink: oldClusterConfig.ShouldEnableDefaultLogsSink,
			}

			newClusters[oldClusterName] = newClusterConfig
		}
	}

	var newCloudConfig *v6.KurtosisCloudConfigV6
	if castedOldConfig.CloudConfig != nil {
		newCloudConfig = &v6.KurtosisCloudConfigV6{
			ApiUrl:           castedOldConfig.CloudConfig.ApiUrl,
			Port:             castedOldConfig.CloudConfig.Port,
			CertificateChain: castedOldConfig.CloudConfig.CertificateChain,
		}
	}

	newConfig := &v6.KurtosisConfigV6{
		ConfigVersion:     config_version.ConfigVersion_v6,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       newCloudConfig,
	}

	return newConfig, nil
}

func migrateFromV4(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v4.KurtosisConfigV4)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	var newClusters map[string]*v5.KurtosisClusterConfigV5
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = map[string]*v5.KurtosisClusterConfigV5{}
		for oldClusterName, oldClusterConfig := range castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config
			oldLogsAggregatorConfig := oldClusterConfig.LogsAggregator

			var newKubernetesConfig *v5.KubernetesClusterConfigV5
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v5.KubernetesClusterConfigV5{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
					EngineNodeName:         oldKubernetesConfig.EngineNodeName,
				}
			}

			var newLogsAggregatorConfig *v5.LogsAggregatorConfigV5
			if oldLogsAggregatorConfig != nil {
				newLogsAggregatorConfig = &v5.LogsAggregatorConfigV5{
					Sinks: oldLogsAggregatorConfig.Sinks,
				}
			}

			newGraflokiConfig := &v5.GrafanaLokiConfigV5{
				ShouldStartBeforeEngine: false,
				GrafanaImage:            "",
				LokiImage:               "",
			}

			newClusterConfig := &v5.KurtosisClusterConfigV5{
				Type:                        oldClusterConfig.Type,
				Config:                      newKubernetesConfig,
				LogsAggregator:              newLogsAggregatorConfig,
				GrafanaLokiConfig:           newGraflokiConfig,
				ShouldEnableDefaultLogsSink: nil,
			}

			newClusters[oldClusterName] = newClusterConfig
		}
	}

	var newCloudConfig *v5.KurtosisCloudConfigV5
	if castedOldConfig.CloudConfig != nil {
		newCloudConfig = &v5.KurtosisCloudConfigV5{
			ApiUrl:           castedOldConfig.CloudConfig.ApiUrl,
			Port:             castedOldConfig.CloudConfig.Port,
			CertificateChain: castedOldConfig.CloudConfig.CertificateChain,
		}
	}

	newConfig := &v5.KurtosisConfigV5{
		ConfigVersion:     config_version.ConfigVersion_v4,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       newCloudConfig,
	}

	return newConfig, nil
}

func migrateFromV3(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v3.KurtosisConfigV3)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	var newClusters map[string]*v4.KurtosisClusterConfigV4
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = map[string]*v4.KurtosisClusterConfigV4{}
		for oldClusterName, oldClusterConfig := range castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config
			oldLogsAggregatorConfig := oldClusterConfig.LogsAggregator

			var newKubernetesConfig *v4.KubernetesClusterConfigV4
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v4.KubernetesClusterConfigV4{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
					EngineNodeName:         nil,
				}
			}

			var newLogsAggregator *v4.LogsAggregatorConfigV4
			if oldLogsAggregatorConfig != nil {
				newLogsAggregator = &v4.LogsAggregatorConfigV4{
					Sinks: oldLogsAggregatorConfig.Sinks,
				}
			}

			newClusterConfig := &v4.KurtosisClusterConfigV4{
				Type:           oldClusterConfig.Type,
				Config:         newKubernetesConfig,
				LogsAggregator: newLogsAggregator,
			}
			newClusters[oldClusterName] = newClusterConfig
		}
	}

	var newCloudConfig *v4.KurtosisCloudConfigV4
	if castedOldConfig.CloudConfig != nil {
		newCloudConfig = &v4.KurtosisCloudConfigV4{
			ApiUrl:           castedOldConfig.CloudConfig.ApiUrl,
			Port:             castedOldConfig.CloudConfig.Port,
			CertificateChain: castedOldConfig.CloudConfig.CertificateChain,
		}
	}

	newConfig := &v4.KurtosisConfigV4{
		ConfigVersion:     config_version.ConfigVersion_v4,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       newCloudConfig,
	}

	return newConfig, nil
}

func migrateFromV2(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v2.KurtosisConfigV2)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	var newClusters map[string]*v3.KurtosisClusterConfigV3
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = map[string]*v3.KurtosisClusterConfigV3{}
		for oldClusterName, oldClusterConfig := range castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config

			var newKubernetesConfig *v3.KubernetesClusterConfigV3
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v3.KubernetesClusterConfigV3{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
				}
			}

			newClusterConfig := &v3.KurtosisClusterConfigV3{
				Type:           oldClusterConfig.Type,
				Config:         newKubernetesConfig,
				LogsAggregator: nil,
			}
			newClusters[oldClusterName] = newClusterConfig
		}
	}

	var newCloudConfig *v3.KurtosisCloudConfigV3
	if castedOldConfig.CloudConfig != nil {
		newCloudConfig = &v3.KurtosisCloudConfigV3{
			ApiUrl:           castedOldConfig.CloudConfig.ApiUrl,
			Port:             castedOldConfig.CloudConfig.Port,
			CertificateChain: castedOldConfig.CloudConfig.CertificateChain,
		}
	}

	newConfig := &v3.KurtosisConfigV3{
		ConfigVersion:     config_version.ConfigVersion_v3,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       newCloudConfig,
	}

	return newConfig, nil
}

func migrateFromV1(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v1.KurtosisConfigV1)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}

	// Migrate cluster configs across
	var newClusters map[string]*v2.KurtosisClusterConfigV2
	if castedOldConfig.KurtosisClusters != nil {
		newClusters = map[string]*v2.KurtosisClusterConfigV2{}
		for oldClusterName, oldClusterConfig := range *castedOldConfig.KurtosisClusters {
			oldKubernetesConfig := oldClusterConfig.Config

			var newKubernetesConfig *v2.KubernetesClusterConfigV2
			if oldKubernetesConfig != nil {
				newKubernetesConfig = &v2.KubernetesClusterConfigV2{
					KubernetesClusterName:  oldKubernetesConfig.KubernetesClusterName,
					StorageClass:           oldKubernetesConfig.StorageClass,
					EnclaveSizeInMegabytes: oldKubernetesConfig.EnclaveSizeInMegabytes,
				}
			}

			newClusterConfig := &v2.KurtosisClusterConfigV2{
				Type:   oldClusterConfig.Type,
				Config: newKubernetesConfig,
			}
			newClusters[oldClusterName] = newClusterConfig
		}
	}

	// create a new configuration object to represent the migrated work
	newConfig := &v2.KurtosisConfigV2{
		ConfigVersion:     config_version.ConfigVersion_v2,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  newClusters,
		CloudConfig:       nil,
	}

	return newConfig, nil
}

func migrateFromV0(uncastedConfig interface{}) (interface{}, error) {
	// cast "uncastedConfig" to current version we're upgrading from
	castedOldConfig, ok := uncastedConfig.(*v0.KurtosisConfigV0)
	if !ok {
		return nil, stacktrace.NewError(
			"Failed to cast old configuration '%+v' to expected configuration struct",
			uncastedConfig,
		)
	}
	// create a new configuration object to represent the migrated work
	newConfig := &v1.KurtosisConfigV1{
		ConfigVersion:     config_version.ConfigVersion_v1,
		ShouldSendMetrics: castedOldConfig.ShouldSendMetrics,
		KurtosisClusters:  nil,
	}
	return newConfig, nil
}
