package render_templates

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderTemplate_TestStringRepresentation(t *testing.T) {
	template := "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
	templateData := map[string]interface{}{"Name": "Stranger", "Answer": 6, "Numbers": []int{1, 2, 3}, "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43}
	templateDataAsJson, err := json.Marshal(templateData)
	require.Nil(t, err)
	templateAndData := binding_constructors.NewTemplateAndData(template, string(templateDataAsJson))
	templateAndDataByDestFilepath := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		"/foo/bar/test.txt": templateAndData,
	}

	renderInstruction := NewRenderTemplatesInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(16, 33),
		templateAndDataByDestFilepath,
	)

	expectedStr := `render_templates(template_and_data_by_dest_rel_filepath="/foo/bar/test.txt":{"template":"Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}.", "template_data":"{"Answer":6,"LargeFloat":1231231243.43,"Name":"Stranger","Numbers":[1,2,3],"UnixTimeStamp":1257894000}"})`

	require.Equal(t, expectedStr, renderInstruction.String())
}

func TestRenderTemplate_TestMultipleTemplates(t *testing.T) {
	templateDataOne := binding_constructors.NewTemplateAndData("Hello {{.Name}}", "{\"Name\": \"John\"}")
	templateDataTwo := binding_constructors.NewTemplateAndData("Hello {{.LastName}}", "{\"LastName\": \"Doe\"}")
	templateAndDataByDestFilepath := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		"/foo/bar/test.txt":   templateDataOne,
		"/fizz/buzz/test.txt": templateDataTwo,
	}

	renderInstruction := NewRenderTemplatesInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(16, 33),
		templateAndDataByDestFilepath,
	)
	stringRep := renderInstruction.String()

	// as template_data_by_dest_rel_filepath is a map, the output can be either of the two
	expectedStrOne := `render_templates(template_and_data_by_dest_rel_filepath="/foo/bar/test.txt":{"template":"Hello {{.Name}}", "template_data":"{"Name": "John"}"}, "/fizz/buzz/test.txt":{"template":"Hello {{.LastName}}", "template_data":"{"LastName": "Doe"}"})`
	expectedStrTwo := `render_templates(template_and_data_by_dest_rel_filepath="/fizz/buzz/test.txt":{"template":"Hello {{.LastName}}", "template_data":"{"LastName": "Doe"}"}, "/foo/bar/test.txt":{"template":"Hello {{.Name}}", "template_data":"{"Name": "John"}"})`
	comparison := func() bool { return stringRep == expectedStrTwo || stringRep == expectedStrOne }
	require.Condition(t, comparison)
}
