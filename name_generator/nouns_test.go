package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// check for duplicates for nouns and adjectives
func Test_noDuplicatesInNouns(t *testing.T) {
	nounsHash := map[string]bool{}
	for _, noun := range NOUNS {
		_, found := nounsHash[noun]
		require.False(t, found, "Duplicate Error: found %v twice in NOUNS", noun)
		nounsHash[noun] = true
	}

	// this will only be called if there are no duplicates
	require.True(t, true)
}
