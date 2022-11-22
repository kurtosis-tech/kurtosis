package store_service_files

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
	StoreServiceFilesBuiltinName = "store_service_files"

	serviceIdArgName = "service_id"
	srcArgName       = "src"

	artifactIdArgName            = "artifact_id?"
	nonOptionalArtifactIdArgName = "artifact_id"

	emptyStarlarkString = starlark.String("")
)

type StoreServiceFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position     *kurtosis_instruction.InstructionPosition
	serviceId    kurtosis_backend_service.ServiceID
	srcPath      string
	artifactUuid enclave_data_directory.FilesArtifactUUID
}

func GenerateStoreServiceFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, srcPath, artifactUuid, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		storeFilesFromServiceInstruction := NewStoreFilesFromServiceInstruction(serviceNetwork, instructionPosition, serviceId, srcPath, artifactUuid)
		*instructionsQueue = append(*instructionsQueue, storeFilesFromServiceInstruction)
		return starlark.String(artifactUuid), nil
	}
}

func NewStoreFilesFromServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, srcPath string, artifactUuid enclave_data_directory.FilesArtifactUUID) *StoreServiceFilesInstruction {
	return &StoreServiceFilesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		srcPath:        srcPath,
		artifactUuid:   artifactUuid,
	}
}

func (instruction *StoreServiceFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *StoreServiceFilesInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(StoreServiceFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), instruction.position)
}

func (instruction *StoreServiceFilesInstruction) Execute(ctx context.Context) (*string, error) {
	_, err := instruction.serviceNetwork.CopyFilesFromServiceToTargetArtifactUUID(ctx, instruction.serviceId, instruction.srcPath, instruction.artifactUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", instruction.srcPath, instruction.serviceId)
	}
	return nil, nil
}

func (instruction *StoreServiceFilesInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(StoreServiceFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs(), instruction.position)
}

func (instruction *StoreServiceFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist for instruction '%v'", instruction.serviceId, instruction.position.String())
	}
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, string, enclave_data_directory.FilesArtifactUUID, *startosis_errors.InterpretationError) {

	var serviceIdArg starlark.String
	var srcPathArg starlark.String
	var artifactUuidArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, srcArgName, &srcPathArg, artifactIdArgName, &artifactUuidArg); err != nil {
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

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return "", "", "", interpretationErr
	}

	artifactUuid, interpretationErr := kurtosis_instruction.ParseArtifactUuid(nonOptionalArtifactIdArgName, artifactUuidArg)
	if interpretationErr != nil {
		return "", "", "", interpretationErr
	}

	return serviceId, srcPath, artifactUuid, nil
}

func (instruction *StoreServiceFilesInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		serviceIdArgName:             starlark.String(instruction.serviceId),
		srcArgName:                   starlark.String(instruction.srcPath),
		nonOptionalArtifactIdArgName: starlark.String(instruction.artifactUuid),
	}
}
