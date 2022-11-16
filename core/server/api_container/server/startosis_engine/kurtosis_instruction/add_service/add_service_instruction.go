package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
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
	"regexp"
	"strings"
)

const (
	AddServiceBuiltinName = "add_service"

	serviceIdArgName = "service_id"

	serviceConfigArgName = "config"

	factNameArgName      = "fact_name"
	factNameSubgroupName = "fact_name"

	serviceIdSubgroupName = "service_id"
	allSubgroupName       = "all"
	kurtosisNamespace     = "kurtosis"
	// The placeholder format & regex should align
	ipAddressReplacementRegex             = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceIdArgName + ">" + service.ServiceIdRegexp + ")\\.ip_address\\}\\})"
	ipAddressReplacementPlaceholderFormat = "{{" + kurtosisNamespace + ":%v.ip_address}}"

	factReplacementRegex = "(?P<" + allSubgroupName + ">\\{\\{" + kurtosisNamespace + ":(?P<" + serviceIdArgName + ">" + service.ServiceIdRegexp + ")" + ":(?P<" + factNameArgName + ">" + service.ServiceIdRegexp + ")\\.fact\\}\\})"

	unlimitedMatches = -1
	singleMatch      = 1
	subExpNotFound   = -1
)

// The compiled regular expression to do IP address replacements
// Treat this as a constant
var (
	compiledIpAddressReplacementRegex = regexp.MustCompile(ipAddressReplacementRegex)
	compiledFactReplacementRegex      = regexp.MustCompile(factReplacementRegex)
)

func GenerateAddServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, factsEngine *facts_engine.FactsEngine) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		addServiceInstruction := newEmptyAddServiceInstruction(serviceNetwork, factsEngine, *shared_helpers.GetCallerPositionFromThread(thread))
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
	factsEngine    *facts_engine.FactsEngine

	position       kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceId     kurtosis_backend_service.ServiceID
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func newEmptyAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, factsEngine *facts_engine.FactsEngine, position kurtosis_instruction.InstructionPosition) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		factsEngine:    factsEngine,
		position:       position,
		starlarkKwargs: starlark.StringDict{},
	}
}

func NewAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position kurtosis_instruction.InstructionPosition, serviceId kurtosis_backend_service.ServiceID, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig, starlarkKwargs starlark.StringDict) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork: serviceNetwork,
		factsEngine:    nil,
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
		entryPointArgWithIPAddressAndFactsReplaced, err := replaceFactsInString(entryPointArgWithIPAddressReplaced, instruction.factsEngine)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing facts in entry point args for '%v'", entryPointArg)
		}
		entryPointArgs[index] = entryPointArgWithIPAddressAndFactsReplaced
	}

	cmdArgs := instruction.serviceConfig.CmdArgs
	for index, cmdArg := range cmdArgs {
		cmdArgWithIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(cmdArg, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in command args for '%v'", cmdArg)
		}
		cmdArgWithIPAddressAndFactsReplaced, err := replaceFactsInString(cmdArgWithIPAddressReplaced, instruction.factsEngine)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing facts in command args for '%v'", cmdArg)
		}
		cmdArgs[index] = cmdArgWithIPAddressAndFactsReplaced
	}

	envVars := instruction.serviceConfig.EnvVars
	for envVarName, envVarValue := range envVars {
		envVarValueWithIPAddressReplaced, err := shared_helpers.ReplaceIPAddressInString(envVarValue, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
		}
		envVarValueWithIPAddressAndFactsReplaced, err := replaceFactsInString(envVarValueWithIPAddressReplaced, instruction.factsEngine)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing facts in command args for '%v'", envVars)
		}
		envVars[envVarName] = envVarValueWithIPAddressAndFactsReplaced
	}

	return nil
}

func replaceIPAddressInString(originalString string, network service_network.ServiceNetwork, serviceIdForLogging string) (string, error) {
	matches := compiledIpAddressReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceIdMatchIndex := compiledIpAddressReplacementRegex.SubexpIndex(serviceIdSubgroupName)
		if serviceIdMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledIpAddressReplacementRegex.String())
		}
		serviceId := service.ServiceID(match[serviceIdMatchIndex])
		ipAddress, found := network.GetIPAddressForService(serviceId)
		if !found {
			return "", stacktrace.NewError("'%v' depends on the IP address of '%v' but we don't have any registrations for it", serviceIdForLogging, serviceId)
		}
		ipAddressStr := ipAddress.String()
		allMatchIndex := compiledIpAddressReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledIpAddressReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, ipAddressStr, singleMatch)
	}
	return replacedString, nil
}

func replaceFactsInString(originalString string, factsEngine *facts_engine.FactsEngine) (string, error) {
	matches := compiledFactReplacementRegex.FindAllStringSubmatch(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		serviceIdMatchIndex := compiledFactReplacementRegex.SubexpIndex(serviceIdSubgroupName)
		if serviceIdMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledIpAddressReplacementRegex.String())
		}
		factNameMatchIndex := compiledFactReplacementRegex.SubexpIndex(factNameSubgroupName)
		if factNameMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledIpAddressReplacementRegex.String())
		}
		factValues, err := factsEngine.FetchLatestFactValues(facts_engine.GetFactId(match[serviceIdMatchIndex], match[factNameMatchIndex]))
		if err != nil {
			return "", stacktrace.Propagate(err, "There was an error fetching fact value while replacing string '%v' '%v' ", match[serviceIdMatchIndex], match[factNameMatchIndex])
		}
		allMatchIndex := compiledFactReplacementRegex.SubexpIndex(allSubgroupName)
		if allMatchIndex == subExpNotFound {
			return "", stacktrace.NewError("There was an error in finding the sub group '%v' in regexp '%v'. This is a Kurtosis Bug", serviceIdSubgroupName, compiledIpAddressReplacementRegex.String())
		}
		allMatch := match[allMatchIndex]
		replacedString = strings.Replace(replacedString, allMatch, factValues[len(factValues)-1].GetStringValue(), singleMatch)
	}
	return replacedString, nil
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
	//  my_config = struct(private_port = port, other_arg = "blah")
	//  ```
	//  But we can do better than this defining our own structures:
	//  ```
	//  my_service_port = port_spec(port = 1234, protocol = "TCP") # port() is a Startosis defined struct
	//  my_config = config(port = port, other_arg = "blah")
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
