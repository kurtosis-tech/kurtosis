package resolved_config

import (
	v1 "github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_config/v1"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetKurtosisBackendSupplier(t *testing.T) {
	for _, clusterType := range KurtosisClusterTypeValues() {
		// Set the config appropriately so that we pass validation
		var config *v1.KubernetesClusterConfigV1
		switch clusterType {
		case KurtosisClusterType_Docker:
			config = nil
		case KurtosisClusterType_Kubernetes:
			clusterName := "test"
			storageClass := "standard"
			enclaveSizeInGb := uint(2)
			config = &v1.KubernetesClusterConfigV1{
				KubernetesClusterName:  &clusterName,
				StorageClass:           &storageClass,
				EnclaveSizeInGigabytes: &enclaveSizeInGb,
			}
		}

		_, err := getKurtosisBackendSupplier(clusterType, config)
		require.NoError(t, err)
	}
}

func TestNewKurtosisConfigFromRequiredFields_MetricsElectionIsSent(t *testing.T) {
	config, err := NewKurtosisConfigFromRequiredFields(false)
	require.NoError(t, err)

	overrides := config.GetOverrides()
	require.NotNil(t, overrides.ShouldSendMetrics)
}

func TestNewKurtosisConfigEmptyOverrides(t *testing.T) {
	_, err := NewKurtosisConfigFromOverrides(&v1.KurtosisConfigV1{})
	// You can not initialize a Kurtosis config with empty overrides - it needs at least `ShouldSendMetrics`
	require.Error(t, err)
}

