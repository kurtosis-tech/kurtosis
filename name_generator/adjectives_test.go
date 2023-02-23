package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// check for duplicates for nouns and adjectives
func Test_noDuplicatesInAdjectives(t *testing.T) {
	adjectivesHash := map[string]bool{}
	for _, adj := range ADJECTIVES {
		_, found := adjectivesHash[adj]
		require.False(t, found, "Duplicate Error: found %v twice in ADJECTIVES", adj)
		adjectivesHash[adj] = true
	}

	// this will only be called if there are no duplicates
	require.True(t, true)
}
