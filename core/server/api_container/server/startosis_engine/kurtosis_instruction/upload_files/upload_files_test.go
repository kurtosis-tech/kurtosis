package upload_files

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestUploadFiles_StringRepresentation(t *testing.T) {
	position := kurtosis_instruction.NewInstructionPosition(1, 13, "dummyFile")
	filePath := "github.com/kurtosis/module/lib/lib.star"
	artifactName := "test-artifact"
	uploadInstruction := newEmptyUploadFilesInstruction(position, nil, nil)
	uploadInstruction.starlarkKwargs = starlark.StringDict{}
	uploadInstruction.starlarkKwargs[srcArgName] = starlark.String(filePath)
	uploadInstruction.starlarkKwargs[nonOptionalArtifactNameArgName] = starlark.String(artifactName)

	expectedStr := `upload_files(name="` + artifactName + `", src="` + filePath + `")`
	require.Equal(t, expectedStr, uploadInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		UploadFilesBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"`+filePath+`"`, srcArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(`"`+artifactName+`"`, nonOptionalArtifactNameArgName, true),
		})
	require.Equal(t, canonicalInstruction, uploadInstruction.GetCanonicalInstruction())
}
