package cli_config_manager

import (
	"github.com/go-yaml/yaml"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

const (
	kurtosisCLLIConfigFilePermissions os.FileMode = 0644
)

var currentKurtosisCLIConfig *KurtosisCLIConfig

type KurtosisCLIConfig struct {
	content *KurtosisCLIConfigContent
}

func NewKurtosisCLIConfig(content *KurtosisCLIConfigContent) *KurtosisCLIConfig{
	return &KurtosisCLIConfig{content: content}
}

func (config *KurtosisCLIConfig) GetAcceptSendingMetrics() bool {
	return config.content.AcceptSendingMetrics //TODO asegurarse de que content no sea nil
}

type KurtosisCLIConfigContent struct {
	AcceptSendingMetrics bool `yaml:"accept-sending-metrics"`
}

func GetKurtosisCLIConfig() (*KurtosisCLIConfig, error) {

	if currentKurtosisCLIConfig != nil {
		return currentKurtosisCLIConfig, nil
	}

	kurtosisCLIConfig := &KurtosisCLIConfig{}

	kurtosisCLIConfigFilepath, err := getKurtosisCLIConfigFilepath()
	if err != nil {
		return nil, err
	}

	kurtosisConfigFile, err := os.Open(kurtosisCLIConfigFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("The Kurtosis CLI config file has not be created yet.")
			kurtosisCLIConfig, err = createDefaultKurtosisCLIConfigFile()
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred creating default Kurtosis CLI config file")
			}
		}
		return nil, stacktrace.Propagate(err, "An error occurred opening the '%v' file", kurtosisCLIConfigFilepath)
	} else  {
		fileContentBytes, err := ioutil.ReadAll(kurtosisConfigFile)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred reading Kurtosis CLI config file")
		}

		if err := yaml.Unmarshal(fileContentBytes, kurtosisCLIConfig); err != nil{
			return nil, stacktrace.Propagate(err, "An error occurred unmarshalling Kurtosis CLI config file content '%v'", fileContentBytes)
		}
	}
	defer func() {
		if err := kurtosisConfigFile.Close(); err != nil {
			logrus.Warnf("We tried to close the Kurtosis CLI config file, but doing so threw an error:\n%v", err)
		}
	}()

	currentKurtosisCLIConfig = kurtosisCLIConfig

	return kurtosisCLIConfig, nil
}

func createDefaultKurtosisCLIConfigFile() (*KurtosisCLIConfig, error){
	logrus.Debugf("Creating the default Kurtosis CLI config file...")
	kurtosisCLIConfigContent := &KurtosisCLIConfigContent{}
	kurtosisCLIConfig := NewKurtosisCLIConfig(kurtosisCLIConfigContent)

	kurtosisCLIConfigYAMLContent, err := yaml.Marshal(&kurtosisCLIConfigContent)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred marshalling Kurtosis CLI config content '%+v' to a YAML content", kurtosisCLIConfigContent)
	}

	kurtosisCLIConfigFilepath, err := getKurtosisCLIConfigFilepath()
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(kurtosisCLIConfigFilepath, kurtosisCLIConfigYAMLContent, kurtosisCLLIConfigFilePermissions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred writing Kurtosis CLI config file '%v'", kurtosisCLIConfigFilepath)
	}
	logrus.Debugf("Kurtosis CLI Config file creation ends")

	return kurtosisCLIConfig, nil
}

func getKurtosisCLIConfigFilepath() (string, error) {
	kurtosisCLIConfigFilepath, err := host_machine_directories.GetKurtosisCLIConfigFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis CLI config filepath")
	}
	logrus.Debugf("Kurtosis CLI config filepath: '%v'", kurtosisCLIConfigFilepath)
	return kurtosisCLIConfigFilepath, nil
}
