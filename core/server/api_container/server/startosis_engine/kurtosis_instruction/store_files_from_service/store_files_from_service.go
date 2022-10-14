package store_files_from_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"strings"
)

const (
	StoreFileFromServiceBuiltinName = "store_file_from_service"

	serviceIdArgName = "service_id"
	srcPathArgName   = "src_path"

	artifactUuidSuffix = "artifact_uuid"
)

type StoreFilesFromServicePosition struct {
	serviceNetwork service_network.ServiceNetwork

	position  kurtosis_instruction.InstructionPosition
	serviceId kurtosis_backend_service.ServiceID
	srcPath   string
}

func GenerateStoreFilesFromServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, srcPath, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		execInstruction := NewStoreFilesFromServicePosition(serviceNetwork, kurtosis_instruction.GetPositionFromThread(thread), serviceId, srcPath)
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		return starlark.String(execInstruction.position.MagicString(artifactUuidSuffix)), nil
	}
}

func NewStoreFilesFromServicePosition(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, srcPath string) *StoreFilesFromServicePosition {
	return &StoreFilesFromServicePosition{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		srcPath:        srcPath,
	}
}

func (instruction *StoreFilesFromServicePosition) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *StoreFilesFromServicePosition) GetCanonicalInstruction() string {
	buffer := new(strings.Builder)
	buffer.WriteString(StoreFileFromServiceBuiltinName + "(")
	buffer.WriteString(serviceIdArgName + "=\"")
	buffer.WriteString(fmt.Sprintf("%v\", ", instruction.serviceId))
	buffer.WriteString(srcPathArgName + "=\"")
	buffer.WriteString(fmt.Sprintf("%v)", instruction.srcPath))
	return buffer.String()
}

func (instruction *StoreFilesFromServicePosition) Execute(ctx context.Context, environment *startosis_executor.ExecutionEnvironment) error {
	artifactUuid, err := instruction.serviceNetwork.CopyFilesFromService(ctx, instruction.serviceId, instruction.srcPath)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", instruction.srcPath, instruction.serviceId)
	}
	environment.SetArtifactUuid(instruction.position.MagicString(artifactUuidSuffix), string(artifactUuid))
	return nil
}

func (instruction *StoreFilesFromServicePosition) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *StoreFilesFromServicePosition) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, string, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var srcPathArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, srcPathArgName, &srcPathArg); err != nil {
		return "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseSrcPath(srcPathArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	return serviceId, srcPath, nil
}
