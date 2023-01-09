package render_templates

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	templateAndDataByDestinationRelFilepathArg = "config"

	// TODO Deprecate artifactIdArg in a future release
	artifactIdArgName = "artifact_id?"

	artifactNameArgName            = "name?"
	nonOptionalArtifactNameArgName = "name"

	emptyStarlarkString = starlark.String("")
)

type RenderTemplatesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	artifactName string

	templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData
}

func GenerateRenderTemplatesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		renderTemplatesInstruction := newEmptyRenderTemplatesInstruction(serviceNetwork, instructionPosition)

		if interpretationError := renderTemplatesInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, renderTemplatesInstruction)
		return starlark.String(renderTemplatesInstruction.artifactName), nil
	}
}

func newEmptyRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		starlarkKwargs:                    starlark.StringDict{},
		artifactName:                      "",
		templatesAndDataByDestRelFilepath: nil,
	}
}

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, starlarkKwargs starlark.StringDict, artifactName string) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		templatesAndDataByDestRelFilepath: templatesAndDataByDestRelFilepath,
		starlarkKwargs:                    starlarkKwargs,
		artifactName:                      artifactName,
	}
}

func (instruction *RenderTemplatesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *RenderTemplatesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalArtifactNameArgName]), nonOptionalArtifactNameArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg]), templateAndDataByDestinationRelFilepathArg, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), RenderTemplatesBuiltinName, instruction.String(), args)
}

func (instruction *RenderTemplatesInstruction) Execute(_ context.Context) (*string, error) {
	for relFilePath := range instruction.templatesAndDataByDestRelFilepath {
		templateStr := instruction.templatesAndDataByDestRelFilepath[relFilePath].Template
		dataAsJson := instruction.templatesAndDataByDestRelFilepath[relFilePath].DataAsJson
		dataAsJsonMaybeIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(dataAsJson, instruction.serviceNetwork, instruction.GetPositionInOriginalScript().String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while replacing IP address with place holder in the render_template instruction for target '%v'", relFilePath)
		}
		instruction.templatesAndDataByDestRelFilepath[relFilePath] = binding_constructors.NewTemplateAndData(templateStr, dataAsJsonMaybeIPAddressReplaced)
	}

	artifactUUID, err := instruction.serviceNetwork.RenderTemplates(instruction.templatesAndDataByDestRelFilepath, instruction.artifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to render templates '%v'", instruction.templatesAndDataByDestRelFilepath)
	}
	instructionResult := fmt.Sprintf("Templates artifact name '%s' rendered with artifact UUID '%s'", instruction.artifactName, artifactUUID)
	return &instructionResult, nil
}

func (instruction *RenderTemplatesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(RenderTemplatesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *RenderTemplatesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if environment.DoesArtifactNameExist(instruction.artifactName) {
		return stacktrace.NewError("There was an error validating '%v' as artifact name '%v' already exists", RenderTemplatesBuiltinName, instruction.artifactName)
	}
	environment.AddArtifactName(instruction.artifactName)
	return nil
}

func (instruction *RenderTemplatesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var templatesAndDataArg *starlark.Dict
	var artifactIdArg = emptyStarlarkString
	var artifactNameArg = emptyStarlarkString

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, templateAndDataByDestinationRelFilepathArg, &templatesAndDataArg, artifactNameArgName, &artifactNameArg, artifactIdArgName, &artifactIdArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", RenderTemplatesBuiltinName, args, kwargs)
	}

	if artifactIdArg == emptyStarlarkString && artifactNameArg == emptyStarlarkString {
		return startosis_errors.NewInterpretationError("A name must be provided for the artifact using the '%v' argument", nonOptionalArtifactNameArgName)
	}

	if artifactNameArg == emptyStarlarkString {
		artifactNameArg = artifactIdArg
	}

	instruction.starlarkKwargs[nonOptionalArtifactNameArgName] = artifactNameArg
	instruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templatesAndDataArg

	templatesAndDataByDestRelFilepath, interpretationErr := kurtosis_instruction.ParseTemplatesAndData(templatesAndDataArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath

	artifactName, interpretationErr := kurtosis_instruction.ParseNonEmptyString(nonOptionalArtifactNameArgName, artifactNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.artifactName = artifactName

	return nil
}
