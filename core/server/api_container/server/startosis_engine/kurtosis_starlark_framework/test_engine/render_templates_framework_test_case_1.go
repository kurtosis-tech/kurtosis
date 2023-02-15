package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	renderTemplate_SingleTemplate_ArtifactName = "test-artifact"

	renderTemplate_SingleTemplate_filePath = "/foo/bar/test.txt"
	renderTemplate_SingleTemplate_data     = "{\"Answer\":6,\"LargeFloat\":1231231243.43,\"Name\":\"Stranger\",\"Numbers\":[1,2,3],\"UnixTimeStamp\":1257894000}"
	renderTemplate_SingleTemplate_template = "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."

	renderTemplate_SingleTemplate_fileArtifactUuid = enclave_data_directory.FilesArtifactUUID("test-artifact-uuid")
)

type renderTemplateTestCase1 struct {
	*testing.T

	serviceNetwork *service_network.MockServiceNetwork
}

func newRenderTemplateTestCase1(t *testing.T) *renderTemplateTestCase1 {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	// We expect double quotes for the serialized JSON, for some reasons... See arg_parser.encodeStarlarkObjectAsJSON
	dataWithDoubleQuote := fmt.Sprintf("%q", renderTemplate_SingleTemplate_data)
	templateAndData := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		renderTemplate_SingleTemplate_filePath: binding_constructors.NewTemplateAndData(renderTemplate_SingleTemplate_template, dataWithDoubleQuote),
	}

	serviceNetwork.EXPECT().RenderTemplates(templateAndData, renderTemplate_SingleTemplate_ArtifactName).Times(1).Return(renderTemplate_SingleTemplate_fileArtifactUuid, nil)
	return &renderTemplateTestCase1{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t renderTemplateTestCase1) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "SingleTemplate")
}

func (t renderTemplateTestCase1) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return render_templates.NewRenderTemplatesInstruction(t.serviceNetwork)
}

func (t renderTemplateTestCase1) GetStarlarkCode() string {
	configValue := fmt.Sprintf(`{%q: struct(data=%q, template=%q)}`, renderTemplate_SingleTemplate_filePath, renderTemplate_SingleTemplate_data, renderTemplate_SingleTemplate_template)
	return fmt.Sprintf(`%s(%s=%s, %s=%q)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, renderTemplate_SingleTemplate_ArtifactName)
}

func (t renderTemplateTestCase1) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(renderTemplate_SingleTemplate_ArtifactName), interpretationResult)
	require.Equal(t, "Templates artifact name 'test-artifact' rendered with artifact UUID 'test-artifact-uuid'", *executionResult)

	// no need to check for the mocked method as we set `.Times(1)` when we declared it
}
