package label_key_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/stretchr/testify/require"
	"testing"
)

//We expect these strings to be reliable between versions.
const (
	expectedLabelKeyPrefixStr 		= "kurtosistech.com/"
	expectedAppIdLabelKeyStr 		= "kurtosistech.com/app-id"
	expectedResourceTypeLabelKeyStr = "kurtosistech.com/resource-type"
)

//When Kurtosis versions change, these particular label keys must be equal.
var crossVersionLabelKeyStringsToEnsure = map[string]string{
	labelKeyPrefixStr:       	expectedLabelKeyPrefixStr,
	appIdLabelKeyStr:        	expectedAppIdLabelKeyStr,
	resourceTypeLabelKeyStr: 	expectedResourceTypeLabelKeyStr,
}

//These are the publicly accessible keys that correspond to the private string constants. They need to stay the same.
var crossVersionLabelKeysToEnsure = map[*kubernetes_label_key.KubernetesLabelKey]string{
	AppIDLabelKey:                	expectedAppIdLabelKeyStr,
	KurtosisResourceTypeLabelKey: 	expectedResourceTypeLabelKeyStr,
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It is VERY important that certain constants don't get modified, else Kurtosis will silently lose track
// of preexisting resources (thereby causing a resource leak). This test ensures that certain constants
// are never modified.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func TestImmutableConstantsArentModified(t *testing.T) {
	for actualValue, expectedValue := range crossVersionLabelKeyStringsToEnsure {
		require.Equal(t, expectedValue, actualValue, "An immutable label key string was modified! Got '%v' but should be '%v'", actualValue, expectedValue)
	}

	for labelKey, expectedValueStr := range crossVersionLabelKeysToEnsure {
		labelKeyStr := labelKey.GetString()
		require.Equal(t, expectedValueStr, labelKeyStr, "An immutable label key was modified! Got '%v' but should be '%v'", labelKeyStr, expectedValueStr)
	}
}
