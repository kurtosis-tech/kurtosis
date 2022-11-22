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
	)
	expectedMultiLineStr := `# from: dummyFile[1:1]
print(
	"foo",
	[
		"bar"
	],
	end="EOL",
	sep="; "
)`
	require.Equal(t, expectedMultiLineStr, instruction.GetCanonicalInstruction())
	expectedSingleLineStr := `print("foo", ["bar"], end="EOL", sep="; ")`
	require.Equal(t, expectedSingleLineStr, instruction.String())
}
