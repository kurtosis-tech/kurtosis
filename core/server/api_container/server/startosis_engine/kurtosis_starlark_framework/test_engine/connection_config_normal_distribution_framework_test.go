package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/connection_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/packet_delay_distribution"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type connectionConfigNormalDistributionTestCase struct {
	*testing.T
}

func newConnectionConfigNormalDistributionTestCase(t *testing.T) *connectionConfigNormalDistributionTestCase {
	return &connectionConfigNormalDistributionTestCase{
		T: t,
	}
}

func (t *connectionConfigNormalDistributionTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", connection_config.ConnectionConfigTypeName, "NormalDistribution")
}

func (t *connectionConfigNormalDistributionTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return connection_config.NewConnectionConfigType()
}

func (t *connectionConfigNormalDistributionTestCase) GetStarlarkCode() string {
	normalDistribution := fmt.Sprintf("%s(%s=%d, %s=%d)", packet_delay_distribution.NormalPacketDelayDistributionTypeName, packet_delay_distribution.MeanAttr, 50, packet_delay_distribution.StdDevAttr, 5)
	return fmt.Sprintf("%s(%s=%s, %s=%s)", connection_config.ConnectionConfigTypeName, connection_config.PacketLossPercentageAttr, "50.0", connection_config.PacketDelayDistributionAttr, normalDistribution)
}

func (t *connectionConfigNormalDistributionTestCase) Assert(typeValue starlark.Value) {
	connectionConfigStarlark, ok := typeValue.(*connection_config.ConnectionConfig)
	require.True(t, ok)
	connectionConfig, err := connectionConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedConnectionConfig := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(50),
		partition_topology.NewNormalPacketDelayDistribution(50, 5, 0))
	require.Equal(t, expectedConnectionConfig, *connectionConfig)
}
