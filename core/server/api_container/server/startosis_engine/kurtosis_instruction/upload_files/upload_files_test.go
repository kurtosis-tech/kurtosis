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
		nil, nil, filePath, "dummyPathOnDisk",
	)
	expectedMultiLineStrRep := `# from: dummyFile[1:13]
upload_files(
	src_path="` + filePath + `"
)`
	require.Equal(t, expectedMultiLineStrRep, uploadInstruction.GetCanonicalInstruction())
	expectedSingleLineStrRep := `upload_files(src_path="` + filePath + `")`
	require.Equal(t, expectedSingleLineStrRep, uploadInstruction.String())
}
