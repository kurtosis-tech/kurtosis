package kurtosis_config

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
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
func (configStore *kurtosisConfigStore) saveKurtosisConfigYAMLFile(kurtosisConfig *KurtosisConfig) error {
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

	newKurtosisConfig := &KurtosisConfigV1{}

	if err := yaml.Unmarshal(fileContentBytes, newKurtosisConfig); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", string(fileContentBytes))
	}

	// ENSURE BACKWARDS COMPATIBILITY - CORRECTLY READ PREVIOUSLY WRITTEN PREFERENCES FROM DIFFERENT CONFIG VERSIONS

	// READ V0 AND CONVERT TO V1
	if newKurtosisConfig.ConfigVersion == nil && newKurtosisConfig.ShouldSendMetrics != nil {
		defaultKurtosisConfig := NewDefaultKurtosisConfig(newKurtosisConfig.ShouldSendMetrics)
		return defaultKurtosisConfig, nil
	}

	return NewKurtosisConfigVersionedConfig(newKurtosisConfig), nil
}
