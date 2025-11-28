package cloud

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	KurtosisCloudApiKeyEnvVarArg = "KURTOSIS_CLOUD_API_KEY"
)

func LoadApiKey() (*string, error) {
	apiKey := os.Getenv(KurtosisCloudApiKeyEnvVarArg)
	if len(apiKey) < 1 {
		return nil, stacktrace.NewError("No API Key was found. An API Key must be provided as env var %s", KurtosisCloudApiKeyEnvVarArg)
	}
	logrus.Info("Successfully Loaded API Key...")
	return &apiKey, nil
}

func GetCloudConfig() (*resolved_config.KurtosisCloudConfig, error) {
	// Get the configuration
	kurtosisConfigStore := kurtosis_config.GetKurtosisConfigStore()
	configProvider := kurtosis_config.NewKurtosisConfigProvider(kurtosisConfigStore)
	kurtosisConfig, err := configProvider.GetOrInitializeConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get or initialize Kurtosis configuration")
	}
	if kurtosisConfig.GetCloudConfig() == nil {
		return nil, stacktrace.Propagate(err, "No cloud config was found. This is an internal Kurtosis error.")
	}
	cloudConfig := kurtosisConfig.GetCloudConfig()

	if cloudConfig.Port == 0 {
		cloudConfig.Port = resolved_config.DefaultCloudConfigPort
	}
	if len(cloudConfig.ApiUrl) < 1 {
		cloudConfig.ApiUrl = resolved_config.DefaultCloudConfigApiUrl
	}
	if len(cloudConfig.CertificateChain) < 1 {
		cloudConfig.CertificateChain = resolved_config.DefaultCertificateChain
	}

	return cloudConfig, nil
}
