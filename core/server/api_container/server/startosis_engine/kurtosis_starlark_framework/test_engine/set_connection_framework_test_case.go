package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/set_connection"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	setConnection_subnetwork1 = service_network_types.PartitionID("subnetwork_1")
	setConnection_subnetwork2 = service_network_types.PartitionID("subnetwork_2")
)

type setConnectionTestCase struct {
	*testing.T
}

func newSetConnectionTestCase(t *testing.T) *setConnectionTestCase {
	return &setConnectionTestCase{
		T: t,
	}
}

func (t *setConnectionTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", set_connection.SetConnectionBuiltinName, "BetweenTwoSubnetworks")
}

func (t *setConnectionTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().SetConnection(
		mock.Anything,
		setConnection_subnetwork1,
		setConnection_subnetwork2,
		mock.MatchedBy(func(actualPartitionConnection partition_topology.PartitionConnection) bool {
			expectedPartitionConnection := partition_topology.NewPartitionConnection(
				partition_topology.NewPacketLoss(50),
				partition_topology.NewPacketDelay(100))
			assert.Equal(t, expectedPartitionConnection, actualPartitionConnection)
			return true
		}),
	).Times(1).Return(
		nil,
	)

	return set_connection.NewSetConnection(serviceNetwork)
}

func (t *setConnectionTestCase) GetStarlarkCode() string {
	connectionConfig := "ConnectionConfig(packet_loss_percentage=50.0, packet_delay=PacketDelay(delay_ms=100))"
	subnetworks := fmt.Sprintf(`(%q, %q)`, setConnection_subnetwork1, setConnection_subnetwork2)
	return fmt.Sprintf("%s(%s=%s, %s=%s)", set_connection.SetConnectionBuiltinName, set_connection.SubnetworksArgName, subnetworks, set_connection.ConnectionConfigArgName, connectionConfig)
}

func (t *setConnectionTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)
	require.Equal(t, "Configured subnetwork connection between 'subnetwork_1' and 'subnetwork_2'", *executionResult)
}
