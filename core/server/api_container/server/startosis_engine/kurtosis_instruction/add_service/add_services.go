package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/service_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

const (
	AddServicesBuiltinName = "add_services"

	ConfigsArgName   = "configs"
	ParallelismParam = "PARALLELISM"
)

func NewAddServices(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: AddServicesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ConfigsArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication
						// with Interpret
						if _, _, err := validateAndConvertConfigsAndReadyConditions(value); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &AddServicesCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceConfigs: nil, // populated at interpretation time

				resultUuids:     map[service.ServiceName]string{}, // populated at interpretation time
				readyConditions: nil,                              // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			// adding the entire config as a representative arg is kind of sad here as it might clutter the output,
			// but we don't really the choice
			ConfigsArgName: true,
		},
	}
}

type AddServicesCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig

	readyConditions map[service.ServiceName]*service_config.ReadyConditions

	resultUuids map[service.ServiceName]string
}

func (builtin *AddServicesCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	ServiceConfigsDict, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, ConfigsArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ConfigsArgName)
	}
	serviceConfigs, readyConditions, interpretationErr := validateAndConvertConfigsAndReadyConditions(ServiceConfigsDict)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.serviceConfigs = serviceConfigs
	builtin.readyConditions = readyConditions

	resultUuids, returnValue, interpretationErr := makeAddServicesInterpretationReturnValue(builtin.serviceConfigs, builtin.runtimeValueStore)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.resultUuids = resultUuids
	return returnValue, nil
}

func (builtin *AddServicesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	for serviceName, serviceConfig := range builtin.serviceConfigs {
		if err := validateSingleService(validatorEnvironment, serviceName, serviceConfig); err != nil {
			return err
		}
	}
	return nil
}

func (builtin *AddServicesCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	renderedServiceConfigs := make(map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig, len(builtin.serviceConfigs))
	parallelism, ok := ctx.Value(ParallelismParam).(int)
	if !ok {
		return "", stacktrace.NewError("An error occurred when getting parallelism level from execution context")
	}
	allServicesReadinessCheckFuncs := map[service.ServiceName]service_network.ServiceReadinessCheckFunc{}
	for serviceName, serviceConfig := range builtin.serviceConfigs {
		renderedServiceName, renderedServiceConfig, err := replaceMagicStrings(builtin.runtimeValueStore, serviceName, serviceConfig)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred replacing a magic string in '%s' instruction arguments for service: '%s'. Execution cannot proceed", AddServicesBuiltinName, serviceName)
		}
		renderedServiceConfigs[renderedServiceName] = renderedServiceConfig

		//filling services' readiness check func
		readyConditions, found := builtin.readyConditions[serviceName]
		if found {
			readinessCheckFunc, err := builtin.createServicesReadinessCheckFuncsForService(ctx, serviceName, readyConditions)
			if err != nil {
				return "", stacktrace.Propagate(err, "An error occurred creating service readiness check function for service '%v' with ready conditions '%+v'", serviceName, readyConditions)
			}
			allServicesReadinessCheckFuncs[renderedServiceName] = readinessCheckFunc
		}
	}

	startedServices, failedServices, err := builtin.serviceNetwork.StartServices(ctx, renderedServiceConfigs, parallelism, allServicesReadinessCheckFuncs)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unexpected error occurred starting a batch of services")
	}
	if len(failedServices) > 0 {
		failedServiceNames := make([]service.ServiceName, len(failedServices))
		idx := 0
		for failedServiceName := range failedServices {
			failedServiceNames[idx] = failedServiceName
			idx++
		}
		return "", stacktrace.NewError("Some errors occurred starting the following services: '%v'. The entire batch was rolled back an no service was started. Errors were: \n%v", failedServiceNames, failedServices)
	}

	instructionResult := strings.Builder{}
	instructionResult.WriteString(fmt.Sprintf("Successfully added the following '%d' services:", len(startedServices)))
	for serviceName, serviceObj := range startedServices {
		fillAddServiceReturnValueWithRuntimeValues(serviceObj, builtin.resultUuids[serviceName], builtin.runtimeValueStore)
		instructionResult.WriteString(fmt.Sprintf("\n  Service '%s' added with UUID '%s'", serviceName, serviceObj.GetRegistration().GetUUID()))

	}
	return instructionResult.String(), nil
}

func (builtin *AddServicesCapabilities) createServicesReadinessCheckFuncsForService(
	ctx context.Context,
	serviceName service.ServiceName,
	readyConditions *service_config.ReadyConditions,
) (service_network.ServiceReadinessCheckFunc, error) {
	serviceReadinessCheckFunc := func() error {
		if err := runServiceReadinessCheck(
			ctx,
			builtin.serviceNetwork,
			builtin.runtimeValueStore,
			serviceName,
			readyConditions,
		); err != nil {
			return stacktrace.Propagate(err, "An error occurred while checking if service '%v' is ready", serviceName)
		}
		return nil
	}

	return serviceReadinessCheckFunc, nil
}

func validateAndConvertConfigsAndReadyConditions(
	configs starlark.Value,
) (
	map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig,
	map[service.ServiceName]*service_config.ReadyConditions,
	*startosis_errors.InterpretationError,
) {
	configsDict, ok := configs.(*starlark.Dict)
	if !ok {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument should be a dictionary of matching each service name to their respective ServiceConfig object. Got '%s'", ConfigsArgName, reflect.TypeOf(configs))
	}
	if configsDict.Len() == 0 {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument should be a non empty dictionary", ConfigsArgName)
	}
	convertedServiceConfigs := map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig{}
	readyConditionsByServiceName := map[service.ServiceName]*service_config.ReadyConditions{}
	for _, serviceName := range configsDict.Keys() {
		serviceNameStr, isServiceNameAString := serviceName.(starlark.String)
		if !isServiceNameAString {
			return nil, nil, startosis_errors.NewInterpretationError("One key of the '%s' dictionary is not a string (was '%s'). Keys of this argument should correspond to service names, which should be strings", ConfigsArgName, reflect.TypeOf(serviceName))
		}

		dictValue, found, err := configsDict.Get(serviceName)
		if err != nil || !found {
			return nil, nil, startosis_errors.NewInterpretationError("Could not extract the value of the '%s' dictionary for key '%s'. This is Kurtosis bug", ConfigsArgName, serviceName)
		}
		serviceConfig, isDictValueAServiceConfig := dictValue.(*service_config.ServiceConfig)
		if !isDictValueAServiceConfig {
			return nil, nil, startosis_errors.NewInterpretationError("One value of the '%s' dictionary is not a ServiceConfig (was '%s'). Values of this argument should correspond to the config of the service to be added", ConfigsArgName, reflect.TypeOf(dictValue))
		}
		apiServiceConfig, interpretationErr := serviceConfig.ToKurtosisType()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		convertedServiceConfigs[service.ServiceName(serviceNameStr.GoString())] = apiServiceConfig

		readyConditions, interpretationErr := serviceConfig.GetReadyConditions()
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}

		readyConditionsByServiceName[service.ServiceName(serviceNameStr.GoString())] = readyConditions
	}
	return convertedServiceConfigs, readyConditionsByServiceName, nil
}

func makeAddServicesInterpretationReturnValue(serviceConfigs map[service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig, runtimeValueStore *runtime_value_store.RuntimeValueStore) (map[service.ServiceName]string, *starlark.Dict, *startosis_errors.InterpretationError) {
	servicesObjectDict := starlark.NewDict(len(serviceConfigs))
	resultUuids := map[service.ServiceName]string{}
	var err error
	for serviceName, serviceConfig := range serviceConfigs {
		serviceNameStr := starlark.String(serviceName)
		resultUuids[serviceName], err = runtimeValueStore.CreateValue()
		if err != nil {
			return nil, nil, startosis_errors.WrapWithInterpretationError(err, "Unable to create runtime value to hold '%v' command return values", AddServicesBuiltinName)
		}
		serviceObject, interpretationErr := makeAddServiceInterpretationReturnValue(serviceConfig, resultUuids[serviceName])
		if interpretationErr != nil {
			return nil, nil, interpretationErr
		}
		if err := servicesObjectDict.SetKey(serviceNameStr, serviceObject); err != nil {
			return nil, nil, startosis_errors.WrapWithInterpretationError(err, "Unable to generate the object that should be returned by the '%s' builtin", AddServicesBuiltinName)
		}
	}
	return resultUuids, servicesObjectDict, nil
}
