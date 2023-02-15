package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/set_connection"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

type setConnectionDefaultTestCase struct {
	*testing.T
}

func newSetConnectionDefaultTestCase(t *testing.T) *setConnectionDefaultTestCase {
	return &setConnectionDefaultTestCase{
		T: t,
	}
}

func (t setConnectionDefaultTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", set_connection.SetConnectionBuiltinName, "DefaultConnection")
}

func (t setConnectionDefaultTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().SetDefaultConnection(
		mock.Anything,
		mock.MatchedBy(func(actualPartitionConnection partition_topology.PartitionConnection) bool {
			expectedPartitionConnection := partition_topology.NewPartitionConnection(
				partition_topology.NewPacketLoss(50),
				partition_topology.NewUniformPacketDelayDistribution(100))
			assert.Equal(t, expectedPartitionConnection, actualPartitionConnection)
			return true
		}),
	).Times(1).Return(
		nil,
	)

	return set_connection.NewSetConnection(serviceNetwork)
}

func (t setConnectionDefaultTestCase) GetStarlarkCode() string {
	connectionConfig := "ConnectionConfig(packet_loss_percentage=50.0, packet_delay=PacketDelay(delay_ms=100))"
	return fmt.Sprintf("%s(%s=%s)", set_connection.SetConnectionBuiltinName, set_connection.ConnectionConfigArgName, connectionConfig)
}

func (t setConnectionDefaultTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)
	require.Equal(t, "Configured default subnetwork connection", *executionResult)
}
