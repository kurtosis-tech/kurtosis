package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_generateNatureThemeNameForArtifactsInternal(t *testing.T) {
	args := GeneratorArgs{
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
