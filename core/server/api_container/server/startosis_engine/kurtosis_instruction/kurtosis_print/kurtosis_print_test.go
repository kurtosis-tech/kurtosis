package kurtosis_print

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestPrintInstruction_StringRepresentation(t *testing.T) {
	instruction := NewPrintInstruction(
		kurtosis_instruction.NewInstructionPosition(1, 1, "dummyFile"),
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
	require.Equal(t, expectedStr, instruction.GetCanonicalInstruction())
	require.Equal(t, expectedStr, instruction.String())
}
