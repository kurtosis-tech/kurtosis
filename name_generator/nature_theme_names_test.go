package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_generateNatureThemeNameForArtifactsInternal(t *testing.T) {
	args := generatorArgs{
		adjectives: []string{"test_adj", "test_adj_two"},
		nouns:      []string{"noun"},
	}

	actual := generateNatureThemeNameForArtifactsInternal(args)
	potentialCandidates := map[string]bool{
		"test_adj-noun":     true,
		"test_adj_two-noun": true,
	}

	require.Contains(t, potentialCandidates, actual)
}

func Test_convertMapSetToStringArray(t *testing.T) {
	data := map[string]bool{
		"test":             true,
		"test_key_another": true,
	}

	actual := convertMapSetToStringArray(data)

	require.Contains(t, actual, "test")
	require.Contains(t, actual, "test_key_another")
	require.NotContains(t, actual, "abc")
	require.Len(t, actual, len(data))
}
