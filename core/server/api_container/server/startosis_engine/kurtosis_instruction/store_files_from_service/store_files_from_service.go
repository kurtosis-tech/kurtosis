package store_files_from_service

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
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
	StoreFileFromServiceBuiltinName = "store_file_from_service"

	serviceIdArgName = "service_id"
	srcPathArgName   = "src_path"

	artifactUuidArgName            = "artifact_uuid?"
	nonOptionalArtifactUuidArgName = "artifact_uuid"

	emptyStarlarkString = starlark.String("")
)

type StoreFilesFromServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position     kurtosis_instruction.InstructionPosition
	serviceId    kurtosis_backend_service.ServiceID
	srcPath      string
	artifactUuid enclave_data_directory.FilesArtifactUUID
}

func GenerateStoreFilesFromServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, srcPath, artifactUuid, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		storeFilesFromServiceInstruction := NewStoreFilesFromServiceInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread), serviceId, srcPath, artifactUuid)
		*instructionsQueue = append(*instructionsQueue, storeFilesFromServiceInstruction)
		return starlark.String(artifactUuid), nil
	}
}

func NewStoreFilesFromServiceInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, srcPath string, artifactUuid enclave_data_directory.FilesArtifactUUID) *StoreFilesFromServiceInstruction {
	return &StoreFilesFromServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		srcPath:        srcPath,
		artifactUuid:   artifactUuid,
	}
}

func (instruction *StoreFilesFromServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *StoreFilesFromServiceInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(StoreFileFromServiceBuiltinName, instruction.getKwargs(), &instruction.position)
}

func (instruction *StoreFilesFromServiceInstruction) Execute(ctx context.Context) error {
	_, err := instruction.serviceNetwork.CopyFilesFromServiceToTargetArtifactUUID(ctx, instruction.serviceId, instruction.srcPath, instruction.artifactUuid)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", instruction.srcPath, instruction.serviceId)
	}
	return nil
}

func (instruction *StoreFilesFromServiceInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(StoreFileFromServiceBuiltinName, instruction.getKwargs(), &instruction.position)
}

func (instruction *StoreFilesFromServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist for instruction '%v'", instruction.serviceId, instruction.position.String())
	}
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, string, enclave_data_directory.FilesArtifactUUID, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var srcPathArg starlark.String
	var artifactUuidArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, srcPathArgName, &srcPathArg, artifactUuidArgName, &artifactUuidArg); err != nil {
		return "", "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	if artifactUuidArg == emptyStarlarkString {
		placeHolderArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
		if err != nil {
			return "", "", "", startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactUuidArg = starlark.String(placeHolderArtifactUuid)
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", "", "", interpretationErr
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseFilePath(srcPathArgName, srcPathArg)
	if interpretationErr != nil {
		return "", "", "", interpretationErr
	}

	artifactUuid, interpretationErr := kurtosis_instruction.ParseArtifactUuid(nonOptionalArtifactUuidArgName, artifactUuidArg)
	if interpretationErr != nil {
		return "", "", "", interpretationErr
	}

	return serviceId, srcPath, artifactUuid, nil
}

func (instruction *StoreFilesFromServiceInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		serviceIdArgName:               starlark.String(instruction.serviceId),
		srcPathArgName:                 starlark.String(instruction.srcPath),
		nonOptionalArtifactUuidArgName: starlark.String(instruction.artifactUuid),
	}
}
