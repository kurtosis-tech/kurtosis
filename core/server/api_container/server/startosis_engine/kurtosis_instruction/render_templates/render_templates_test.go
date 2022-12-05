package render_templates

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"testing"
)

func TestRenderTemplate_TestStringRepresentation(t *testing.T) {
	template := "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
	templateData := map[string]interface{}{"Name": "Stranger", "Answer": 6, "Numbers": []int{1, 2, 3}, "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43}
	templateDataAsJson, err := json.Marshal(templateData)
	require.Nil(t, err)
	templateAndDataDict := &starlark.Dict{}
	templateStrDict := starlark.StringDict{}
	templateStrDict["template"] = starlark.String(template)
	templateStrDict["data"] = starlark.String(templateDataAsJson)
	require.Nil(t, templateAndDataDict.SetKey(starlark.String("/foo/bar/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateStrDict)))

	position := kurtosis_instruction.NewInstructionPosition(16, 33, "dummyFile")
	renderInstruction := newEmptyRenderTemplatesInstruction(nil, position)
	renderInstruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templateAndDataDict
	testArtifactId, err := enclave_data_directory.NewFilesArtifactID()
	require.Nil(t, err)
	renderInstruction.starlarkKwargs[nonOptionalArtifactIdArgName] = starlark.String(testArtifactId)

	expectedConfig := `{"/foo/bar/test.txt": struct(data="{\"Answer\":6,\"LargeFloat\":1231231243.43,\"Name\":\"Stranger\",\"Numbers\":[1,2,3],\"UnixTimeStamp\":1257894000}", template="Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}.")}`
	expectedStr := `render_templates(artifact_id="` + string(testArtifactId) + `", config=` + expectedConfig + `)`
	require.Equal(t, expectedStr, renderInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		RenderTemplatesBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"`+string(testArtifactId)+`"`, nonOptionalArtifactIdArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(expectedConfig, templateAndDataByDestinationRelFilepathArg, false),
		})
	require.Equal(t, canonicalInstruction, renderInstruction.GetCanonicalInstruction())
}

func TestRenderTemplate_TestMultipleTemplates(t *testing.T) {
	templateDataOneStrDict := starlark.StringDict{}
	templateDataOneStrDict["template"] = starlark.String("Hello {{.Name}}")
	templateDataOneStrDict["data"] = starlark.String(`{"Name": "John"}`)
	templateDataTwoStrDict := starlark.StringDict{}
	templateDataTwoStrDict["template"] = starlark.String("Hello {{.LastName}}")
	templateDataTwoStrDict["data"] = starlark.String(`{"LastName": "Doe"}`)

	templateAndDataByDestFilepath := &starlark.Dict{}
	require.Nil(t, templateAndDataByDestFilepath.SetKey(starlark.String("/foo/bar/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataOneStrDict)))
	require.Nil(t, templateAndDataByDestFilepath.SetKey(starlark.String("/fizz/buzz/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataTwoStrDict)))

	position := kurtosis_instruction.NewInstructionPosition(16, 33, "dummyFile")
	renderInstruction := newEmptyRenderTemplatesInstruction(nil, position)
	renderInstruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templateAndDataByDestFilepath
	testArtifactId, err := enclave_data_directory.NewFilesArtifactID()
	require.Nil(t, err)
	renderInstruction.starlarkKwargs[nonOptionalArtifactIdArgName] = starlark.String(testArtifactId)

	// keys of the map are sorted alphabetically by the canonicalizer
	expectedConfig := `{"/fizz/buzz/test.txt": struct(data="{\"LastName\": \"Doe\"}", template="Hello {{.LastName}}"), "/foo/bar/test.txt": struct(data="{\"Name\": \"John\"}", template="Hello {{.Name}}")}`
	expectedStr := `render_templates(artifact_id="` + string(testArtifactId) + `", config=` + expectedConfig + `)`
	require.Equal(t, expectedStr, renderInstruction.String())

	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		position.ToAPIType(),
		RenderTemplatesBuiltinName,
		expectedStr,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionKwarg(`"`+string(testArtifactId)+`"`, nonOptionalArtifactIdArgName, true),
			binding_constructors.NewStarlarkInstructionKwarg(expectedConfig, templateAndDataByDestinationRelFilepathArg, false),
		})
	require.Equal(t, canonicalInstruction, renderInstruction.GetCanonicalInstruction())
}
