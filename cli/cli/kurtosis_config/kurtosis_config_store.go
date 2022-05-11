package kurtosis_config

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v0"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	kurtosisConfigFilePermissions os.FileMode = 0644
	latestConfigFileVersion = 1
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
func (configStore *kurtosisConfigStore) GetConfig() (*KurtosisConfig, error) {
	configStore.mutex.RLock()
	defer configStore.mutex.RUnlock()

	kurtosisConfig, err := configStore.getKurtosisConfigFromYAMLFile()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config from YAML file")
	}

	return kurtosisConfig, nil
}

func (configStore *kurtosisConfigStore) SetConfig(kurtosisConfig *KurtosisConfig) error {
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

type kurtosisConfigVersionDetector struct {
	ConfigVersion *int `yaml:"config-version"`
}

func (configStore *kurtosisConfigStore) saveKurtosisConfigYAMLFile(kurtosisConfig *KurtosisConfig) error {
	err := kurtosisConfig.Validate()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to validate configuration before saving to file.")
	}
	kurtosisConfigYAMLContent, err := yaml.Marshal(kurtosisConfig.GetVersionSpecificConfig())
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

func (configStore *kurtosisConfigStore) getKurtosisConfigFromYAMLFile() (*KurtosisConfig, error) {
	fileContentBytes, err := configStore.readConfigFileBytes()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis config YAML file")
	}

	configVersionDetector := &kurtosisConfigVersionDetector{}
	if err := yaml.Unmarshal(fileContentBytes, configVersionDetector); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
	}

	// If ConfigVersion pointer is nil, the config version is V0
	configVersion := 0
	if configVersionDetector.ConfigVersion != nil {
		configVersion = *configVersionDetector.ConfigVersion
	}

	kurtosisConfig, err := configStore.migrateOverridesAcrossYAMLVersions(configVersion)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to migrate overrides across YAML versions, starting with config version %d", configVersion)
	}
	err = kurtosisConfig.Validate()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to validate configuration when reading from file.")
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

func  (configStore *kurtosisConfigStore) migrateOverridesAcrossYAMLVersions(configVersionOnDisk int) (*KurtosisConfig, error) {
	v0ConfigOverrides := &v0.KurtosisConfigV0{}
	v1ConfigOverrides := &v1.KurtosisConfigV1{}
	fileContentBytes, err := configStore.readConfigFileBytes()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis config YAML file")
	}

	if configVersionOnDisk == 0 {
		if err := yaml.Unmarshal(fileContentBytes, v0ConfigOverrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
		}
	} else if configVersionOnDisk == 1 {
		if err := yaml.Unmarshal(fileContentBytes, v1ConfigOverrides); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
		}
	} else {
		return nil, stacktrace.NewError("Read invalid configuration version number %d from Kurtosis configuration file", configVersionOnDisk)
	}

	// PASS OVERRIDES STRUCT THROUGH A SERIES OF MIGRATIONS
	for versionToUpgradeTo := configVersionOnDisk; versionToUpgradeTo <= latestConfigFileVersion; versionToUpgradeTo++ {
		/*
			If versionToUpgradeTo is 0, there are no upgrade actions because
			v0 is the first version (no previous version to upgrade from)
		 */

		// If versionToUpgradeTo is 1, migrate overrides from v0 to v1
		if versionToUpgradeTo == 1 {
			// Migrate overrides from V0 to V1
			if v0ConfigOverrides.ShouldSendMetrics != nil {
				v1ConfigOverrides.ShouldSendMetrics = v0ConfigOverrides.ShouldSendMetrics
			}
		}
		// =====================================
		// WHEN ADDING ANOTHER YAML VERSION, ADD ANOTHER IF STATEMENT
		// TO HANDLE MIGRATION BETWEEN THE LATEST VERSION ABOVE THIS COMMENT
		// AND THE NEWEST VERSION YOU'RE ADDING
		// ==================================
	}

	// Initialize default configuration for the latest version available
	kurtosisConfig := v1.NewDefaultKurtosisConfigV1()
	// Overlay overrides that are now stored in the latest versions' override struct
	kurtosisConfig.OverlayOverrides(v1ConfigOverrides)
	return NewKurtosisConfig(kurtosisConfig), nil
}