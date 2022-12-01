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
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	templateAndDataByDestinationRelFilepathArg = "config"

	artifactIdArgName            = "artifact_id?"
	nonOptionalArtifactIdArgName = "artifact_id"

	emptyStarlarkString = starlark.String("")
)

type RenderTemplatesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	artifactId enclave_data_directory.FilesArtifactID

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
		return starlark.String(renderTemplatesInstruction.artifactId), nil
	}
}

func newEmptyRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		starlarkKwargs:                    starlark.StringDict{},
		artifactId:                        "",
		templatesAndDataByDestRelFilepath: nil,
	}
}

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, starlarkKwargs starlark.StringDict, artifactUuid enclave_data_directory.FilesArtifactID) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		templatesAndDataByDestRelFilepath: templatesAndDataByDestRelFilepath,
		starlarkKwargs:                    starlarkKwargs,
		artifactId:                        artifactUuid,
	}
}

func (instruction *RenderTemplatesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *RenderTemplatesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalArtifactIdArgName]), nonOptionalArtifactIdArgName, kurtosis_instruction.Representative),
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

	artifactId, err := instruction.serviceNetwork.RenderTemplatesToTargetFilesArtifactUUID(instruction.templatesAndDataByDestRelFilepath, instruction.artifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to render templates '%v'", instruction.templatesAndDataByDestRelFilepath)
	}
	instructionResult := fmt.Sprintf("Templates rendered and stored with artifact ID '%s'", artifactId)
	return &instructionResult, nil
}

func (instruction *RenderTemplatesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(RenderTemplatesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *RenderTemplatesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if environment.DoesArtifactIdExist(instruction.artifactId) {
		return stacktrace.NewError("There was an error validating '%v' as artifact UUID '%v' already exists", RenderTemplatesBuiltinName, instruction.artifactId)
	}
	environment.AddArtifactId(instruction.artifactId)
	return nil
}

func (instruction *RenderTemplatesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var templatesAndDataArg *starlark.Dict
	var artifactIdArg = emptyStarlarkString

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, templateAndDataByDestinationRelFilepathArg, &templatesAndDataArg, artifactIdArgName, &artifactIdArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", RenderTemplatesBuiltinName, args, kwargs)
	}

	if artifactIdArg == emptyStarlarkString {
		placeHolderArtifactUuid, err := enclave_data_directory.NewFilesArtifactID()
		if err != nil {
			return startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactIdArg = starlark.String(placeHolderArtifactUuid)
	}
	instruction.starlarkKwargs[nonOptionalArtifactIdArgName] = artifactIdArg
	instruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templatesAndDataArg

	templatesAndDataByDestRelFilepath, interpretationErr := kurtosis_instruction.ParseTemplatesAndData(templatesAndDataArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath

	artifactUuid, interpretationErr := kurtosis_instruction.ParseArtifactId(nonOptionalArtifactIdArgName, artifactIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.artifactId = artifactUuid
	instruction.starlarkKwargs[nonOptionalArtifactIdArgName] = starlark.String(artifactUuid)

	return nil
}
