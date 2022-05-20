package overrides_migrators

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v0"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v1"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v2"
	"github.com/kurtosis-tech/stacktrace"
)

/*
This file contains functions that will migrate version N of the config overrides to version N+1
 */

// Takes a version of the config, casts it, migrates it to the N+1 version, and returns it
type configOverridesMigrator = func(uncastedOldConfig interface{}) (interface{}, error)

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
// Adding a new version here is:
//   1) creating a new migrateFromVX function, where X = the latest-1 config version
//   2) adding an entry for the latest-1 config version with your new function
// We keep these sorted in REVERSE chronological order so you don't need to scroll
//  to the bottom each time
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>> INSTRUCTIONS <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
var allConfigOverridesMigrators = map[config_version.ConfigVersion]configOverridesMigrator{
	config_version.ConfigVersion_v1: migrateFromV1,
	config_version.ConfigVersion_v0: migrateFromV0,
}

// vvvvvvvvvvvvvvvvvvvvvvv REVERSE chronological order so you don't have to scroll forever vvvvvvvvvvvvvvvvvvvv
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

