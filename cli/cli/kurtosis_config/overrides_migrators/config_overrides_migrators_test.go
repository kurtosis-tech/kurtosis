package overrides_migrators

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOverridesMigratorsCompletenessTest(t *testing.T) {
	// Dynamically get latest config version
	var latestConfigVersion config_version.ConfigVersion
	for _, configVersion := range config_version.ConfigVersionValues() {
		if uint(configVersion) > uint(latestConfigVersion) {
			latestConfigVersion = configVersion
		}
	}

	for _, configVersion := range config_version.ConfigVersionValues() {
		_, found := allConfigOverridesMigrators[configVersion]
		require.True(t, found, "No config overrides migrator found for config version '%v'; you'll need to add one", configVersion.String())
	}
	numMigrators := len(allConfigOverridesMigrators)
	numConfigVersions := len(config_version.ConfigVersionValues())
	require.Equal(
		t,
		numConfigVersions - 1,
		numMigrators,
		"There are %v Kurtosis config versions but %v config overrides migrators were declared; this likely means " +
			"extra migrators were declared that shouldn't be (there should always be migrators = num_versions - 1)",
	)
}
