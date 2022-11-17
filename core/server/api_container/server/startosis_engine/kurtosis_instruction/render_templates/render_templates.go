package render_templates

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	templateAndDataByDestinationRelFilepathArg = "template_and_data_by_dest_rel_filepath"

	artifactUuidArgName            = "artifact_uuid?"
	nonOptionalArtifactUuidArgName = "artifact_uuid"

	emptyStarlarkString = starlark.String("")
)

type RenderTemplatesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	artifactUuid enclave_data_directory.FilesArtifactUUID

	templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData
}

func GenerateRenderTemplatesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		renderTemplatesInstruction := newEmptyRenderTemplatesInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread))

		if interpretationError := renderTemplatesInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, renderTemplatesInstruction)
		return starlark.String(renderTemplatesInstruction.artifactUuid), nil
	}
}

func newEmptyRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		starlarkKwargs: starlark.StringDict{},
	}
}

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, starlarkKwargs starlark.StringDict, artifactUuid enclave_data_directory.FilesArtifactUUID) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		templatesAndDataByDestRelFilepath: templatesAndDataByDestRelFilepath,
		starlarkKwargs:                    starlarkKwargs,
		artifactUuid:                      artifactUuid,
	}
}

func (instruction *RenderTemplatesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *RenderTemplatesInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(RenderTemplatesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs, &instruction.position)
}

func (instruction *RenderTemplatesInstruction) Execute(_ context.Context) (*string, error) {
	for relFilePath := range instruction.templatesAndDataByDestRelFilepath {
		templateStr := instruction.templatesAndDataByDestRelFilepath[relFilePath].Template
		dataAsJson := instruction.templatesAndDataByDestRelFilepath[relFilePath].DataAsJson
		dataAsJsonMaybeIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(dataAsJson, instruction.serviceNetwork, instruction.GetPositionInOriginalScript().String())
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while replacing IP address with place holder in the render_template instruction for target '%v'", relFilePath)
		}
		instruction.templatesAndDataByDestRelFilepath[relFilePath] = binding_constructors.NewTemplateAndData(templateStr, dataAsJsonMaybeIPAddressReplaced)
	}

	_, err := instruction.serviceNetwork.RenderTemplatesToTargetFilesArtifactUUID(instruction.templatesAndDataByDestRelFilepath, instruction.artifactUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to render templates '%v'", instruction.templatesAndDataByDestRelFilepath)
	}
	return nil, nil
}

func (instruction *RenderTemplatesInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(RenderTemplatesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs, &instruction.position)
}

func (instruction *RenderTemplatesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func (instruction *RenderTemplatesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var templatesAndDataArg *starlark.Dict
	var artifactUuidArg = emptyStarlarkString

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, templateAndDataByDestinationRelFilepathArg, &templatesAndDataArg, artifactUuidArgName, &artifactUuidArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}

	if artifactUuidArg == emptyStarlarkString {
		placeHolderArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
		if err != nil {
			return startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactUuidArg = starlark.String(placeHolderArtifactUuid)
	}

	instruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templatesAndDataArg

	templatesAndDataByDestRelFilepath, interpretationErr := kurtosis_instruction.ParseTemplatesAndData(templatesAndDataArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath

	artifactUuid, interpretationErr := kurtosis_instruction.ParseArtifactUuid(nonOptionalArtifactUuidArgName, artifactUuidArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.artifactUuid = artifactUuid
	instruction.starlarkKwargs[nonOptionalArtifactUuidArgName] = starlark.String(artifactUuid)

	return nil
}
