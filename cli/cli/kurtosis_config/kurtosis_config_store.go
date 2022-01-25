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
	currentKurtosisConfigStore *kurtosisConfigStore
	once sync.Once
)


type kurtosisConfigStore struct {
	mutex *sync.RWMutex
}

func GetKurtosisConfigStore() *kurtosisConfigStore {
	once.Do(func() {
		currentKurtosisConfigStore = &kurtosisConfigStore{mutex: &sync.RWMutex{}}
	})
	return currentKurtosisConfigStore
}

func (configStore *kurtosisConfigStore) HasConfig() bool {

	kurtosisConfig, err := configStore.getKurtosisConfigFromYAMLFile()
	if err != nil {
		if !os.IsNotExist(err) {
			logrus.Warnf("An error occurred getting Kurtosis config from YAML file, error:\n%v", err)
		}
		return false
	}

	if kurtosisConfig != nil {
		return true
	}

	return false
}

//TDOD if this process tends to be slow we could improve performance applying "Write Through and Write Back in Cache"
func (configStore *kurtosisConfigStore) GetConfig() (*KurtosisConfig, error) {

	kurtosisConfig, err := configStore.getKurtosisConfigFromYAMLFile()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config from YAML file")
	}

	return kurtosisConfig, nil
}

func (configStore *kurtosisConfigStore) SetConfig(kurtosisConfig *KurtosisConfig) error {

	if err := configStore.saveKurtosisConfigYAMLFile(kurtosisConfig); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving Kurtosis config YAML file")
	}

	return nil
}
// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func (configStore *kurtosisConfigStore) saveKurtosisConfigYAMLFile(kurtosisConfig *KurtosisConfig) error {
	kurtosisConfigYAMLContent, err := yaml.Marshal(kurtosisConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis config '%+v'", kurtosisConfig)
	}

	logrus.Debugf("Saving latest changes in Kurtosis config YAML file...")
	kurtosisConfigYAMLFilepath, err := getKurtosisConfigYAMLFilepath()
	if err != nil {
		return err
	}

	configStore.mutex.Lock()
	defer configStore.mutex.Unlock()

	err = ioutil.WriteFile(kurtosisConfigYAMLFilepath, kurtosisConfigYAMLContent, kurtosisConfigFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing Kurtosis config YAML file '%v'", kurtosisConfigYAMLFilepath)
	}
	logrus.Debugf("Kurtosis config YAML file saved")
	return nil
}

func (configStore *kurtosisConfigStore) getKurtosisConfigFromYAMLFile() (*KurtosisConfig, error) {
	kurtosisConfigYAMLFilepath, err := getKurtosisConfigYAMLFilepath()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis config YAML filepath")
	}
	logrus.Debugf("Kurtosis config YAML filepath: '%v'", kurtosisConfigYAMLFilepath)

	configStore.mutex.RLock()
	defer configStore.mutex.RUnlock()
	kurtosisConfigYAMLFile, err := os.Open(kurtosisConfigYAMLFilepath)
	if err != nil {
		return nil, err
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

	newKurtosisConfig := &KurtosisConfig{}

	if err := yaml.Unmarshal(fileContentBytes, newKurtosisConfig); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", fileContentBytes)
	}

	return newKurtosisConfig, nil
}

func getKurtosisConfigYAMLFilepath() (string, error) {
	kurtosisConfigYAMLFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML filepath")
	}
	return kurtosisConfigYAMLFilepath, nil
}
