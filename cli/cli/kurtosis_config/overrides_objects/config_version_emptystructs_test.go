package overrides_objects

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/config_version"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfigVersionEmptystructsCompleteness(t *testing.T) {
	for _, configVersion := range config_version.ConfigVersionValues() {
		_, found := AllConfigVersionEmptyStructs[configVersion]
		require.True(t, found, "No empty struct found for config version '%v'; you'll need to add one", configVersion.String())
	}
	numEmptystructs := len(AllConfigVersionEmptyStructs)
	numConfigVersions := len(config_version.ConfigVersionValues())
	require.Equal(
		t,
		numConfigVersions,
		numEmptystructs,
		"There are %v Kurtosis config versions but %v config version emptystructs were declared; this likely means " +
			"an extra emptystruct was declared that shouldn't be",
	)
}