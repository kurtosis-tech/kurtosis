package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	AddServiceBuiltinName = "add_service"

	serviceIdArgName     = "service_id"
	serviceConfigArgName = "service_config"
)

func GenerateAddServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		addServiceInstruction := newEmptyAddServiceInstruction(serviceNetwork, *shared_helpers.GetCallerPositionFromThread(thread))
		if interpretationError := addServiceInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, addServiceInstruction)
		returnValue, interpretationError := addServiceInstruction.makeAddServiceInterpretationReturnValue()
		if interpretationError != nil {
			return nil, interpretationError
		}
		return returnValue, nil
	}
}

type AddServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position       kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId     kurtosis_backend_service.ServiceID
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func newEmptyAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		starlarkKwargs: starlark.StringDict{},
	}
}

func NewAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig, starlarkKwargs starlark.StringDict) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		serviceConfig:  serviceConfig,
		starlarkKwargs: starlarkKwargs,
	}
}

func (instruction *AddServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *AddServiceInstruction) GetCanonicalInstruction() string {
	return shared_helpers.MultiLineCanonicalizer.CanonicalizeInstruction(AddServiceBuiltinName, instruction.starlarkKwargs, &instruction.position)
}

func (instruction *AddServiceInstruction) Execute(ctx context.Context, environment *startosis_executor.ExecutionEnvironment) error {
	err := instruction.replaceIPAddress()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred replacing IP Address with actual values in add service instruction for service '%v'", instruction.serviceId)
	}

	for maybeArtifactUuidMagicStringValue, pathOnContainer := range instruction.serviceConfig.FilesArtifactMountpoints {
		artifactUuidActualValue, err := shared_helpers.ReplaceArtifactUuidMagicStringWithValue(maybeArtifactUuidMagicStringValue, string(instruction.serviceId), environment)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while replacing the placeholder '%v' artifact uuid with actual value", maybeArtifactUuidMagicStringValue)
		}
		delete(instruction.serviceConfig.FilesArtifactMountpoints, maybeArtifactUuidMagicStringValue)
		instruction.serviceConfig.FilesArtifactMountpoints[artifactUuidActualValue] = pathOnContainer
	}

	serviceConfigMap := map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig{
		instruction.serviceId: instruction.serviceConfig,
	}

	// TODO Pull partition from user in Starlark
	serviceSuccessful, serviceFailed, err := instruction.serviceNetwork.StartServices(ctx, serviceConfigMap, service_network_types.PartitionID(""))
	if err != nil {
		return stacktrace.Propagate(err, "Failed adding service to enclave with an unexpected error")
	}
	if failure, found := serviceFailed[instruction.serviceId]; found {
		return stacktrace.Propagate(failure, "Failed adding service to enclave")
	}
	if _, found := serviceSuccessful[instruction.serviceId]; !found {
		return stacktrace.NewError("Service wasn't accounted as failed nor successfully added. This is a product bug")
	}
	return nil
}

func (instruction *AddServiceInstruction) String() string {
	return shared_helpers.SingleLineCanonicalizer.CanonicalizeInstruction(AddServiceBuiltinName, instruction.starlarkKwargs, &instruction.position)
}

func (instruction *AddServiceInstruction) replaceIPAddress() error {
	serviceIdStr := string(instruction.serviceId)
	entryPointArgs := instruction.serviceConfig.EntrypointArgs
	for index, entryPointArg := range entryPointArgs {
		entryPointArgWithIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(entryPointArg, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in entry point args for '%v'", entryPointArg)
		}
		entryPointArgs[index] = entryPointArgWithIPAddressReplaced
	}

	cmdArgs := instruction.serviceConfig.CmdArgs
	for index, cmdArg := range cmdArgs {
		cmdArgWithIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(cmdArg, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in command args for '%v'", cmdArg)
		}
		cmdArgs[index] = cmdArgWithIPAddressReplaced
	}

	envVars := instruction.serviceConfig.EnvVars
	for envVarName, envVarValue := range envVars {
		envVarValueWithIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(envVarValue, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
		}
		envVars[envVarName] = envVarValueWithIPAddressReplaced
	}

	return nil
}

func (instruction *AddServiceInstruction) makeAddServiceInterpretationReturnValue() (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
	ports := instruction.serviceConfig.GetPrivatePorts()
	portSpecsDict := starlark.NewDict(len(ports))
	for portId, port := range ports {
		portNumber := starlark.MakeUint(uint(port.GetNumber()))
		portProtocol := starlark.String(port.GetProtocol().String())
		portSpec := kurtosis_types.NewPortSpec(portNumber, portProtocol)
		if err := portSpecsDict.SetKey(starlark.String(portId), portSpec); err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while creating a port spec for values (number: '%v', port: '%v') the add instruction return value", portNumber, portProtocol)
		}
	}
	ipAddress := starlark.String(fmt.Sprintf(shared_helpers.IpAddressReplacementPlaceholderFormat, instruction.serviceId))
	returnValue := kurtosis_types.NewService(ipAddress, portSpecsDict)
	return returnValue, nil
}

func (instruction *AddServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if environment.DoesServiceIdExist(instruction.serviceId) {
		return stacktrace.NewError("There was an error validating add service as service ID '%v' already exists", instruction.serviceId)
	}
	environment.AddServiceId(instruction.serviceId)
	environment.AppendRequiredDockerImage(instruction.serviceConfig.ContainerImageName)
	return nil
}

func (instruction *AddServiceInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	// TODO(gb): Right now, we expect the Startosis script to be very "untyped" like:
	//  ```startosis
	//  my_service_port = struct(port = 1234, protocol = "TCP")
	//  my_service_config = struct(private_port = port, other_arg = "blah")
	//  ```
	//  But we can do better than this defining our own structures:
	//  ```
	//  my_service_port = port_spec(port = 1234, protocol = "TCP") # port() is a Startosis defined struct
	//  my_service_config = service_config(port = port, other_arg = "blah")
	//  ```
	//  With custom types, we can parse the args directly to our own Go types and potentially isolate the checks

	var serviceIdArg starlark.String
	var serviceConfigArg *starlarkstruct.Struct
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, serviceConfigArgName, &serviceConfigArg); err != nil {
		return startosis_errors.NewInterpretationError(err.Error())
	}
	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	instruction.starlarkKwargs[serviceConfigArgName] = serviceConfigArg
	instruction.starlarkKwargs.Freeze()

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	serviceConfig, interpretationErr := kurtosis_instruction.ParseServiceConfigArg(serviceConfigArg)
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.serviceId = serviceId
	instruction.serviceConfig = serviceConfig
	return nil
}
