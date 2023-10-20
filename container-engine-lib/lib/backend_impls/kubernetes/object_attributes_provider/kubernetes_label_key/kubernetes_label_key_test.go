package kubernetes_label_key

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	oneRandomChar  = "a"
	maxLabelLength = 63
)

var testLabelsWithValidity = map[string][]bool{
	"":                        {false, false},
	" ":                       {false, false}, // whitespace not allowed
	"a":                       {true, true},
	"aaa":                     {true, true},
	"aAa":                     {true, true}, // caps allowed
	"a99a9":                   {true, true},
	"a.7.3.5":                 {true, true},
	"kurtosistech.com/app-id": {true, false},
	"foo_blah":                {true, true},
	"com.kurtosistech.app-id": {true, true}, // Docker labels standard allowed
}

func TestEdgeCaseLabels(t *testing.T) {
	for label, shouldPass := range testLabelsWithValidity {
		_, err := createNewKubernetesLabelKey(label)
		didPass := err == nil
		require.Equal(t, shouldPass[0], didPass, "Expected label key string '%v' validity to be '%v' but was '%v'", label, shouldPass[0], didPass)
		_, err = CreateNewKubernetesUserCustomLabelKey(label)
		didPass = err == nil
		require.Equal(t, shouldPass[1], didPass, "Expected user custom label key string '%v' validity to be '%v' but was '%v'", label, shouldPass[1], didPass)
	}
}

func TestTooLongLabel(t *testing.T) {
	invalidLabel := strings.Repeat(oneRandomChar, 9999)
	_, err := createNewKubernetesLabelKey(invalidLabel)
	require.Error(t, err)
	_, err = CreateNewKubernetesUserCustomLabelKey(invalidLabel)
	require.Error(t, err)
}

func TestMaxAllowedLabel(t *testing.T) {
	validMaxLabel := strings.Repeat(oneRandomChar, maxLabelLength)
	_, err := createNewKubernetesLabelKey(validMaxLabel)
	require.NoError(t, err)

	overValidMaxLabel := validMaxLabel + oneRandomChar
	_, err = createNewKubernetesLabelKey(overValidMaxLabel)
	require.Error(t, err)

	userCustomValidMaxLabel := strings.Repeat(oneRandomChar, maxLabelLength)
	_, err = CreateNewKubernetesUserCustomLabelKey(userCustomValidMaxLabel)
	require.NoError(t, err)

	overUserCustomValidMaxLabel := userCustomValidMaxLabel + oneRandomChar
	_, err = CreateNewKubernetesUserCustomLabelKey(overUserCustomValidMaxLabel)
	require.Error(t, err)
}
