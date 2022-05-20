package kurtosis_config

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v0"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v1"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v2"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/resolved_config"
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

// Every Kurtosis config version except v0 should have the config-version key, so we'll use that to determine
// which version of config overrides the user is supply
type versionDetectingKurtosisConfig struct {
	ConfigVersion config_version.ConfigVersion `yaml:"config-version"`
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
	fileContentBytes, err := configStore.readConfigFileBytes()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading the Kurtosis config YAML file")
	}

	// Overlay overrides that are now stored in the latest versions' override struct
	// kurtosisConfig.OverlayOverrides(v1ConfigOverrides)
	kurtosisConfigOverrides, err := migrateConfigOverridesToLatest(fileContentBytes)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Failed to migrate config overrides to the latest version",
		)
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
func migrateConfigOverridesToLatest(configFileBytes []byte) (*v1.KurtosisConfigV1, error) {
	versionDetectingConfig := &versionDetectingKurtosisConfig{
		ConfigVersion: config_version.ConfigVersion_v0,
	}
	if err := yaml.Unmarshal(configFileBytes, versionDetectingConfig); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(configFileBytes))
	}

	configVersionOnDisk := versionDetectingConfig.ConfigVersion

	var uncastedConfig interface{}
	switch configVersionOnDisk {
	case config_version.ConfigVersion_v0:
		overrides := &v0.KurtosisConfigV0{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v' using config %v", string(configFileBytes), configVersionOnDisk.String())
		}
		uncastedConfig = overrides
	case config_version.ConfigVersion_v1:
		overrides := &v1.KurtosisConfigV1{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v' using config %v", string(configFileBytes), configVersionOnDisk.String())
		}
		uncastedConfig = overrides
	case config_version.ConfigVersion_v2:
		overrides := &v2.KurtosisConfigV2{}
		if err := yaml.Unmarshal(configFileBytes, overrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v' using config %v", string(configFileBytes), configVersionOnDisk.String())
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
			newConfig, err := migrateV0ToV1(uncastedConfig)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred migrating from config version '%v'", versionToUpgradeFrom)
			}
			uncastedConfig = newConfig
		case config_version.ConfigVersion_v1:
			newConfig, err := migrateV1ToV2(uncastedConfig)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred migrating from config version '%v'", versionToUpgradeFrom)
			}
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

