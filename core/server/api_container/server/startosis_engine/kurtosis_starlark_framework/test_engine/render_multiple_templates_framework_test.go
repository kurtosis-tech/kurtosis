package test_engine

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network"
	render_templates2 "github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	renderTemplate_MultipleTemplates_1_filePath = "/fizz/buzz/test.txt"
	renderTemplate_MultipleTemplates_1_data     = `{"LastName": "Doe"}`
	renderTemplate_MultipleTemplates_1_template = "Hello {{.LastName}}"

	renderTemplate_MultipleTemplates_2_filePath = "/foo/bar/test.txt"
	renderTemplate_MultipleTemplates_2_data     = `{"Name": "John"}`
	renderTemplate_MultipleTemplates_2_template = "Hello {{.Name}}"
	mockedFileArtifactName                      = "nature-theme-name-mocked"
)

type renderMultipleTemplatesTestCase struct {
	*testing.T

	serviceNetwork    *service_network.MockServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore
}

func (suite *KurtosisPlanInstructionTestSuite) TestRenderMultipleTemplates() {
	// We expect double quotes for the serialized JSON, for some reasons... See arg_parser.encodeStarlarkObjectAsJSON
	data1WithDoubleQuote := fmt.Sprintf("%q", renderTemplate_MultipleTemplates_1_data)
	templateData1, err := render_templates2.CreateTemplateData(renderTemplate_MultipleTemplates_1_template, data1WithDoubleQuote)
	suite.Require().Nil(err)
	data2WithDoubleQuote := fmt.Sprintf("%q", renderTemplate_MultipleTemplates_2_data)
	templateData2, err := render_templates2.CreateTemplateData(renderTemplate_MultipleTemplates_2_template, data2WithDoubleQuote)
	suite.Require().Nil(err)
	templatesAndData := map[string]*render_templates2.TemplateData{
		renderTemplate_MultipleTemplates_1_filePath: templateData1,
		renderTemplate_MultipleTemplates_2_filePath: templateData2,
	}

	suite.serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Times(1).Return(
		mockedFileArtifactName,
		nil,
	)

	suite.serviceNetwork.EXPECT().RenderTemplates(templatesAndData, mockedFileArtifactName).Times(1).Return(testArtifactUuid, nil)

	suite.run(&renderMultipleTemplatesTestCase{
		T:                 suite.T(),
		serviceNetwork:    suite.serviceNetwork,
		runtimeValueStore: suite.runtimeValueStore,
	})
}

func (t *renderMultipleTemplatesTestCase) GetInstruction() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return render_templates.NewRenderTemplatesInstruction(t.serviceNetwork, t.runtimeValueStore)
}

func (t *renderMultipleTemplatesTestCase) GetStarlarkCode() string {
	configValue := `{"/fizz/buzz/test.txt": struct(data="{\"LastName\": \"Doe\"}", template="Hello {{.LastName}}"), "/foo/bar/test.txt": struct(data="{\"Name\": \"John\"}", template="Hello {{.Name}}")}`
	return fmt.Sprintf(`%s(%s=%s)`, render_templates.RenderTemplatesBuiltinName, render_templates.TemplateAndDataByDestinationRelFilepathArg, configValue)
}

func (t *renderMultipleTemplatesTestCase) GetStarlarkCodeForAssertion() string {
	return ""
}

func (t *renderMultipleTemplatesTestCase) Assert(interpretationResult starlark.Value, executionResult *string) {
	require.Equal(t, starlark.String(mockedFileArtifactName), interpretationResult)

	expectedExecutionResult := fmt.Sprintf("Templates artifact name '%v' rendered with artifact UUID '%s'", mockedFileArtifactName, testArtifactUuid)
	require.Equal(t, expectedExecutionResult, *executionResult)

	// no need to check for the mocked method as we set `.Times(1)` when we declared it
}
