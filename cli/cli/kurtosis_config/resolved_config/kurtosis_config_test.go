package resolved_config

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/overrides_objects/v1"
	"github.com/stretchr/testify/require"
	"sort"
	"testing"
)

func TestNewKurtosisConfigFromRequiredFields_MetricsElectionIsSent(t *testing.T) {
	config, err := NewKurtosisConfigFromRequiredFields(false)
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.NotNil(t, overrides.ShouldSendMetrics)
}

func TestNewKurtosisConfigEmptyOverrides(t *testing.T) {
	_, err := NewKurtosisConfigFromOverrides(&v1.KurtosisConfigV1{})
	// You can not initialize a Kurtosis config with empty overrides - it needs at least `ShouldSendMetrics`
	require.Error(t, err)
}

func TestNewKurtosisConfigJustMetrics(t *testing.T) {
	version := config_version.ConfigVersion_v0
	shouldSendMetrics := true
	originalOverrides := v1.KurtosisConfigV1{
		ConfigVersion: version,
		ShouldSendMetrics: &shouldSendMetrics,
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