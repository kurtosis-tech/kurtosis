package snapshots

import (
	"testing"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
)

func TestSnapshotCreator_CreateSnapshot(t *testing.T) {
	serviceConfig := service.GetEmptyServiceConfig()

	apiServiceConfig, err := convertServiceConfigToJsonServiceConfig(serviceConfig)
	require.NoError(t, err)

	require.Nil(t, apiServiceConfig)
}
