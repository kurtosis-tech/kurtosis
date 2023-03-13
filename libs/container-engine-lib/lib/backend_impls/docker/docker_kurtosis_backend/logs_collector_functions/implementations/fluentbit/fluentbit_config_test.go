package fluentbit

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetOutputKurtosisLabelsForLogs(t *testing.T) {
	expectedKurtosisFluentbitOutputLabels := []string{
		"$comKurtosistechGuid",
		"$comKurtosistechContainerType",
	}

	fluentbitKurtosisOutputLabels := getOutputKurtosisLabelsForLogs()
	require.Equal(t, expectedKurtosisFluentbitOutputLabels, fluentbitKurtosisOutputLabels)
}

func TestGetModifyFilterRulesKurtosisLabels(t *testing.T) {

	expectedKurtosisGUIDDockerLabelRenameFilterRule := "rename com.kurtosistech.guid comKurtosistechGuid"
	expectedKurtosisContainerTypeDockerLabelRenameFilterRule := "rename com.kurtosistech.guid comKurtosistechGuid"
	expectedAmountFilterRules := 2

	filterRulesKurtosisLabels := getModifyFilterRulesKurtosisLabels()
	require.Contains(t, filterRulesKurtosisLabels, expectedKurtosisGUIDDockerLabelRenameFilterRule)
	require.Contains(t, filterRulesKurtosisLabels, expectedKurtosisContainerTypeDockerLabelRenameFilterRule)
	require.Equal(t, expectedAmountFilterRules, len(filterRulesKurtosisLabels))
}