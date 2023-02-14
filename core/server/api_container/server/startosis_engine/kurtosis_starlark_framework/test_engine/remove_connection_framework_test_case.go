package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_connection"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	removeConnection_subnetwork1 = service_network_types.PartitionID("subnetwork_1")
	removeConnection_subnetwork2 = service_network_types.PartitionID("subnetwork_2")
)

type removeConnectionTestCase struct {
	*testing.T
}

func newRemoveConnectionTestCase(t *testing.T) *removeConnectionTestCase {
	return &removeConnectionTestCase{
		T: t,
	}
}

func (t *removeConnectionTestCase) GetId() string {
	return remove_connection.RemoveConnectionBuiltinName
}

func (t *removeConnectionTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	serviceNetwork.EXPECT().UnsetConnection(
		mock.Anything,
		removeConnection_subnetwork1,
		removeConnection_subnetwork2,
	).Times(1).Return(nil)
	return remove_connection.NewRemoveConnection(serviceNetwork)
}

func (t *removeConnectionTestCase) GetStarlarkCode() string {
	subnetworks := fmt.Sprintf("(%q, %q)", removeConnection_subnetwork1, removeConnection_subnetwork2)
	return fmt.Sprintf("%s(%s=%s)", remove_connection.RemoveConnectionBuiltinName, remove_connection.SubnetworksArgName, subnetworks)
}

func (t *removeConnectionTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.None, interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Removed subnetwork connection override between '%s' and '%s'", removeConnection_subnetwork1, removeConnection_subnetwork2)
	require.Equal(t, expectedExecutionResult, *executionResult)
}
