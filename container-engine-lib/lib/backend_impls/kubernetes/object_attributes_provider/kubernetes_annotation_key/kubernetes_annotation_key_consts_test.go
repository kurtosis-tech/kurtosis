package kubernetes_annotation_key

import (
	"github.com/stretchr/testify/require"
	"testing"
)

//We expect these strings to be reliable between versions.
const (
	expectedAnnotationKeyNamespaceStr = "com.kurtosistech."
)

//When Kurtosis versions change, these particular annotation key strings must be equal.
//If these change between versions, Kurtosis will not be able to find and manage resources with these annotation keys.
//They will effectively be lost to Kurtosis and the user will have to clean up any mess.
var crossVersionLabelKeyStringsToEnsure = map[string]string{
	keyNamespaceStr: expectedAnnotationKeyNamespaceStr,
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It is VERY important that certain constants don't get modified, else Kurtosis will silently lose track
// of preexisting resources (thereby causing a resource leak). This test ensures that certain constants
// are never modified.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func TestImmutableConstantsArentModified(t *testing.T) {
	for actualValue, expectedValue := range crossVersionLabelKeyStringsToEnsure {
		require.Equal(t, expectedValue, actualValue, "An immutable annotation key string was modified! Got '%v' but should be '%v'", actualValue, expectedValue)
	}
}
