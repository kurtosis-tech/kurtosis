package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// check for duplicates for nouns and adjectives
func Test_noDuplicatesInNouns(t *testing.T) {
	var duplicateArr []string
	nounsHash := map[string]int{}

	for _, noun := range NOUNS {
		freq, found := nounsHash[noun]
		if found && freq == 1 {
			duplicateArr = append(duplicateArr, noun)
		}

		nounsHash[noun] = nounsHash[noun] + 1
	}

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v occurred twice in NOUNS", duplicateArr)
}
