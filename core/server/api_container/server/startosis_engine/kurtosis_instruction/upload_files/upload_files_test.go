package upload_files

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUploadFiles_StringRepresentation(t *testing.T) {
	filePath := "github.com/kurtosis/module/lib/lib.star"
	uploadInstruction := NewUploadFilesInstruction(
		*kurtosis_instruction.NewInstructionPosition(1, 13, "dummyFile"),
		nil, nil, filePath,
	)
	expectedStrRep := `upload_files(src_path="` + filePath + `")`
	require.Equal(t, expectedStrRep, uploadInstruction.String())
	require.Equal(t, expectedStrRep, uploadInstruction.GetCanonicalInstruction())
}
