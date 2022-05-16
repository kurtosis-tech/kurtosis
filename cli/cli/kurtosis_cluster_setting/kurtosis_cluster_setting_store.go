package kurtosis_cluster_setting

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sync"
)

const (
	kurtosisClusterSettingFilePermissions os.FileMode = 0644
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	currentKurtosisClusterSettingStore *kurtosisClusterSettingStore
	once sync.Once
)

type kurtosisClusterSettingStore struct {
	mutex *sync.RWMutex
}

func GetKurtosisClusterSettingStore() *kurtosisClusterSettingStore {
	// NOTE: We use a 'once' to initialize the kurtosisClusterSettingStore because it contains a mutex to guard
	//the setting file, and we don't ever want multiple kurtosisClusterSettingStore instances in existence
	once.Do(func() {
		currentKurtosisClusterSettingStore = &kurtosisClusterSettingStore{mutex: &sync.RWMutex{}}
	})
	return currentKurtosisClusterSettingStore
}

func (settingStore *kurtosisClusterSettingStore) HasClusterSetting() (bool, error) {
	settingStore.mutex.RLock()
	defer settingStore.mutex.RUnlock()

	fileExists, err := settingStore.doesClusterSettingFilepathExist()
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to discover if cluster setting filepath exists.")
	}
	return fileExists, nil
}

func (settingStore *kurtosisClusterSettingStore) SetClusterSetting(clusterName string) error {
	settingStore.mutex.Lock()
	defer settingStore.mutex.Unlock()

	err := settingStore.saveClusterSettingFile(clusterName)
	return err
}

func (settingStore *kurtosisClusterSettingStore) GetClusterSetting() (string, error) {
	settingStore.mutex.Lock()
	defer settingStore.mutex.Unlock()

	name, err := settingStore.getClusterSettingFromFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to get cluster setting from setting file.")
	}
	return name, nil
}


// ======================================== Private Helpers ===========================================
func (settingStore *kurtosisClusterSettingStore)  doesClusterSettingFilepathExist() (bool, error){
	filepath, err := host_machine_directories.GetKurtosisClusterSettingFilepath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the cluster setting filepath")
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred checking if filepath '%v' exists", filepath)
	}
	return true, nil
}

func (settingStore *kurtosisClusterSettingStore) getClusterSettingFromFile() (string, error) {
	filepath, err := host_machine_directories.GetKurtosisClusterSettingFilepath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the cluster setting filepath")
	}
	logrus.Debugf("Cluster setting filepath: '%v'", filepath)

	fileContentBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading cluster setting file")
	}

	fileContentStr := string(fileContentBytes)

	return fileContentStr, nil
}


func (settingStore *kurtosisClusterSettingStore) saveClusterSettingFile(clusterName string) error {
	validClusterName, err := settingStore.validateClusterName(clusterName)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to validate cluster setting.")
	}
	if !validClusterName {
		return stacktrace.NewError("Cluster name '%v' is not a valid Kurtosis cluster name.", clusterName)
	}

	fileContent := []byte(clusterName)

	logrus.Debugf("Saving cluster setting in file...")

	filepath, err := host_machine_directories.GetKurtosisClusterSettingFilepath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the cluster setting filepath")
	}

	err = ioutil.WriteFile(filepath, fileContent, kurtosisClusterSettingFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing cluster setting file '%v'", filepath)
	}
	logrus.Debugf("Cluster setting file saved")
	return nil
}

func (settingStore *kurtosisClusterSettingStore) validateClusterName(clusterName string) (bool, error) {
	configStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(configStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return false, stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration when validating cluster name '%v'.", clusterName)
	}
	if _, ok := kurtosisConfig.GetKurtosisClusters()[clusterName]; ok {
		return true, nil
	}
	return false, nil
}
