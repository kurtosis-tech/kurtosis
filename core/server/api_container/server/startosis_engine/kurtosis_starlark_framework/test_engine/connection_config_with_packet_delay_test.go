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

type connectionConfigWithPacketDelayTestCase struct {
	*testing.T
}

func newConnectionConfigWithPacketDelayTestCase(t *testing.T) *connectionConfigWithPacketDelayTestCase {
	return &connectionConfigWithPacketDelayTestCase{
		T: t,
	}
}

func (t *connectionConfigWithPacketDelayTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", connection_config.ConnectionConfigTypeName, "WithPacketDelay")
}

func (t *connectionConfigWithPacketDelayTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return connection_config.NewConnectionConfigType()
}

func (t *connectionConfigWithPacketDelayTestCase) GetStarlarkCode() string {
	packetDelayArg := fmt.Sprintf("%s(%s=%d)", packet_delay_distribution.UniformPacketDelayDistributionTypeName, packet_delay_distribution.DelayAttr, 10)
	return fmt.Sprintf("%s(%s=%s)", connection_config.ConnectionConfigTypeName, connection_config.PacketDelayDistributionAttr, packetDelayArg)
}

func (t *connectionConfigWithPacketDelayTestCase) Assert(typeValue starlark.Value) {
	connectionConfigStarlark, ok := typeValue.(*connection_config.ConnectionConfig)
	require.True(t, ok)
	connectionConfig, err := connectionConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedConnectionConfig := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(0),
		partition_topology.NewUniformPacketDelayDistribution(10))
	require.Equal(t, expectedConnectionConfig, *connectionConfig)
}
