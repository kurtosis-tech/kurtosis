package test_engine

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	render_templates2 "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	renderTemplate_SingleTemplate_filePath = "/foo/bar/test.txt"
	renderTemplate_SingleTemplate_data     = `{"Answer":6,"LargeFloat":1231231243.43,"Name":"Stranger","Numbers":[1,2,3],"UnixTimeStamp":1257894000}`
	renderTemplate_SingleTemplate_template = "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}."
)

type renderSingleTemplateTestCase struct {
	*testing.T

	serviceNetwork *service_network.MockServiceNetwork
}

func newRenderSingleTemplateTestCase(t *testing.T) *renderSingleTemplateTestCase {
	serviceNetwork := service_network.NewMockServiceNetwork(t)
	// We expect double quotes for the serialized JSON, for some reasons... See arg_parser.encodeStarlarkObjectAsJSON
	dataWithDoubleQuote := fmt.Sprintf("%q", renderTemplate_SingleTemplate_data)
	templateData, err := render_templates2.CreateTemplateData(renderTemplate_SingleTemplate_template, dataWithDoubleQuote)
	require.Nil(t, err)
	templateAndData := map[string]*render_templates2.TemplateData{
		renderTemplate_SingleTemplate_filePath: templateData,
	}

	serviceNetwork.EXPECT().RenderTemplates(templateAndData, TestArtifactName).Times(1).Return(TestArtifactUuid, nil)
	return &renderSingleTemplateTestCase{
		T:              t,
		serviceNetwork: serviceNetwork,
	}
}

func (t renderSingleTemplateTestCase) GetId() string {
	return fmt.Sprintf("%s_%s", render_templates.RenderTemplatesBuiltinName, "SingleTemplate")
}

func (t renderSingleTemplateTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	enclaveDb := getEnclaveDBForTest(t.T)
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(enclaveDb)
	require.NoError(t, err)
	return render_templates.NewRenderTemplatesInstruction(t.serviceNetwork, runtimeValueStore)
}

func (t renderSingleTemplateTestCase) GetStarlarkCode() string {
	configValue := fmt.Sprintf(`{%q: struct(data=%q, template=%q)}`, renderTemplate_SingleTemplate_filePath, renderTemplate_SingleTemplate_data, renderTemplate_SingleTemplate_template)
	return fmt.Sprintf(`%s(%s=%s, %s=%q)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue, render_templates.ArtifactNameArgName, TestArtifactName)
}

func (t *renderSingleTemplateTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t renderSingleTemplateTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(TestArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Templates artifact name '%s' rendered with artifact UUID '%s'", TestArtifactName, TestArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)

	// no need to check for the mocked method as we set `.Times(1)` when we declared it
}
