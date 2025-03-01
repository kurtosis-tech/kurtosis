package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/overrides_objects/v3"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

/*
Explanation: when a Kurtosis dev adds a new config version, we want to make sure they update the KurtosisConfig
wrapper to use it, so KurtosisConfig is always using the latest-declared config. The trick we came up with to do
so is:
1. Get the latest config version (which we can do dynamically thanks to 'enumer' enum generation)
2. Plug the latest config version into the empty-structs-per-config-version map to get an instance of the latest config struct
3. Ask KurtosisConfig to try and cast it

If it can cast it, KurtosisConfig is using the latest version; if it can't
then KurtosisConfig is using an old version.
*/
func TestKurtosisConfigIsUsingLatestConfigStruct(t *testing.T) {
	// Dynamically get latest config version
	var latestConfigVersion config_version.ConfigVersion
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	latestEmptystruct, found := overrides_objects.AllConfigVersionEmptyStructs[latestConfigVersion]
	require.True(t, found, "No config emptystruct was defined for latest config version '%v'; you'll need to define one there", latestConfigVersion)

	_, err := castUncastedOverrides(latestEmptystruct)
	require.NoError(
		t,
		err,
		"An error occurred casting an emptystruct of the latest config version, indicating that KurtosisConfig is not "+
			"using the latest config struct; update KurtosisConfig to use the latest config struct version!",
	)
}

func TestNewKurtosisConfigFromRequiredFields_MetricsElectionIsSent(t *testing.T) {
	config, err := NewKurtosisConfigFromRequiredFields(false)
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.NotNil(t, overrides.ShouldSendMetrics)
}

func TestNewKurtosisConfigEmptyOverrides(t *testing.T) {
	_, err := NewKurtosisConfigFromOverrides(&v3.KurtosisConfigV3{
		ConfigVersion:     0,
		ShouldSendMetrics: nil,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
		LogsAggregator:    nil,
	})
	// You can not initialize a Kurtosis config with empty overrides - it needs at least `ShouldSendMetrics`
	require.Error(t, err)
}

func TestNewKurtosisConfigJustMetrics(t *testing.T) {
	version := config_version.ConfigVersion_v0
	shouldSendMetrics := true
	originalOverrides := v3.KurtosisConfigV3{
		ConfigVersion:     version,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
		LogsAggregator:    nil,
	}
	config, err := NewKurtosisConfigFromOverrides(&originalOverrides)
	// You can not initialize a Kurtosis config with empty originalOverrides - it needs at least `ShouldSendMetrics`
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.NotNil(t, overrides.ShouldSendMetrics)
}

func TestNewKurtosisConfigOverridesAreLatestVersion(t *testing.T) {
	config, err := NewKurtosisConfigFromRequiredFields(false)
	require.NoError(t, err)

	configValues := config_version.ConfigVersionStrings()
	sort.Strings(configValues)
	latestVersion := configValues[len(configValues)-1]

	overrides := config.GetOverrides()
	// check that overrides are actually the latest version
	require.Equal(t, latestVersion, overrides.ConfigVersion.String())
}

func TestCloudConfigOverridesApiUrl(t *testing.T) {
	version := config_version.ConfigVersion_v3
	shouldSendMetrics := true
	apiUrl := "test.com"
	originalOverrides := v3.KurtosisConfigV3{
		ConfigVersion:     version,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig: &v3.KurtosisCloudConfigV3{
			ApiUrl:           &apiUrl,
			Port:             nil,
			CertificateChain: nil,
		},
		LogsAggregator: nil,
	}
	config, err := NewKurtosisConfigFromOverrides(&originalOverrides)
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.Equal(t, apiUrl, *overrides.CloudConfig.ApiUrl)
	require.Nil(t, overrides.CloudConfig.Port)
	require.Nil(t, overrides.CloudConfig.CertificateChain)
}

func TestCloudConfigReconciliation(t *testing.T) {
	version := config_version.ConfigVersion_v3
	shouldSendMetrics := true
	apiUrl := "test.com"
	originalOverrides := v3.KurtosisConfigV3{
		ConfigVersion:     version,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig: &v3.KurtosisCloudConfigV3{
			ApiUrl:           &apiUrl,
			Port:             nil,
			CertificateChain: nil,
		},
		LogsAggregator: nil,
	}

	config, err := NewKurtosisConfigFromOverrides(&originalOverrides)
	require.NoError(t, err)

	// Test reconciliation behaviour
	require.Equal(t, shouldSendMetrics, config.GetShouldSendMetrics())
	require.Equal(t, apiUrl, config.GetCloudConfig().ApiUrl)
	require.Equal(t, DefaultCloudConfigPort, config.GetCloudConfig().Port)
	require.Equal(t, DefaultCertificateChain, config.GetCloudConfig().CertificateChain)
}

func TestLogsAggregatorOverrides(t *testing.T) {
	version := config_version.ConfigVersion_v3
	shouldSendMetrics := true
	sinks := map[string]map[string]interface{}{
		"elasticsearch": {
			"type": "elasticsearch",
		},
	}
	originalOverrides := v3.KurtosisConfigV3{
		ConfigVersion:     version,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
		LogsAggregator: &v3.KurtosisLogsAggregatorConfigV3{
			Sinks: sinks,
		},
	}
	config, err := NewKurtosisConfigFromOverrides(&originalOverrides)
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.Equal(t, sinks, overrides.LogsAggregator.Sinks)
}

func TestLogsAggregatorReconciliation(t *testing.T) {
	version := config_version.ConfigVersion_v3
	shouldSendMetrics := true
	sinks := map[string]map[string]interface{}{
		"elasticsearch": {
			"type": "elasticsearch",
		},
	}
	originalOverrides := v3.KurtosisConfigV3{
		ConfigVersion:     version,
		ShouldSendMetrics: &shouldSendMetrics,
		KurtosisClusters:  nil,
		CloudConfig:       nil,
		LogsAggregator: &v3.KurtosisLogsAggregatorConfigV3{
			Sinks: sinks,
		},
	}
	config, err := NewKurtosisConfigFromOverrides(&originalOverrides)
	require.NoError(t, err)

	// Test reconciliation behaviour
	require.Equal(t, shouldSendMetrics, config.GetShouldSendMetrics())
	require.Equal(t, sinks, config.GetLogsAggregatorConfig().Sinks)
}
