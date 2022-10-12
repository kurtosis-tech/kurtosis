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
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"regexp"
	"strings"
)

const (
	AddServiceBuiltinName = "add_service"

	serviceIdArgName     = "service_id"
	serviceConfigArgName = "service_config"

	ipAddressReplacementRegex = "(?P<all>\\{\\{kurtosis:(?P<service_id>" + service.ServiceIDRegex + ")\\.ip_address\\}\\})"
	serviceIdSubgroupName     = "service_id"
	allSubgroupName           = "all"

	unlimitedMatches = -1
	singleMatch      = 1
)

// The compiled regular expression to do IP address replacements
// Treat this as a constant
var compiledRegex = regexp.MustCompile(ipAddressReplacementRegex)

func GenerateAddServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		serviceId, serviceConfig, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		addServiceInstruction := NewAddServiceInstruction(serviceNetwork, getPosition(thread), serviceId, serviceConfig)
		*instructionsQueue = append(*instructionsQueue, addServiceInstruction)
		returnValue, interpretationError := makeAddServiceInterpretationReturnValue(serviceId, serviceConfig)
		if interpretationError != nil {
			return nil, interpretationError
		}
		return returnValue, nil
	}
}

type AddServiceInstruction struct {
	serviceNetwork service_network.ServiceNetwork

	position      kurtosis_instruction.InstructionPosition
	serviceId     kurtosis_backend_service.ServiceID
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func NewAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		position:       position,
		serviceId:      serviceId,
		serviceConfig:  serviceConfig,
	}
}

func (instruction *AddServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *AddServiceInstruction) GetCanonicalInstruction() string {
	// TODO(gb): implement when we need to return the canonicalized version of the script.
	//  Maybe there's a way to retrieve the serialized instruction from starlark-go
	return "add_service(...)"
}

func (instruction *AddServiceInstruction) Execute(ctx context.Context) error {
	err := instruction.replaceIPAddress()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred replacing IP Address with actual values in add service instruction for service '%v'", instruction.serviceId)
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
	return instruction.GetCanonicalInstruction()
}

func (instruction *AddServiceInstruction) replaceIPAddress() error {
	serviceIdStr := string(instruction.serviceId)
	entryPointArgs := instruction.serviceConfig.EntrypointArgs
	for index, value := range entryPointArgs {
		valueWithIPAddress, err := replaceIPAddressInString(value, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in entry point args")
		}
		entryPointArgs[index] = valueWithIPAddress
	}

	cmdArgs := instruction.serviceConfig.CmdArgs
	for index, value := range cmdArgs {
		valueWithIPAddress, err := replaceIPAddressInString(value, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in command args")
		}
		cmdArgs[index] = valueWithIPAddress
	}

	envVars := instruction.serviceConfig.EnvVars
	for key, value := range envVars {
		valueWithIPAddress, err := replaceIPAddressInString(value, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars")
		}
		envVars[key] = valueWithIPAddress
	}

	return nil
}

func replaceIPAddressInString(originalString string, network service_network.ServiceNetwork, serviceIdForLogging string) (string, error) {
	matches := compiledRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceIdMatchIndex := compiledRegex.SubexpIndex(serviceIdSubgroupName)
		serviceId := service.ServiceID(match[serviceIdMatchIndex])
		ipAddress, found := network.GetIPAddressForService(serviceId)
		ipAddressStr := ipAddress.String()
		if !found {
			return "", stacktrace.NewError("'%v' depends on the IP address of '%v' but we don't have any registrations for it", serviceIdForLogging, serviceId)
		}
		allMatchIndex := compiledRegex.SubexpIndex(allSubgroupName)
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, ipAddressStr, singleMatch)
	}
	return replacedString, nil
}

func makeAddServiceInterpretationReturnValue(serviceId service.ServiceID, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig) (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
	ports := serviceConfig.GetPrivatePorts()
	portSpecsDict := starlark.NewDict(len(ports))
	for portId, port := range ports {
		portNumber := starlark.MakeUint(uint(port.GetNumber()))
		portProtocol := starlark.String(port.GetProtocol().String())
		portSpec := kurtosis_types.NewPortSpec(portNumber, portProtocol)
		err := portSpecsDict.SetKey(starlark.String(portId), portSpec)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while creating a port spec for the add instruction return value")
		}
	}
	ipAddress := starlark.String(fmt.Sprintf("{{kurtosis:%v.ip_address}}", serviceId))
	returnValue := kurtosis_types.NewService(ipAddress, portSpecsDict)
	return returnValue, nil
}

func (instruction *AddServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	environment.AppendRequiredDockerImage(instruction.serviceConfig.ContainerImageName)
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (service.ServiceID, *kurtosis_core_rpc_api_bindings.ServiceConfig, *startosis_errors.InterpretationError) {
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
		return "", nil, startosis_errors.NewInterpretationError(err.Error())
	}

	serviceId, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return "", nil, interpretationErr
	}

	serviceConfig, interpretationErr := kurtosis_instruction.ParseServiceConfigArg(serviceConfigArg)
	if interpretationErr != nil {
		return "", nil, interpretationErr
	}
	return serviceId, serviceConfig, nil
}

func getPosition(thread *starlark.Thread) kurtosis_instruction.InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return *kurtosis_instruction.NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col)
}
