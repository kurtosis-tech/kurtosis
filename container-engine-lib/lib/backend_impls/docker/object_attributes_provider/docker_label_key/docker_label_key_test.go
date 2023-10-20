package docker_label_key

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	oneRandomChar = "a"
)

var testLabelsWithValidity = map[string]bool{
	"":                        false,
	" ":                       false, // whitespace not allowed
	"a":                       true,
	"aaa":                     true,
	"aAa":                     false, // caps not allowed
	"a99a9":                   true,
	"a.7.3.5":                 true,
	"com.kurtosistech.app-id": true,
	"kurtosistech.com/app-id": false, // Kubernetes labels standard not allowed
}

func TestEdgeCaseLabels(t *testing.T) {
	for label, shouldPass := range testLabelsWithValidity {
		_, err := createNewDockerLabelKey(label)
		didPass := err == nil
		require.Equal(t, shouldPass, didPass, "Expected label key string '%v' validity to be '%v' but was '%v'", label, shouldPass, didPass)
		_, err = CreateNewDockerUserCustomLabelKey(label)
		didPass = err == nil
		require.Equal(t, shouldPass, didPass, "Expected user custom label key string '%v' validity to be '%v' but was '%v'", label, shouldPass, didPass)
	}
}

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat("a", 9999)
	_, err := createNewDockerLabelKey(invalidLabel)
	require.Error(t, err)
	_, err = CreateNewDockerUserCustomLabelKey(invalidLabel)
	require.Error(t, err)
}

func TestMaxAllowedLabel(t *testing.T) {
	validMaxLabel := strings.Repeat(oneRandomChar, maxLabelLength)
	_, err := createNewDockerLabelKey(validMaxLabel)
	require.NoError(t, err)

	overValidMaxLabel := validMaxLabel + oneRandomChar
	_, err = createNewDockerLabelKey(overValidMaxLabel)
	require.Error(t, err)

	userCustomValidMaxLabel := strings.Repeat(oneRandomChar, maxLabelLength-len(customUserLabelsKeyPrefixStr))
	_, err = CreateNewDockerUserCustomLabelKey(userCustomValidMaxLabel)
	require.NoError(t, err)

	overUserCustomValidMaxLabel := userCustomValidMaxLabel + oneRandomChar
	_, err = CreateNewDockerUserCustomLabelKey(overUserCustomValidMaxLabel)
	require.Error(t, err)
}
