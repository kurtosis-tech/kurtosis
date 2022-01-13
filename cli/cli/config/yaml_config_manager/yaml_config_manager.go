package yaml_config_manager

import (
	"encoding/json"
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
)

const (
	kurtosisCLIConfigFilePermissions os.FileMode = 0644
)

var currentYAMLConfig *yamlConfig

type yamlConfig struct {
	content *content
}

type content struct {
	MetricsPromptShowed bool `yaml:"metrics-prompt-showed"`
	AcceptSendingMetrics bool `yaml:"accept-sending-metrics"`
}

func newYAMLConfig(content *content) *yamlConfig {
	return &yamlConfig{content: content}
}

func (config *yamlConfig) HasMetricsConsentPromptBeenDisplayed() bool {
	return config.content.MetricsPromptShowed
}

func (config *yamlConfig) MetricsConsentPromptHasBeenDisplayed() {
	config.content.MetricsPromptShowed = true
}

func (config *yamlConfig) HasUserAcceptedSendingMetrics() bool {
	return config.content.AcceptSendingMetrics
}

func (config *yamlConfig) UserAcceptSendingMetrics() {
	config.content.AcceptSendingMetrics = true
}

func (config *yamlConfig) UserDoNotAcceptSendingMetrics() {
	config.content.AcceptSendingMetrics = false
}

// Always execute the save method after doing any change in the configs
// We suggest to defer this method right after getting the configs
func (config *yamlConfig) Save() error {
	areEqual, err := isCurrentYAMLConfigContentEqualToConfigYAMLFileContent()
	//If the current content if the same as the YAML file content we avoid writing the config YAML file again
	if areEqual {
		logrus.Debugf("Current config content is equal to the CLI config YAML file content, avoiding to write the file")
		return nil
	}
	if err != nil{
		return stacktrace.Propagate(err, "An error occurred checking if current config is equal to the stored config YAML file")
	}

	kurtosisCLIConfigYAMLContent, err := yaml.Marshal(config.content)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling Kurtosis CLI config content '%+v' to a YAML content", config.content)
	}

	logrus.Debugf("Saving latest changes in Kurtosis CLI config YAML file...")
	kurtosisCLIConfigYAMLFilepath, err := getKurtosisCLIConfigYAMLFilepath()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(kurtosisCLIConfigYAMLFilepath, kurtosisCLIConfigYAMLContent, kurtosisCLIConfigFilePermissions)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing Kurtosis CLI config file '%v'", kurtosisCLIConfigYAMLFilepath)
	}
	logrus.Debugf("Kurtosis CLI config YAML file saved with values:\n%v", config.String())
	return nil
}

func (config *yamlConfig) String() string {
	configArrayByte, err := json.MarshalIndent(config.content, "", "\t")
	if err != nil {
		return "Config string won't be returned because and error occurred when trying to marshall the config struct. This is a bug in Kurtosis CLI"
	}
	configString := string(configArrayByte)
	return configString

}

func GetCurrentOrCreateDefaultConfig() (*yamlConfig, error) {

	if currentYAMLConfig != nil {
		return currentYAMLConfig, nil
	}

	kurtosisConfigYAMLFile, err := getKurtosisCLIConfigYAMLFileConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Kurtosis CLI config YAML file")
	}

	currentYAMLConfig = kurtosisConfigYAMLFile

	return kurtosisConfigYAMLFile, nil
}

// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
func isCurrentYAMLConfigContentEqualToConfigYAMLFileContent() (bool, error)  {
	if currentYAMLConfig == nil {
		return false, nil
	}

	currentContent := currentYAMLConfig.content

	yamlFile, err := getKurtosisCLIConfigYAMLFileConfig()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting Kurtosis CLI config YAML file content")
	}

	areEqual := reflect.DeepEqual(currentContent, yamlFile.content)

	return areEqual, nil
}

func getKurtosisCLIConfigYAMLFileConfig() (*yamlConfig, error) {
	kurtosisCLIConfigYAMLFilepath, err := getKurtosisCLIConfigYAMLFilepath()
	if err != nil {
		return nil, err
	}
	kurtosisConfigYAMLFile, err := os.Open(kurtosisCLIConfigYAMLFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("The Kurtosis CLI config YAML file has not been created yet")
			yamlCfg, err := createDefaultKurtosisCLIConfigYAMLFile()
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating default Kurtosis CLI config YAML file")
			}
			return yamlCfg, nil
		} else {
			return nil, stacktrace.Propagate(err, "An error occurred opening the Kurtosis CLI config YAML file content")
		}
	}
	defer func(){
		if err := kurtosisConfigYAMLFile.Close(); err != nil {
			logrus.Warnf("We tried to close the Kurtosis CLI config YAML file, but doing so threw an error:\n%v", err)
		}
	}()

	fileContentBytes, err := ioutil.ReadAll(kurtosisConfigYAMLFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis CLI config file")
	}

	yamlContent := &content{}

	if err := yaml.Unmarshal(fileContentBytes, yamlContent); err != nil{
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis CLI config YAML file content '%v'", fileContentBytes)
	}
	kurtosisCLIConfig := newYAMLConfig(yamlContent)

	return kurtosisCLIConfig, nil
}

func createDefaultKurtosisCLIConfigYAMLFile() (*yamlConfig, error){
	logrus.Debugf("Creating the default Kurtosis CLI config YAML file...")
	//Create content with default values
	kurtosisCLIConfigContent := &content{}
	kurtosisCLIConfig := newYAMLConfig(kurtosisCLIConfigContent)

	if err := kurtosisCLIConfig.Save(); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred saving Kurtosis CLI config YAML file")
	}
	logrus.Debugf("Default Kurtosis CLI config YAML file succesfully created")

	return kurtosisCLIConfig, nil
}

func getKurtosisCLIConfigYAMLFilepath() (string, error) {
	kurtosisCLIConfigYAMLFilepath, err := host_machine_directories.GetKurtosisCLIConfigYAMLFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis CLI config YAML filepath")
	}
	logrus.Debugf("Kurtosis CLI config YAML filepath: '%v'", kurtosisCLIConfigYAMLFilepath)
	return kurtosisCLIConfigYAMLFilepath, nil
}
