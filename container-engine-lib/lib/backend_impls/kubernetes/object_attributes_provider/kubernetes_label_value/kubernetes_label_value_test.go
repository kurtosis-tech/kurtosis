package kubernetes_label_value

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testLabelValuesWithValidity = map[string]bool{
	"":        true,
	" ":       false,
	"a":       true,
	"aaa":     true,
	"aAa":     true,
	"a99a9":   true,
	"a.7.3.5": true,
	"foo_bar": true,
	"foo/bar": false,
	"foo-bar": true,
}

func TestEdgeCases(t *testing.T) {
	for value, shouldPass := range testLabelValuesWithValidity {
		_, err := CreateNewKubernetesLabelValue(value)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected label value string '%v' validity to be '%v' but was '%v'", value, shouldPass, didPass)
	}
}

const maxLength = 63

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat("a", maxLength+1)
	_, err := CreateNewKubernetesLabelValue(invalidLabel)
	require.Error(t, err)
}

func TestJustLongEnough(t *testing.T) {
	validLabel := strings.Repeat("a", maxLength)
	_, err := CreateNewKubernetesLabelValue(validLabel)
	require.NoError(t, err)
}
