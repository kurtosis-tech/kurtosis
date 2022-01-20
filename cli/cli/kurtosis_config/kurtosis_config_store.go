package kurtosis_config

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
)

const (
	kurtosisConfigFilePermissions os.FileMode = 0644
)

var currentKurtosisConfig *KurtosisConfig

type yamlContent struct {
	UserAcceptSendingMetrics bool `yaml:"user-accept-sending-metrics"`
}

func newYamlContent(userAcceptSendingMetrics bool) *yamlContent {
	return &yamlContent{UserAcceptSendingMetrics: userAcceptSendingMetrics}
}

type KurtosisConfigStore struct {
	mutex *sync.Mutex
}

func newKurtosisConfigStore() *KurtosisConfigStore {
	return &KurtosisConfigStore{mutex: &sync.Mutex{}}
}

func (configStore *KurtosisConfigStore) HasConfig() bool {
	if currentKurtosisConfig != nil {
		return true
	}

	yamlFileContent, err := getKurtosisConfigYAMLFileContent()
	if err != nil {
		if !os.IsNotExist(err) {
			logrus.Warnf("An error occurred getting Kurtosis config YAML file content, error:\n%v", err)
		}
		return false
	}

	if yamlFileContent != nil {
		return true
	}

	return false
}

func (configStore *KurtosisConfigStore) GetConfig() (*KurtosisConfig, error) {
	if currentKurtosisConfig == nil {
		kurtosisConfigYAMLFileContent, err := getKurtosisConfigYAMLFileContent()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML file content")
		}
		currentKurtosisConfig = newKurtosisConfigFromYAMLFileContent(kurtosisConfigYAMLFileContent)
	}

	return currentKurtosisConfig, nil
}

func (configStore *KurtosisConfigStore) SetConfig(kurtosisConfig *KurtosisConfig) error {
	currentKurtosisConfig = kurtosisConfig
	currentYAMLContent := newYamlContentFromCurrentKurtosisConfig()

	areEqual, err := isCurrentConfigEqualToStoredConfig(currentYAMLContent)
	//If the current content if the same as the YAML file content we avoid writing the config YAML file again
	if areEqual {
		logrus.Debugf("Current YAML content is equal to the Kurtosis config YAML file content, avoiding to write the file")
		return nil
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if current config is equal to the stored config YAML file")
	}

	if err := configStore.saveKurtosisConfigYAMLFile(currentYAMLContent); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving Kurtosis config YAML file")
	}

	return nil
}
// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func (configStore *KurtosisConfigStore) saveKurtosisConfigYAMLFile(yamlContent *yamlContent) error {
	kurtosisConfigYAMLContent, err := yaml.Marshal(yamlContent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis config content '%+v' to a YAML content", yamlContent)
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

func isCurrentConfigEqualToStoredConfig(currentYAMLContent *yamlContent) (bool, error) {
	if currentKurtosisConfig == nil {
		return false, nil
	}

	yamlFileContent, err := getKurtosisConfigYAMLFileContent()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred getting Kurtosis config YAML file content")
	}

	areEqual := reflect.DeepEqual(currentYAMLContent, yamlFileContent)

	return areEqual, nil
}

func getKurtosisConfigYAMLFileContent() (*yamlContent, error) {
	kurtosisConfigYAMLFilepath, err := getKurtosisConfigYAMLFilepath()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kurtosis config YAML filepath")
	}
	logrus.Debugf("Kurtosis config YAML filepath: '%v'", kurtosisConfigYAMLFilepath)
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

	newYAMLContent := &yamlContent{}

	if err := yaml.Unmarshal(fileContentBytes, newYAMLContent); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis config YAML file content '%v'", fileContentBytes)
	}

	return newYAMLContent, nil
}

func getKurtosisConfigYAMLFilepath() (string, error) {
	kurtosisConfigYAMLFilepath, err := host_machine_directories.GetKurtosisConfigYAMLFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML filepath")
	}
	return kurtosisConfigYAMLFilepath, nil
}

func newYamlContentFromCurrentKurtosisConfig() *yamlContent {
	yamlContent := newYamlContent(currentKurtosisConfig.shouldSendMetrics)
	return yamlContent
}

func newKurtosisConfigFromYAMLFileContent(yamlFileContent *yamlContent) *KurtosisConfig {
	kurtosisConfig := NewKurtosisConfig(yamlFileContent.UserAcceptSendingMetrics)
	return kurtosisConfig
}
