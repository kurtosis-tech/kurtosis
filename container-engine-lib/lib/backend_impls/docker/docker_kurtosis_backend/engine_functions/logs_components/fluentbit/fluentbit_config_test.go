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