package kurtosis_cluster_setting

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
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

type kurtosisClusterSetting struct {
	CurrentKurtosisCluster string `yaml:"current-kurtosis-cluster"`
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

	kurtosisClusterSettingFileYAMLPath, err := host_machine_directories.GetKurtosisClusterSettingYAMLFilepath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the Kurtosis cluster setting YAML filepath")
	}

	kurtosisClusterSettingFileInfo, err := os.Stat(kurtosisClusterSettingFileYAMLPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, stacktrace.Propagate(err, "An error occurred getting Kurtosis cluster setting YAML file info")
		}
	}

	if kurtosisClusterSettingFileInfo != nil {
		return true, nil
	}

	return false, nil
}

func (settingStore *kurtosisClusterSettingStore) SetConfigClusterSetting(clusterName string) error {
	settingStore.mutex.Lock()
	defer settingStore.mutex.Unlock()

	return nil
}

// ======================================== Private Helpers ===========================================

func (settingStore *kurtosisClusterSettingStore) saveConfigClusterFile(clusterName string) error {
	fileContent := []byte(clusterName)

	logrus.Debugf("Saving cluster setting in file...")

	filepath, err := host_machine_directories.GetKurtosisClusterSettingYAMLFilepath()
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
