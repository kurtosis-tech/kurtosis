package name_generator

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRandomNameGenerator_GetNameGenerator(t *testing.T) {
	nameGenerator := getNameGenerator()
	nameGenerator2 := getNameGenerator()
	// testing the memory address equality
	require.True(t, nameGenerator == nameGenerator2)
}

func TestRandomNameGenerator_GenerateName(t *testing.T) {
	args := generatorArgs{
		adjectives: []string{"test_adj", "test_adj_two"},
		nouns:      []string{"noun"},
	}

	nameGenerator := getNameGenerator()
	actual := nameGenerator.generateName(args)
	potentialCandidates := map[string]bool{
		"test_adj-noun":     true,
		"test_adj_two-noun": true,
	}

	require.Contains(t, potentialCandidates, actual)
}

// check for duplicates for adjectives
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

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in ADJECTIVES", duplicateArr)
}

// check for duplicates for nouns
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

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in NOUNS", duplicateArr)
}
