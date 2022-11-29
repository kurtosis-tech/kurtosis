package kurtosis_print

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestPrintInstruction_StringRepresentation(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile")
	instruction := NewPrintInstruction(
		position,
		[]starlark.Value{
			starlark.String("foo"),
			starlark.NewList([]starlark.Value{
				starlark.String("bar"),
			}),
		},
		"; ",
		"EOL",
		nil,
	)
	expectedStr := `print("foo", ["bar"], end="EOL", sep="; ")`
	require.Equal(t, expectedStr, instruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		PrintBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionArg(`"foo"`, true),
			binding_constructors.NewStarlarkInstructionArg(`["bar"]`, true),
			binding_constructors.NewStarlarkInstructionKwarg(`"; "`, "sep", false),
			binding_constructors.NewStarlarkInstructionKwarg(`"EOL"`, "end", false),
		})
	require.Equal(t, canonicalInstruction, instruction.GetCanonicalInstruction())
}
