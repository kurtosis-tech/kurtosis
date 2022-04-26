package label_key_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_label_key"
	"github.com/stretchr/testify/require"
	"testing"
)

var labelKeyStrsToEnsure = map[string]string{
	labelKeyPrefixStr:       "kurtosistech.com/",
	appIdLabelKeyStr:        "kurtosistech.com/app-id",
	resourceTypeLabelKeyStr: "kurtosistech.com/resource-type",
}

var labelKeysToEnsure = map[*kubernetes_label_key.KubernetesLabelKey]string{
	AppIDLabelKey:        "kurtosistech.com/app-id",
	ResourceTypeLabelKey: "kurtosistech.com/resource-type",
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It is VERY important that certain constants don't get modified, else Kurtosis will silently lose track
// of preexisting resources (thereby causing a resource leak). This test ensures that certain constants
// are never modified.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func TestImmutableConstantsArentModified(t *testing.T) {
	for actualValue, expectedValue := range labelKeyStrsToEnsure {
		require.Equal(t, expectedValue, actualValue, "An immutable label key string was modified! Got '%v' but should be '%v'", actualValue, expectedValue)
	}

	for labelKey, expectedValueStr := range labelKeysToEnsure {
		labelKeyStr := labelKey.GetString()
		require.Equal(t, expectedValueStr, labelKeyStr, "An immutable label key was modified! Got '%v' but should be '%v'", labelKeyStr, expectedValueStr)
	}
}
