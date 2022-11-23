package render_templates

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

func TestRenderTemplate_TestStringRepresentation(t *testing.T) {
	template := "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
	templateData := map[string]interface{}{"Name": "Stranger", "Answer": 6, "Numbers": []int{1, 2, 3}, "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43}
	templateDataAsJson, err := json.Marshal(templateData)
	require.Nil(t, err)
	templateAndDataDict := &starlark.Dict{}
	templateDict := &starlark.Dict{}
	require.Nil(t, templateDict.SetKey(starlark.String("template"), starlark.String(template)))
	require.Nil(t, templateDict.SetKey(starlark.String("template_data_json"), starlark.String(templateDataAsJson)))
	require.Nil(t, templateAndDataDict.SetKey(starlark.String("/foo/bar/test.txt"), templateDict))

	renderInstruction := newEmptyRenderTemplatesInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(16, 33, "dummyFile"),
	)
	renderInstruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templateAndDataDict
	testArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	renderInstruction.starlarkKwargs[nonOptionalArtifactUuidArgName] = starlark.String(testArtifactUuid)

	expectedStr := `# from: dummyFile[16:33]
render_templates(
	artifact_uuid="` + string(testArtifactUuid) + `",
	config={
		"/foo/bar/test.txt": {
			"template": "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}.",
			"template_data_json": "{\"Answer\":6,\"LargeFloat\":1231231243.43,\"Name\":\"Stranger\",\"Numbers\":[1,2,3],\"UnixTimeStamp\":1257894000}"
		}
	}
)`
	require.Equal(t, expectedStr, renderInstruction.GetCanonicalInstruction())
}

func TestRenderTemplate_TestMultipleTemplates(t *testing.T) {
	templateDataOne := &starlark.Dict{}
	require.Nil(t, templateDataOne.SetKey(starlark.String("template"), starlark.String("Hello {{.Name}}")))
	require.Nil(t, templateDataOne.SetKey(starlark.String("template_data_json"), starlark.String(`{"Name": "John"}`)))
	templateDataTwo := &starlark.Dict{}
	require.Nil(t, templateDataTwo.SetKey(starlark.String("template"), starlark.String("Hello {{.LastName}}")))
	require.Nil(t, templateDataTwo.SetKey(starlark.String("template_data_json"), starlark.String(`{"LastName": "Doe"}`)))

	templateAndDataByDestFilepath := &starlark.Dict{}
	require.Nil(t, templateAndDataByDestFilepath.SetKey(starlark.String("/foo/bar/test.txt"), templateDataOne))
	require.Nil(t, templateAndDataByDestFilepath.SetKey(starlark.String("/fizz/buzz/test.txt"), templateDataTwo))

	renderInstruction := newEmptyRenderTemplatesInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(16, 33, "dummyFile"),
	)
	renderInstruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templateAndDataByDestFilepath
	testArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	renderInstruction.starlarkKwargs[nonOptionalArtifactUuidArgName] = starlark.String(testArtifactUuid)

	// keys of the map are sorted alphabetically by the canonicalizer
	expectedStr := `# from: dummyFile[16:33]
render_templates(
	artifact_uuid="` + string(testArtifactUuid) + `",
	config={
		"/fizz/buzz/test.txt": {
			"template": "Hello {{.LastName}}",
			"template_data_json": "{\"LastName\": \"Doe\"}"
		},
		"/foo/bar/test.txt": {
			"template": "Hello {{.Name}}",
			"template_data_json": "{\"Name\": \"John\"}"
		}
	}
)`
	require.Equal(t, expectedStr, renderInstruction.GetCanonicalInstruction())
}
