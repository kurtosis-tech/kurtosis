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
	duplicateArr := checkIfDuplicateExistsInArray(ADJECTIVES)
	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in ADJECTIVES", duplicateArr)
}

// check for duplicates for nouns
func Test_noDuplicatesInNouns(t *testing.T) {
	// Engine Nouns
	duplicateArr := checkIfDuplicateExistsInArray(ENGINE_NOUNS)
	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in ENGINE_NOUNS with len: %v", duplicateArr, len(duplicateArr))

	// File Artifact Nouns
	duplicateArr = checkIfDuplicateExistsInArray(FILE_ARTIFACT_NOUNS)
	require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in FILE_ARTIFACT_NOUNS with len: %v", duplicateArr, len(duplicateArr))
}

// return an array containing duplicates
// if empty array is returned that means no duplicates exists
func checkIfDuplicateExistsInArray(arrayUnderTest []string) []string {
	var duplicateArr []string
	frequencyMap := map[string]int{}

	for _, noun := range arrayUnderTest {
		freq, found := frequencyMap[noun]
		if found && freq == 1 {
			duplicateArr = append(duplicateArr, noun)
		}

		frequencyMap[noun] = frequencyMap[noun] + 1
	}

	return duplicateArr
}
