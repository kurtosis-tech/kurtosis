package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// TODO: Remove when artifactId gets removed. This is only to test backward compat
type renderTemplateTestCase3 struct{}

func (test renderTemplateTestCase3) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "ValidateArtifactIdParam")
}

func (test renderTemplateTestCase3) GetInstruction() (*kurtosis_plan_instruction.KurtosisPlanInstruction, error) {
	return render_templates.NewRenderTemplatesInstruction(nil), nil
}

func (test renderTemplateTestCase3) GetStarlarkCode() (string, error) {
	artifactNameValue := "test-artifact"
	configValue := `{"/foo/bar/test.txt": struct(data="{\"Name\":\"World\"}", template="Hello {{.Name}}!")}`
	return fmt.Sprintf(`%s(%s=%s, %s="", %s=%q)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, render_templates.ArtifactIdArgName, artifactNameValue), nil
}

func (test renderTemplateTestCase3) GetExpectedArguments() (starlark.StringDict, error) {
	templateDataStrDict := starlark.StringDict{}
	templateDataStrDict["template"] = starlark.String("Hello {{.Name}}!")
	templateDataStrDict["data"] = starlark.String(`{"Name":"World"}`)

	templateAndDataByDestFilepath := &starlark.Dict{}
	if err := templateAndDataByDestFilepath.SetKey(starlark.String("/foo/bar/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataStrDict)); err != nil {
		return nil, err
	}

	return starlark.StringDict{
		render_templates.TemplateAndDataByDestinationRelFilepathArg: templateAndDataByDestFilepath,
		render_templates.ArtifactIdArgName:                          starlark.String("test-artifact"),
	}, nil
}
