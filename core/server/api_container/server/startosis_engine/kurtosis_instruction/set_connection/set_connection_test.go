package set_connection

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	thread                                               = shared_helpers.NewStarlarkThread("test-set-connection")
	packetConnectionPercentageValueForUnblockedPartition = partition_topology.NewPacketLoss(0)
	packetConnectionPercentageValueForSoftPartition      = partition_topology.NewPacketLoss(50)
	packetConnectionPercentageValueForBlockedPartition   = partition_topology.NewPacketLoss(100)
)

func TestSetConnection_Interpreter(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `set_connection(("subnetwork1", "subnetwork2"), ConnectionConfig(50.0))`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	subnetwork1 := service_network_types.PartitionID("subnetwork1")
	subnetwork2 := service_network_types.PartitionID("subnetwork2")
	expectedInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		&subnetwork1,
		&subnetwork2,
		partition_topology.NewPartitionConnection(packetConnectionPercentageValueForSoftPartition, partition_topology.ConnectionWithNoPacketDelay),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay),
			"subnetworks": starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestSetConnection_InterpreterWithNamedArgs(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `set_connection(("subnetwork1", "subnetwork2"), 
	ConnectionConfig(
		packet_delay=PacketDelay(delay_ms=100), 
		packet_loss_percentage=50.0
	)
)`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		kurtosis_types.PacketDelayName:          starlark.NewBuiltin(kurtosis_types.PacketDelayName, kurtosis_types.MakePacketDelay),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	subnetwork1 := service_network_types.PartitionID("subnetwork1")
	subnetwork2 := service_network_types.PartitionID("subnetwork2")
	expectedInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		&subnetwork1,
		&subnetwork2,
		partition_topology.NewPartitionConnection(
			packetConnectionPercentageValueForSoftPartition,
			partition_topology.NewPacketDelay(100),
		),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(50, kurtosis_types.NewPacketDelay(100)),
			"subnetworks": starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestSetConnection_InterpreterWithDefaultNamedArgs(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `set_connection(("subnetwork1", "subnetwork2"), ConnectionConfig())`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		kurtosis_types.PacketDelayName:          starlark.NewBuiltin(kurtosis_types.PacketDelayName, kurtosis_types.MakePacketDelay),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	subnetwork1 := service_network_types.PartitionID("subnetwork1")
	subnetwork2 := service_network_types.PartitionID("subnetwork2")
	expectedInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		&subnetwork1,
		&subnetwork2,
		partition_topology.NewPartitionConnection(
			packetConnectionPercentageValueForUnblockedPartition,
			partition_topology.NewPacketDelay(0),
		),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(0, kurtosis_types.NewPacketDelay(0)),
			"subnetworks": starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestSetConnection_Interpreter_SetDefaultConnection(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `set_connection(ConnectionConfig(50.0))`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	expectedInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		nil,
		nil,
		partition_topology.NewPartitionConnection(packetConnectionPercentageValueForSoftPartition, partition_topology.ConnectionWithNoPacketDelay),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestSetConnection_Interpreter_SetDefaultConnection_PreBuiltConnections(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `set_connection(kurtosis.connection.BLOCKED)`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		builtins.KurtosisModuleName: builtins.KurtosisModule(),
		SetConnectionBuiltinName:    starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	expectedInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		nil,
		nil,
		partition_topology.NewPartitionConnection(packetConnectionPercentageValueForBlockedPartition, partition_topology.ConnectionWithNoPacketDelay),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(100, kurtosis_types.NoPacketDelay),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestSetConnection_GetCanonicalizedInstruction(t *testing.T) {
	subnetwork1 := "subnetwork1"
	subnetwork2 := "subnetwork2"

	connectionConfig := kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay)
	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(subnetwork1),
		starlark.String(subnetwork2),
	})
	setConnectionInstruction := newEmptySetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
	)
	setConnectionInstruction.starlarkKwargs[subnetworksArgName] = subnetworks
	setConnectionInstruction.starlarkKwargs[connectionConfigArgName] = connectionConfig

	expectedOutput := `set_connection(config=ConnectionConfig(packet_loss_percentage=50.0, packet_delay=PacketDelay(delay_ms=0)), subnetworks=("subnetwork1", "subnetwork2"))`
	require.Equal(t, expectedOutput, setConnectionInstruction.String())
}

func TestSetConnection_GetCanonicalizedInstruction_NoSubnetworks(t *testing.T) {
	connectionConfig := kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay)
	setConnectionInstruction := newEmptySetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
	)
	setConnectionInstruction.starlarkKwargs[connectionConfigArgName] = connectionConfig

	expectedOutput := `set_connection(config=ConnectionConfig(packet_loss_percentage=50.0, packet_delay=PacketDelay(delay_ms=0)))`
	require.Equal(t, expectedOutput, setConnectionInstruction.String())
}

func TestSetConnection_SerializeAndParseAgain(t *testing.T) {
	subnetwork1 := service_network_types.PartitionID("subnetwork1")
	subnetwork2 := service_network_types.PartitionID("subnetwork2")
	initialInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		&subnetwork1,
		&subnetwork2,
		partition_topology.NewPartitionConnection(packetConnectionPercentageValueForSoftPartition, partition_topology.ConnectionWithNoPacketDelay),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay),
			"subnetworks": starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})

	canonicalizedInstruction := initialInstruction.String()

	var instructions []kurtosis_instruction.KurtosisInstruction
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, canonicalizedInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		kurtosis_types.PacketDelayName:          starlark.NewBuiltin(kurtosis_types.PacketDelayName, kurtosis_types.MakePacketDelay),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	require.Equal(t, initialInstruction, instructions[0])
}

func TestSetConnection_SerializeAndParseAgain_DefaultConnection(t *testing.T) {
	initialInstruction := NewSetConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		nil,
		nil,
		partition_topology.NewPartitionConnection(packetConnectionPercentageValueForSoftPartition, partition_topology.ConnectionWithNoPacketDelay),
		starlark.StringDict{
			"config": kurtosis_types.NewConnectionConfig(50, kurtosis_types.NoPacketDelay),
		})

	canonicalizedInstruction := initialInstruction.String()

	var instructions []kurtosis_instruction.KurtosisInstruction
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, canonicalizedInstruction, starlark.StringDict{
		kurtosis_types.ConnectionConfigTypeName: starlark.NewBuiltin(kurtosis_types.ConnectionConfigTypeName, kurtosis_types.MakeConnectionConfig),
		kurtosis_types.PacketDelayName:          starlark.NewBuiltin(kurtosis_types.PacketDelayName, kurtosis_types.MakePacketDelay),
		SetConnectionBuiltinName:                starlark.NewBuiltin(SetConnectionBuiltinName, GenerateSetConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	require.Equal(t, initialInstruction, instructions[0])
}
