package fluentbit

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetOutputKurtosisLabelsForLogs(t *testing.T) {
	expectedKurtosisFluentbitOutputLabels := []string{
		"$comKurtosistechGuid",
		"$comKurtosistechContainerType",
		"$comKurtosistechEnclaveId",
	}

	fluentbitKurtosisOutputLabels := getOutputKurtosisLabelsForLogs()
	require.Equal(t, expectedKurtosisFluentbitOutputLabels, fluentbitKurtosisOutputLabels)
}

func TestGetModifyFilterRulesKurtosisLabels(t *testing.T) {

	expectedKurtosisGUIDDockerLabelRenameFilterRule := "rename com.kurtosistech.guid comKurtosistechGuid"
	expectedKurtosisContainerTypeDockerLabelRenameFilterRule := "rename com.kurtosistech.guid comKurtosistechGuid"
	expectedKurtosisEnclaveIdDockerLabelRenameFilterRule := "rename com.kurtosistech.enclave-id comKurtosistechGuid"
	expectedAmountFilterRules := 3

	filterRulesKurtosisLabels := getModifyFilterRulesKurtosisLabels()
	require.Contains(t, filterRulesKurtosisLabels, expectedKurtosisGUIDDockerLabelRenameFilterRule, expectedKurtosisEnclaveIdDockerLabelRenameFilterRule)
	require.Contains(t, filterRulesKurtosisLabels, expectedKurtosisContainerTypeDockerLabelRenameFilterRule, expectedKurtosisEnclaveIdDockerLabelRenameFilterRule)
	require.Equal(t, expectedAmountFilterRules, len(filterRulesKurtosisLabels))
}
