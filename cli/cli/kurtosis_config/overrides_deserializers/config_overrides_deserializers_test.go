package overrides_deserializers

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/config_version"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOverridesDeserializerCompletenessTest(t *testing.T) {
	for _, configVersion := range config_version.ConfigVersionValues() {
		_, found := allConfigOverridesDeserializers[configVersion]
		require.True(t, found, "No config overrides deserializer found for config version '%v'; you'll need to add one", configVersion.String())
	}
	numDeserializers := len(allConfigOverridesDeserializers)
	numConfigVersions := len(config_version.ConfigVersionValues())
	require.Equal(
		t,
		numConfigVersions,
		numDeserializers,
		"There are %v Kurtosis config versions but %v config overrides deserializers were declared; this likely means " +
			"an extra deserializer that shouldn't be",
	)
}
