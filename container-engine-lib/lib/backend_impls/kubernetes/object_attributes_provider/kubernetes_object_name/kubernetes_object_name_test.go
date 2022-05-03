package kubernetes_object_name

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var testNamesWithValidity = map[string]bool{
	"":                                   false,
	" ":                                  false,
	"a":                                  true,
	"aaa":                                true,
	"foOBaR":                             false,
	"aAa":                                false,
	"a99a9":                              true,
	"a.7.3.5":                            false,
	"kurtosis-engine--82388239--service": true,
	"foo.bar":                            false,
	"foo-bar":                            true,
	"foo_bar":                            false,
	"foobar$":                            false,
	"foo,bar":                            false,
	"foo bar":                            false,
	"foo/bar":                            false,
}

func TestEdgeCases(t *testing.T) {
	for value, shouldPass := range testNamesWithValidity {
		_, err := CreateNewKubernetesObjectName(value)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected object name string '%v' validity to be '%v' but was '%v'", value, shouldPass, didPass)
	}
}

const maxLength = 63

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat("a", maxLength+1)
	_, err := CreateNewKubernetesObjectName(invalidLabel)
	require.Error(t, err)
}

func TestJustLongEnough(t *testing.T) {
	validLabel := strings.Repeat("a", maxLength)
	_, err := CreateNewKubernetesObjectName(validLabel)
	require.NoError(t, err)
}
