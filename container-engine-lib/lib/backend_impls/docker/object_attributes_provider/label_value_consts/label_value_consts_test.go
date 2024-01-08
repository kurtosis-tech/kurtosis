package label_value_consts

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/stretchr/testify/require"
	"testing"
)

var labelValueStrsToEnsure = map[string]string{
	appIdLabelValueStr:                       "kurtosis",
	engineContainerTypeLabelValueStr:         "kurtosis-engine",
	logsCollectorContainerTypeLabelValueStr:  "kurtosis-logs-collector",
	logsAggregatorContainerTypeLabelValueStr: "kurtosis-logs-aggregator",
	reverseProxyContainerTypeLabelValueStr:   "kurtosis-reverse-proxy",
}

var labelValuesToEnsure = map[*docker_label_value.DockerLabelValue]string{
	AppIDDockerLabelValue:               "kurtosis",
	EngineContainerTypeDockerLabelValue: "kurtosis-engine",
	LogsAggregatorTypeDockerLabelValue:  "kurtosis-logs-aggregator",
	LogsCollectorTypeDockerLabelValue:   "kurtosis-logs-collector",
	ReverseProxyTypeDockerLabelValue:    "kurtosis-reverse-proxy",
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It is VERY important that certain constants don't get modified, else Kurtosis will silently lose track
// of preexisting resources (thereby causing a resource leak). This test ensures that certain constants
// are never modified.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func TestImmutableConstantsArentModified(t *testing.T) {
	for actualValue, expectedValue := range labelValueStrsToEnsure {
		require.Equal(t, expectedValue, actualValue, "An immutable label value string was modified! Got '%v' but should be '%v'", actualValue, expectedValue)
	}

	for labelKey, expectedValueStr := range labelValuesToEnsure {
		labelKeyStr := labelKey.GetString()
		require.Equal(t, expectedValueStr, labelKeyStr, "An immutable label value was modified! Got '%v' but should be '%v'", labelKeyStr, expectedValueStr)
	}
}
