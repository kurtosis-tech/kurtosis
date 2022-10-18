package render_templates

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"strings"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	templateAndDataByDestinationRelFilepathArg = "templates_and_data_by_destination_rel_filepath"
)

type RenderTemplatesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position         kurtosis_instruction.InstructionPosition
	templatesAndData map[string]*enclaves.TemplateAndData
}

func GenerateRenderTemplatesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		templatesAndData, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		storeFilesFromServiceInstruction := NewRenderTemplatesInstruction(serviceNetwork, *shared_helpers.GetPositionFromThread(thread), templatesAndData)
		*instructionsQueue = append(*instructionsQueue, storeFilesFromServiceInstruction)
		return starlark.String(storeFilesFromServiceInstruction.position.MagicString(shared_helpers.ArtifactUUIDSuffix)), nil
	}
}

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, templatesAndData map[string]*enclaves.TemplateAndData) *RenderTemplatesInstruction {
	return &RenderTemplatesInstruction{
		serviceNetwork:   serviceNetwork,
		position:         position,
		templatesAndData: templatesAndData,
	}
}

func (instruction *RenderTemplatesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *RenderTemplatesInstruction) GetCanonicalInstruction() string {
	buffer := new(strings.Builder)
	buffer.WriteString(RenderTemplatesBuiltinName + "(")
	buffer.WriteString(templateAndDataByDestinationRelFilepathArg + "=\"")
	buffer.WriteString(fmt.Sprintf("%v\")", instruction.templatesAndData))
	return buffer.String()
}

func (instruction *RenderTemplatesInstruction) Execute(ctx context.Context, environment *startosis_executor.ExecutionEnvironment) error {
	artifactUuid, err := instruction.serviceNetwork.CopyFilesFromService(ctx, "", "")
	if err != nil {
		return stacktrace.Propagate(err, "Failed to render template '%v'", instruction.templatesAndData)
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

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (map[string]*enclaves.TemplateAndData, *startosis_errors.InterpretationError) {

	var templatesAndData starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, templateAndDataByDestinationRelFilepathArg, &templatesAndData); err != nil {
		return nil, startosis_errors.NewInterpretationError(err.Error())
	}

	//serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	//if interpretationErr != nil {
	//	return "", "", interpretationErr
	//}

	return nil, nil
}
