package read_file

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	filePath = "github.com/foo/bar/file.star"
)

func TestReadFile_StringRepresentation(t *testing.T) {
	testInstruction := NewReadFileInstruction(
		*kurtosis_instruction.NewInstructionPosition(3, 33),
		filePath,
	)
	expectedStr := `read_file(src_path="` + filePath + `")`
	require.Equal(t, expectedStr, testInstruction.String())
	require.Equal(t, expectedStr, testInstruction.GetCanonicalInstruction())
}
