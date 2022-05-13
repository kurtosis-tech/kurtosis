package kurtosis_config

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v0"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/versioned_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	kurtosisConfigFilePermissions os.FileMode = 0644
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentKurtosisConfigStore *kurtosisConfigStore
	once sync.Once
)


type kurtosisConfigStore struct {
	mutex *sync.RWMutex
}

func GetKurtosisConfigStore() *kurtosisConfigStore {
	// NOTE: We use a 'once' to initialize the KurtosisConfigStore because it contains a mutex to guard
	//the config file, and we don't ever want multiple KurtosisConfigStore instances in existence
	once.Do(func() {
		currentKurtosisConfigStore = &kurtosisConfigStore{mutex: &sync.RWMutex{}}
	})
	return currentKurtosisConfigStore
}

func (configStore *kurtosisConfigStore) HasConfig() (bool, error) {
	configStore.mutex.RLock()
	defer configStore.mutex.RUnlock()

	kurtosisConfigYAMLFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFilepath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML filepath")
	}

	kurtosisConfigFileInfo, err := os.Stat(kurtosisConfigYAMLFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, stacktrace.Propagate(err, "An error occurred getting Kurtosis config YAML file info")
		}
	}

	if kurtosisConfigFileInfo != nil {
		return true, nil
	}

	return false, nil
}

//TDOD if this process tends to be slow we could improve performance applying "Write Through and Write Back in Cache"
func (configStore *kurtosisConfigStore) GetConfig() (*resolved_config.KurtosisConfig, error) {
	configStore.mutex.RLock()
	defer configStore.mutex.RUnlock()

	kurtosisConfig, err := configStore.getKurtosisConfigFromYAMLFile()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config from YAML file")
	}

	return kurtosisConfig, nil
}

func (configStore *kurtosisConfigStore) SetConfig(kurtosisConfig *resolved_config.KurtosisConfig) error {
	configStore.mutex.Lock()
	defer configStore.mutex.Unlock()

	if err := configStore.saveKurtosisConfigYAMLFile(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving Kurtosis config YAML file")
	}

	return nil
}
// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================

func (configStore *kurtosisConfigStore) saveKurtosisConfigYAMLFile(kurtosisConfig *resolved_config.KurtosisConfig) error {
	kurtosisConfigYAMLContent, err := yaml.Marshal(kurtosisConfig.GetOverrides())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis config '%+v'", kurtosisConfig)
	}

	logrus.Debugf("Saving latest changes in Kurtosis config YAML file...")
	kurtosisConfigYAMLFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML filepath")
	}

	err = ioutil.WriteFile(kurtosisConfigYAMLFilepath, kurtosisConfigYAMLContent, kurtosisConfigFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing Kurtosis config YAML file '%v'", kurtosisConfigYAMLFilepath)
	}
	logrus.Debugf("Kurtosis config YAML file saved")
	return nil
}

func (configStore *kurtosisConfigStore) getKurtosisConfigFromYAMLFile() (*resolved_config.KurtosisConfig, error) {
	// Overlay overrides that are now stored in the latest versions' override struct
	// kurtosisConfig.OverlayOverrides(v1ConfigOverrides)
	kurtosisConfigOverrides, err := configStore.migrateOverridesAcrossYAMLVersions()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to migrate overrides across YAML versions, starting with config version")
	}

	kurtosisConfig, err := resolved_config.NewKurtosisConfigFromOverrides(kurtosisConfigOverrides)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Kurtosis config from overrides: %+v", kurtosisConfigOverrides)
	}
	return kurtosisConfig, nil
}

func (configStore *kurtosisConfigStore) readConfigFileBytes() ([]byte, error) {
	kurtosisConfigYAMLFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFilepath()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis config YAML filepath")
	}
	logrus.Debugf("Kurtosis config YAML filepath: '%v'", kurtosisConfigYAMLFilepath)

	kurtosisConfigYAMLFile, err := os.Open(kurtosisConfigYAMLFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred opening Kurtosis config YAML file")
	}
	defer func() {
		if err := kurtosisConfigYAMLFile.Close(); err != nil {
			logrus.Warnf("We tried to close the Kurtosis config YAML file, but doing so threw an error:\n%v", err)
		}
	}()

	fileContentBytes, err := ioutil.ReadAll(kurtosisConfigYAMLFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis config YAML file")
	}
	return fileContentBytes, nil
}

/*
	Takes in user overrides (which make come from any version from v0->current version), and migrates them across version upgrades
	to maintain as much backwards compatibility as possible.
	This is a delicate operation: be careful to write override migration logic carefully.
	Overrides are partial fillings of YAML structs for v0->current version. The correct process for ensuring backwards compatibility is:
		1. Migrate overrides sequentially from their own version up to the latest version
		2. Overlay migrated overrides on top of the "default" latest version YAML struct
 */
func  (configStore *kurtosisConfigStore) migrateOverridesAcrossYAMLVersions() (*v1.KurtosisConfigV1, error) {
	fileContentBytes, err := configStore.readConfigFileBytes()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis config YAML file")
	}

	configVersionDetector := &versioned_config.VersionedKurtosisConfig{
		ConfigVersion: config_version.ConfigVersion_v0,
	}
	if err := yaml.Unmarshal(fileContentBytes, configVersionDetector); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
	}
	configVersionOnDisk := configVersionDetector.ConfigVersion

	var uncastedConfig interface{}
	switch configVersionOnDisk {
	case config_version.ConfigVersion_v0:
		overrides := &v0.KurtosisConfigV0{}
		if err := yaml.Unmarshal(fileContentBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
		}
		uncastedConfig = overrides
	case config_version.ConfigVersion_v1:
		overrides := &v1.KurtosisConfigV1{}
		if err := yaml.Unmarshal(fileContentBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
		}
		uncastedConfig = overrides
	default:
		return nil, stacktrace.NewError("Read invalid configuration version number %d from Kurtosis configuration file", configVersionOnDisk)
	}

	// Dynamically get latest config version
	var latestConfigVersion config_version.ConfigVersion
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	// PASS OVERRIDES STRUCT THROUGH A SERIES OF MIGRATIONS
	for versionToUpgradeFrom := configVersionOnDisk; versionToUpgradeFrom < latestConfigVersion; versionToUpgradeFrom++ {
		switch versionToUpgradeFrom {
		case config_version.ConfigVersion_v0:
			// cast "uncastedConfig" to current version we're upgrading from
			castedOldConfig, ok := uncastedConfig.(*v0.KurtosisConfigV0)
			if !ok {
				return nil, stacktrace.NewError(
					"Failed to cast configuration '%+v' to expected configuration version v%d",
					uncastedConfig,
					versionToUpgradeFrom,
				)
			}
			// create a new configuration object to represent the migrated work
			newConfig := &v1.KurtosisConfigV1{}
			// do migration steps
			if castedOldConfig.ShouldSendMetrics != nil {
				newConfig.ShouldSendMetrics = castedOldConfig.ShouldSendMetrics
			}
			// uncast new config interface with upgrade done
			uncastedConfig = newConfig
		default:
			return nil, stacktrace.NewError(
				"Needed to migrate from config version '%v' to the next highest version but " +
					 "no migration was defined for this version; this is a bug in Kurtosis",
				versionToUpgradeFrom,
			)
		}

	}

	// cast back to expected latest version
	resultConfig, ok := uncastedConfig.(*v1.KurtosisConfigV1)
	if !ok {
		return nil, stacktrace.NewError("Failed to cast configuration '%+v' to expected configuration version %d.", uncastedConfig, latestConfigVersion)
	}
	return resultConfig, nil
}