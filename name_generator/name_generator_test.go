package name_generator

import (
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

var allowedCharacters = regexp.MustCompile("^[a-z]+$")

var listsToVerify = map[string][]string{
	"adjectives":           ADJECTIVES,
	"enclave nouns":        ENCLAVE_NOUNS,
	"files artifact nouns": FILE_ARTIFACT_NOUNS,
}

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

func Test_noDuplicates(t *testing.T) {
	for listDescription, list := range listsToVerify {
		duplicateArr := checkIfDuplicateExistsInArray(list)
		require.Len(t, duplicateArr, 0, "Duplicate Error: found %v multiple times in list '%v'", duplicateArr, listDescription)
	}
}

func Test_onlyAllowedCharacters(t *testing.T) {
	for listDescription, list := range listsToVerify {
		for _, item := range list {
			require.True(
				t,
				allowedCharacters.MatchString(item),
				"Item '%v' in list '%v' doesn't match allowed item regex",
				item,
				listDescription,
			)
		}
	}
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
