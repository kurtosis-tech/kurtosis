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
	renderTemplate_MultipleTemplates_artifactName     = "test-artifact"
	renderTemplate_MultipleTemplates_fileArtifactUuid = enclave_data_directory.FilesArtifactUUID("test-artifact-uuid")

	renderTemplate_MultipleTemplates_1_filePath = "/fizz/buzz/test.txt"
	renderTemplate_MultipleTemplates_1_data     = `{"LastName": "Doe"}`
	renderTemplate_MultipleTemplates_1_template = "Hello {{.LastName}}"

	renderTemplate_MultipleTemplates_2_filePath = "/foo/bar/test.txt"
	renderTemplate_MultipleTemplates_2_data     = `{"Name": "John"}`
	renderTemplate_MultipleTemplates_2_template = "Hello {{.Name}}"
)

type renderTemplateTestCase2 struct {
	*testing.T

	serviceNetwork *service_network.MockServiceNetwork
}

func newRenderTemplateTestCase2(t *testing.T) *renderTemplateTestCase2 {
	serviceNetwork := service_network.NewMockServiceNetwork(t)

	// We expect double quotes for the serialized JSON, for some reasons... See arg_parser.encodeStarlarkObjectAsJSON
	data1WithDoubleQuote := fmt.Sprintf("%q", renderTemplate_MultipleTemplates_1_data)
	data2WithDoubleQuote := fmt.Sprintf("%q", renderTemplate_MultipleTemplates_2_data)
	templatesAndData := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		renderTemplate_MultipleTemplates_1_filePath: binding_constructors.NewTemplateAndData(renderTemplate_MultipleTemplates_1_template, data1WithDoubleQuote),
		renderTemplate_MultipleTemplates_2_filePath: binding_constructors.NewTemplateAndData(renderTemplate_MultipleTemplates_2_template, data2WithDoubleQuote),
	}
	serviceNetwork.EXPECT().RenderTemplates(templatesAndData, renderTemplate_MultipleTemplates_artifactName).Times(1).Return(renderTemplate_MultipleTemplates_fileArtifactUuid, nil)

	return &renderTemplateTestCase2{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t renderTemplateTestCase2) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "MultipleTemplates")
}

func (t renderTemplateTestCase2) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return render_templates.NewRenderTemplatesInstruction(t.serviceNetwork)
}

func (t renderTemplateTestCase2) GetStarlarkCode() string {
	configValue := `{"/fizz/buzz/test.txt": struct(data="{\"LastName\": \"Doe\"}", template="Hello {{.LastName}}"), "/foo/bar/test.txt": struct(data="{\"Name\": \"John\"}", template="Hello {{.Name}}")}`
	return fmt.Sprintf(`%s(%s=%s, %s=%q)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, renderTemplate_MultipleTemplates_artifactName)
}

func (t *renderTemplateTestCase2) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t renderTemplateTestCase2) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(renderTemplate_MultipleTemplates_artifactName), interpretationResult)
	require.Equal(t, "Templates artifact name 'test-artifact' rendered with artifact UUID 'test-artifact-uuid'", *executionResult)

	// no need to check for the mocked method as we set `.Times(1)` when we declared it
}
