package store_service_files

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
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

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId  kurtosis_backend_service.ServiceID
	src        string
	artifactId enclave_data_directory.FilesArtifactUUID
}

func GenerateStoreServiceFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		storeFilesFromServiceInstruction := newEmptyStoreServiceFilesInstruction(serviceNetwork, instructionPosition)
		if interpretationError := storeFilesFromServiceInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, storeFilesFromServiceInstruction)
		return starlark.String(storeFilesFromServiceInstruction.artifactId), nil
	}
}

func NewStoreServiceFilesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, srcPath string, artifactUuid enclave_data_directory.FilesArtifactUUID, starlarkKwargs starlark.StringDict) *StoreServiceFilesInstruction {
	return &StoreServiceFilesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		src:            srcPath,
		artifactId:     artifactUuid,
		starlarkKwargs: starlarkKwargs,
	}
}

func newEmptyStoreServiceFilesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *StoreServiceFilesInstruction {
	return &StoreServiceFilesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      "",
		src:            "",
		artifactId:     "",
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *StoreServiceFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *StoreServiceFilesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.KurtosisInstruction {
	args := []*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, true),
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[srcArgName]), srcArgName, true),
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalArtifactIdArgName]), nonOptionalArtifactIdArgName, true),
	}
	return binding_constructors.NewKurtosisInstruction(instruction.position.ToAPIType(), StoreServiceFilesBuiltinName, instruction.String(), args)
}

func (instruction *StoreServiceFilesInstruction) Execute(ctx context.Context) (*string, error) {
	_, err := instruction.serviceNetwork.CopyFilesFromServiceToTargetArtifactUUID(ctx, instruction.serviceId, instruction.src, instruction.artifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", instruction.src, instruction.serviceId)
	}
	return nil, nil
}

func (instruction *StoreServiceFilesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(StoreServiceFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *StoreServiceFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating exec with service ID '%v' that does not exist for instruction '%v'", instruction.serviceId, instruction.position.String())
	}
	return nil
}

func (instruction *StoreServiceFilesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var serviceIdArg starlark.String
	var srcPathArg starlark.String
	var artifactIdArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, srcArgName, &srcPathArg, artifactIdArgName, &artifactIdArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}

	if artifactIdArg == emptyStarlarkString {
		placeHolderArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
		if err != nil {
			return startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactIdArg = starlark.String(placeHolderArtifactUuid)
	}

	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[srcArgName] = srcPathArg
	instruction.starlarkKwargs[nonOptionalArtifactIdArgName] = artifactIdArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	artifactId, interpretationErr := kurtosis_instruction.ParseArtifactUuid(nonOptionalArtifactIdArgName, artifactIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.serviceId = serviceId
	instruction.src = srcPath
	instruction.artifactId = artifactId
	return nil
}
