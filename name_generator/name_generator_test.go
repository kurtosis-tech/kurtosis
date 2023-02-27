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
	adjectivesHashMap := map[string]int{}

	for _, adj := range ADJECTIVES {
		freq, found := adjectivesHashMap[adj]
		if found && freq == 1 {
			duplicateArr = append(duplicateArr, adj)
		}

		adjectivesHashMap[adj] = adjectivesHashMap[adj] + 1
	}

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in ADJECTIVES", duplicateArr)
}

// check for duplicates for nouns
func Test_noDuplicatesInNouns(t *testing.T) {
	var duplicateArr []string
	nounsHashMap := map[string]int{}

	for _, noun := range NOUNS {
		freq, found := nounsHashMap[noun]
		if found && freq == 1 {
			duplicateArr = append(duplicateArr, noun)
		}

		nounsHashMap[noun] = nounsHashMap[noun] + 1
	}

	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in NOUNS", duplicateArr)
}
