package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type renderTemplateTestCase2 struct{}

func (test renderTemplateTestCase2) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "MultipleTemplates")
}

func (test renderTemplateTestCase2) GetInstruction() (*kurtosis_plan_instruction.KurtosisPlanInstruction, error) {
	return render_templates.NewRenderTemplatesInstruction(nil), nil
}

func (test renderTemplateTestCase2) GetStarlarkCode() (string, error) {
	artifactNameValue := "test-artifact"
	configValue := `{"/fizz/buzz/test.txt": struct(data="{\"LastName\": \"Doe\"}", template="Hello {{.LastName}}"), "/foo/bar/test.txt": struct(data="{\"Name\": \"John\"}", template="Hello {{.Name}}")}`
	return fmt.Sprintf(`%s(%s=%s, %s=%q, %s="")`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, artifactNameValue, render_templates.ArtifactIdArgName), nil
}

func (test renderTemplateTestCase2) GetExpectedArguments() (starlark.StringDict, error) {
	templateDataOneStrDict := starlark.StringDict{}
	templateDataOneStrDict["template"] = starlark.String("Hello {{.Name}}")
	templateDataOneStrDict["data"] = starlark.String(`{"Name": "John"}`)
	templateDataTwoStrDict := starlark.StringDict{}
	templateDataTwoStrDict["template"] = starlark.String("Hello {{.LastName}}")
	templateDataTwoStrDict["data"] = starlark.String(`{"LastName": "Doe"}`)

	templateAndDataByDestFilepath := &starlark.Dict{}
	if err := templateAndDataByDestFilepath.SetKey(starlark.String("/fizz/buzz/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataTwoStrDict)); err != nil {
		return nil, err
	}
	if err := templateAndDataByDestFilepath.SetKey(starlark.String("/foo/bar/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataOneStrDict)); err != nil {
		return nil, err
	}

	return starlark.StringDict{
		render_templates.TemplateAndDataByDestinationRelFilepathArg: templateAndDataByDestFilepath,
		render_templates.ArtifactNameArgName:                        starlark.String("test-artifact"),
	}, nil
}
