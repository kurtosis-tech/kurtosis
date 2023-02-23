package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// check for duplicates for nouns and adjectives
func Test_noDuplicatesInAdjectives(t *testing.T) {
	var duplicateArr []string
	adjectivesHash := map[string]int{}

	for _, adj := range NOUNS {
		freq, found := adjectivesHash[adj]
		if found && freq == 1 {
			duplicateArr = append(duplicateArr, adj)
		}

		adjectivesHash[adj] = adjectivesHash[adj] + 1
	}

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v twice in ADJECTIVES", duplicateArr)
}
