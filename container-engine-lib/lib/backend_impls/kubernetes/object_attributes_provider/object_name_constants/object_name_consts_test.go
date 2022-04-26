package object_name_constants

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_object_name"
	"github.com/stretchr/testify/require"
	"testing"
)

var labelValueStrsToEnsure = map[string]string{
	kurtosisInternalContainerGrpcPortId:      "grpc",
	kurtosisInternalContainerGrpcProxyPortId: "grpc-proxy",
}

var labelValuesToEnsure = map[*kubernetes_object_name.KubernetesObjectName]string{
	KurtosisInternalContainerGrpcPortName:      "grpc",
	KurtosisInternalContainerGrpcProxyPortName: "grpc-proxy",
}

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// It is VERY important that certain constants don't get modified, else Kurtosis will silently lose track
// of preexisting resources (thereby causing a resource leak). This test ensures that certain constants
// are never modified.
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! IMPORTANT !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func TestImmutableConstantsArentModified(t *testing.T) {
	for actualValue, expectedValue := range labelValueStrsToEnsure {
		require.Equal(t, expectedValue, actualValue, "An immutable label value string was modified! Got '%v' but should be '%v'", actualValue, expectedValue)
	}

	for labelKey, expectedValueStr := range labelValuesToEnsure {
		labelKeyStr := labelKey.GetString()
		require.Equal(t, expectedValueStr, labelKeyStr, "An immutable label value was modified! Got '%v' but should be '%v'", labelKeyStr, expectedValueStr)
	}
}
