package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/connection_config"
	"github.com/stretchr/testify/require"
	"testing"
)

type connectionConfigWithPacketLossTestCase struct {
	*testing.T
}

func newConnectionConfigWithPacketLossTestCase(t *testing.T) *connectionConfigWithPacketLossTestCase {
	return &connectionConfigWithPacketLossTestCase{
		T: t,
	}
}

func (t *connectionConfigWithPacketLossTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", connection_config.ConnectionConfigTypeName, "WithPacketLoss")
}

func (t *connectionConfigWithPacketLossTestCase) GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return connection_config.NewConnectionConfigType()
}

func (t *connectionConfigWithPacketLossTestCase) GetStarlarkCode() string {
	return fmt.Sprintf("%s(%s=%s)", connection_config.ConnectionConfigTypeName, connection_config.PacketLossPercentageAttr, "50.0")
}

func (t *connectionConfigWithPacketLossTestCase) Assert(typeValue builtin_argument.KurtosisValueType) {
	connectionConfigStarlark, ok := typeValue.(*connection_config.ConnectionConfig)
	require.True(t, ok)
	connectionConfig, err := connectionConfigStarlark.ToKurtosisType()
	require.Nil(t, err)

	expectedConnectionConfig := partition_topology.NewPartitionConnection(
		partition_topology.NewPacketLoss(50),
		partition_topology.NewUniformPacketDelayDistribution(0))
	require.Equal(t, expectedConnectionConfig, *connectionConfig)
}
