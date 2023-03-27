package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/connection_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/packet_delay_distribution"
	"github.com/stretchr/testify/require"
	"testing"
)

type connectionConfigFullTestCase struct {
	*testing.T
}

func newConnectionConfigFullTestCase(t *testing.T) *connectionConfigFullTestCase {
	return &connectionConfigFullTestCase{
		T: t,
	}
}

func (t *connectionConfigFullTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", connection_config.ConnectionConfigTypeName, "Full")
}

func (t *connectionConfigFullTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return connection_config.NewConnectionConfigType()
}

func (t *connectionConfigFullTestCase) GetStarlarkCode() string {
	packetDelayArg := fmt.Sprintf("%s(%s=%d, %s=%d)", packet_delay_distribution.NormalPacketDelayDistributionTypeName, packet_delay_distribution.MeanAttr, 100, packet_delay_distribution.StdDevAttr, 10)
	return fmt.Sprintf("%s(%s=%s, %s=%s)", connection_config.ConnectionConfigTypeName, connection_config.PacketLossPercentageAttr, "50.0", connection_config.PacketDelayDistributionAttr, packetDelayArg)
}

func (t *connectionConfigFullTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	connectionConfigStarlark, ok := typeValue.(*connection_config.ConnectionConfig)
	require.True(t, ok)
	connectionConfig, err := connectionConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedConnectionConfig := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(50),
		partition_topology.NewNormalPacketDelayDistribution(100, 10, 0))
	require.Equal(t, expectedConnectionConfig, *connectionConfig)
}
