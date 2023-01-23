package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type renderTemplateTestCase1 struct{}

func (test renderTemplateTestCase1) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "SingleTemplate")
}

func (test renderTemplateTestCase1) GetInstruction() (*kurtosis_plan_instruction.KurtosisPlanInstruction, error) {
	return render_templates.NewRenderTemplatesInstruction(nil), nil
}

func (test renderTemplateTestCase1) GetStarlarkCode() (string, error) {
	artifactNameValue := "test-artifact"
	configValue := `{"/foo/bar/test.txt": struct(data="{\"Answer\":6,\"LargeFloat\":1231231243.43,\"Name\":\"Stranger\",\"Numbers\":[1,2,3],\"UnixTimeStamp\":1257894000}", template="Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}.")}`
	return fmt.Sprintf(`%s(%s=%s, %s=%q)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, artifactNameValue), nil
}

func (test renderTemplateTestCase1) GetExpectedArguments() (starlark.StringDict, error) {
	templateDataStrDict := starlark.StringDict{}
	templateDataStrDict["template"] = starlark.String("Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}.")
	templateDataStrDict["data"] = starlark.String(`{"Answer":6,"LargeFloat":1231231243.43,"Name":"Stranger","Numbers":[1,2,3],"UnixTimeStamp":1257894000}`)

	templateAndDataByDestFilepath := &starlark.Dict{}
	if err := templateAndDataByDestFilepath.SetKey(starlark.String("/foo/bar/test.txt"), starlarkstruct.FromStringDict(starlarkstruct.Default, templateDataStrDict)); err != nil {
		return nil, err
	}

	return starlark.StringDict{
		render_templates.TemplateAndDataByDestinationRelFilepathArg: templateAndDataByDestFilepath,
		render_templates.ArtifactNameArgName:                        starlark.String("test-artifact"),
	}, nil
}
