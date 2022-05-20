package kubernetes_annotation_key

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testAnnotationsWithValidity = map[string]bool{
	"":                        false,
	" ":                       false, // whitespace not allowed
	"a":                       true,
	"aaa":                     true,
	"aAa":                     true, // caps allowed
	"a99a9":                   true,
	"a.7.3.5":                 true,
	"kurtosistech.com/app-id": true,
}

func TestEdgeCaseLabels(t *testing.T) {
	for label, shouldPass := range testAnnotationsWithValidity {
		_, err := CreateNewKubernetesAnnotationKey(label)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected label key string '%v' validity to be '%v' but was '%v'", label, shouldPass, didPass)
	}
}

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat("a", 9999)
	_, err := CreateNewKubernetesAnnotationKey(invalidLabel)
	require.Error(t, err)
}

