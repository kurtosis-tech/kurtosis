package render_templates

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	TemplateAndDataByDestinationRelFilepathArg = "config"
	ArtifactNameArgName                        = "name"
)

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RenderTemplatesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              TemplateAndDataByDestinationRelFilepathArg,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator:         nil,
				},
				{
					Name:              ArtifactNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RenderTemplatesCapabilities{
				serviceNetwork:                    serviceNetwork,
				artifactName:                      "",  // will be populated at interpretation time
				templatesAndDataByDestRelFilepath: nil, // will be populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ArtifactNameArgName: true,
		},
	}
}

type RenderTemplatesCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	artifactName                      string
	templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData
}

func (builtin *RenderTemplatesCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	artifactName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ArtifactNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%s'", ArtifactNameArgName)
	}
	builtin.artifactName = artifactName.GoString()

	config, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, TemplateAndDataByDestinationRelFilepathArg)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%s'", TemplateAndDataByDestinationRelFilepathArg)
	}
	templatesAndDataByDestRelFilepath, interpretationErr := kurtosis_instruction.ParseTemplatesAndData(config)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath
	return artifactName, nil
}

func (builtin *RenderTemplatesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	artifactName := builtin.artifactName
	if validatorEnvironment.DoesArtifactNameExist(artifactName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' already exists", RenderTemplatesBuiltinName, artifactName)
	}
	validatorEnvironment.AddArtifactName(artifactName)
	return nil
}

func (builtin *RenderTemplatesCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	for relFilePath := range builtin.templatesAndDataByDestRelFilepath {
		templateStr := builtin.templatesAndDataByDestRelFilepath[relFilePath].Template
		dataAsJson := builtin.templatesAndDataByDestRelFilepath[relFilePath].DataAsJson
		dataAsJsonMaybeIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(dataAsJson, builtin.serviceNetwork, TemplateAndDataByDestinationRelFilepathArg)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while replacing IP address with place holder in the render_template instruction for target '%v'", relFilePath)
		}
		builtin.templatesAndDataByDestRelFilepath[relFilePath] = binding_constructors.NewTemplateAndData(templateStr, dataAsJsonMaybeIPAddressReplaced)
	}

	artifactUUID, err := builtin.serviceNetwork.RenderTemplates(builtin.templatesAndDataByDestRelFilepath, builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to render templates '%v'", builtin.templatesAndDataByDestRelFilepath)
	}
	instructionResult := fmt.Sprintf("Templates artifact name '%s' rendered with artifact UUID '%s'", builtin.artifactName, artifactUUID)
	return instructionResult, nil
}
