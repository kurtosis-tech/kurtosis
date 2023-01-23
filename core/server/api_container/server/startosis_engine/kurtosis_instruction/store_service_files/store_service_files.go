package store_service_files

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	StoreServiceFilesBuiltinName = "store_service_files"

	serviceNameArgName = "service_name"
	srcArgName         = "src"

	artifactNameArgName = "name"
)

type StoreServiceFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceName  kurtosis_backend_service.ServiceName
	src          string
	artifactName string
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
		return starlark.String(storeFilesFromServiceInstruction.artifactName), nil
	}
}

func NewStoreServiceFilesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceName kurtosis_backend_service.ServiceName, srcPath string, artifactName string, starlarkKwargs starlark.StringDict) *StoreServiceFilesInstruction {
	return &StoreServiceFilesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceName:    serviceName,
		src:            srcPath,
		artifactName:   artifactName,
		starlarkKwargs: starlarkKwargs,
	}
}

func newEmptyStoreServiceFilesInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition) *StoreServiceFilesInstruction {
	return &StoreServiceFilesInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceName:    "",
		src:            "",
		artifactName:   "",
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *StoreServiceFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *StoreServiceFilesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceNameArgName]), serviceNameArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[srcArgName]), srcArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[artifactNameArgName]), artifactNameArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), StoreServiceFilesBuiltinName, instruction.String(), args)
}

func (instruction *StoreServiceFilesInstruction) Execute(ctx context.Context) (*string, error) {
	artifactUuid, err := instruction.serviceNetwork.CopyFilesFromService(ctx, string(instruction.serviceName), instruction.src, instruction.artifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", instruction.src, instruction.serviceName)
	}
	instructionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", instruction.artifactName, artifactUuid)
	return &instructionResult, nil
}

func (instruction *StoreServiceFilesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(StoreServiceFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *StoreServiceFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if !environment.DoesServiceNameExist(instruction.serviceName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' with service name '%v' that does not exist", StoreServiceFilesBuiltinName, instruction.serviceName)
	}
	if environment.DoesArtifactNameExist(instruction.artifactName) {
		return stacktrace.NewError("There was an error validating '%v' as artifact name '%v' already exists", StoreServiceFilesBuiltinName, instruction.artifactName)
	}
	environment.AddArtifactName(instruction.artifactName)
	return nil
}

func (instruction *StoreServiceFilesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var serviceNameArg starlark.String
	var srcPathArg starlark.String
	var artifactNameArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceNameArgName, &serviceNameArg, srcArgName, &srcPathArg, artifactNameArgName, &artifactNameArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", StoreServiceFilesBuiltinName, args, kwargs)
	}

	instruction.starlarkKwargs[serviceNameArgName] = serviceNameArg
	instruction.starlarkKwargs[srcArgName] = srcPathArg
	instruction.starlarkKwargs[artifactNameArgName] = artifactNameArg
	instruction.starlarkKwargs.Freeze()

	serviceName, interpretationErr := kurtosis_instruction.ParseServiceName(serviceNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	artifactName, interpretationErr := kurtosis_instruction.ParseNonEmptyString(artifactNameArgName, artifactNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	instruction.serviceName = serviceName
	instruction.src = srcPath
	instruction.artifactName = artifactName
	return nil
}
