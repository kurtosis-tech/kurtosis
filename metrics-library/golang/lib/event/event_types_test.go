package event

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	someValueToTestHash     = "some value some value"
	expectedHashedSomeValue = "f9b668fec50b7eebb1319e0a2cce08ee56e5b123ace456428342d4b60dc19968"

	anotherSentenceToTestHash = "It's not a simple random text."
	expectedHashedSentence    = "029f21eeab056e2e98933d2b872d2ff365e5e837876534227d534fcdba248ac7"
)

func TestHashString(t *testing.T) {
	hashedValue := hashString(someValueToTestHash)

	require.Equal(t, expectedHashedSomeValue, hashedValue)

	hashedSentence := hashString(anotherSentenceToTestHash)

	require.Equal(t, expectedHashedSentence, hashedSentence)
}
