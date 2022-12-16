package remove_connection

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	thread = shared_helpers.NewStarlarkThread("test-remove-connection")
)

func TestRemoveService_Interpreter(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `remove_connection(("subnetwork1", "subnetwork2"))`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		RemoveConnectionBuiltinName: starlark.NewBuiltin(RemoveConnectionBuiltinName, GenerateRemoveConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	expectedInstruction := NewRemoveConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 18, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		"subnetwork1",
		"subnetwork2",
		starlark.StringDict{
			subnetworksArgName: starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestRemoveService_CanonicalRepresentation(t *testing.T) {
	subnetwork1 := "subnetwork1"
	subnetwork2 := "subnetwork2"

	subnetworks := starlark.Tuple([]starlark.Value{
		starlark.String(subnetwork1),
		starlark.String(subnetwork2),
	})
	setConnectionInstruction := newEmptyRemoveConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
	)
	setConnectionInstruction.starlarkKwargs[subnetworksArgName] = subnetworks

	expectedOutput := `remove_connection(subnetworks=("subnetwork1", "subnetwork2"))`
	require.Equal(t, expectedOutput, setConnectionInstruction.String())
}

func TestRemoveService_SerializeAndParseAgain(t *testing.T) {
	initialInstruction := NewRemoveConnectionInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 18, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		"subnetwork1",
		"subnetwork2",
		starlark.StringDict{
			subnetworksArgName: starlark.Tuple([]starlark.Value{
				starlark.String("subnetwork1"),
				starlark.String("subnetwork2"),
			}),
		})

	canonicalizedInstruction := initialInstruction.String()

	var instructions []kurtosis_instruction.KurtosisInstruction
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, canonicalizedInstruction, starlark.StringDict{
		RemoveConnectionBuiltinName: starlark.NewBuiltin(RemoveConnectionBuiltinName, GenerateRemoveConnectionBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	require.Equal(t, initialInstruction, instructions[0])
}
