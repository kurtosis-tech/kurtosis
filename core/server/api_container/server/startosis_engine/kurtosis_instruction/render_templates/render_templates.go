package render_templates

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	templateAndDataByDestinationRelFilepathArg = "template_and_data_by_dest_rel_filepath"
)

type RenderTemplatesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

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
		return starlark.String(renderTemplatesInstruction.position.MagicString(shared_helpers.ArtifactUUIDSuffix)), nil
	}
}

func newEmptyRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		starlarkKwargs: starlark.StringDict{},
	}
}

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, templatesAndDataByDestRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, starlarkKwargs starlark.StringDict) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:                    serviceNetwork,
		position:                          position,
		templatesAndDataByDestRelFilepath: templatesAndDataByDestRelFilepath,
		starlarkKwargs:                    starlarkKwargs,
	}
}

func (instruction *RenderTemplatesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *RenderTemplatesInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(RenderTemplatesBuiltinName, instruction.starlarkKwargs, &instruction.position)
}

func (instruction *RenderTemplatesInstruction) Execute(ctx context.Context, environment *startosis_executor.ExecutionEnvironment) error {
	artifactUuid, err := instruction.serviceNetwork.RenderTemplates(instruction.templatesAndDataByDestRelFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to render templates '%v'", instruction.templatesAndDataByDestRelFilepath)
	}
	environment.SetArtifactUuid(instruction.position.MagicString(shared_helpers.ArtifactUUIDSuffix), string(artifactUuid))
	return nil
}

func (instruction *RenderTemplatesInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *RenderTemplatesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func (instruction *RenderTemplatesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var templatesAndDataArg *starlark.Dict
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, templateAndDataByDestinationRelFilepathArg, &templatesAndDataArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}
	instruction.starlarkKwargs[templateAndDataByDestinationRelFilepathArg] = templatesAndDataArg

	templatesAndDataByDestRelFilepath, interpretationErr := kurtosis_instruction.ParseTemplatesAndData(templatesAndDataArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath
	return nil
}
