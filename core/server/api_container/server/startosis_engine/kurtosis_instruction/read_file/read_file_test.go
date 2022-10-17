package read_file

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadFile_StringRepresentation(t *testing.T) {
	testInstruction := NewReadFileInstruction(
		*kurtosis_instruction.NewInstructionPosition(3, 33),
		"path/to/file.star",
	)
	expectedStr := `read_file(src_path="path/to/file.star")`
	require.Equal(t, expectedStr, testInstruction.String())
	require.Equal(t, expectedStr, testInstruction.GetCanonicalInstruction())
}
