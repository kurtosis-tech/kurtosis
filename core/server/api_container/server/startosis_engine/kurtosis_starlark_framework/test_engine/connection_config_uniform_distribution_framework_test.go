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

type connectionConfigUniformDistributionTestCase struct {
	*testing.T
}

func newConnectionConfigUniformDistributionTestCase(t *testing.T) *connectionConfigUniformDistributionTestCase {
	return &connectionConfigUniformDistributionTestCase{
		T: t,
	}
}

func (t *connectionConfigUniformDistributionTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", connection_config.ConnectionConfigTypeName, "NormalDistribution")
}

func (t *connectionConfigUniformDistributionTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return connection_config.NewConnectionConfigType()
}

func (t *connectionConfigUniformDistributionTestCase) GetStarlarkCode() string {
	uniformDistribution := fmt.Sprintf("%s(%s=%d)", packet_delay_distribution.UniformPacketDelayDistributionTypeName, packet_delay_distribution.DelayAttr, 50)
	return fmt.Sprintf("%s(%s=%s, %s=%s)", connection_config.ConnectionConfigTypeName, connection_config.PacketLossPercentageAttr, "50.0", connection_config.PacketDelayDistributionAttr, uniformDistribution)
}

func (t *connectionConfigUniformDistributionTestCase) Assert(typeValue starlark.Value) {
	connectionConfigStarlark, ok := typeValue.(*connection_config.ConnectionConfig)
	require.True(t, ok)
	connectionConfig, err := connectionConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedConnectionConfig := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(50),
		partition_topology.NewUniformPacketDelayDistribution(50))
	require.Equal(t, expectedConnectionConfig, *connectionConfig)
}
