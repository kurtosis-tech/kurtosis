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
)

const (
	AddServiceBuiltinName = "add_service"

	ServiceNameArgName   = "service_name"
	ServiceConfigArgName = "config"
)

func NewAddService(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: AddServiceBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameArgName)
					},
				},
				{
					Name:              ServiceConfigArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*service_config.ServiceConfig],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// we just try to convert the configs here to validate their shape, to avoid code duplication
						// with Interpret
						if _, _, err := validateAndConvertConfigAndReadyConditions(value); err != nil {
							return err
						}
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &AddServiceCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				serviceName:   "",  // populated at interpretation time
				serviceConfig: nil, // populated at interpretation time

				resultUuid:      "",  // populated at interpretation time
				readyConditions: nil, // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName:   true,
			ServiceConfigArgName: true,
		},
	}
}

type AddServiceCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	serviceName     service.ServiceName
	serviceConfig   *kurtosis_core_rpc_api_bindings.ServiceConfig
	readyConditions *service_config.ReadyConditions

	resultUuid string
}

func (builtin *AddServiceCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	serviceConfig, err := builtin_argument.ExtractArgumentValue[*service_config.ServiceConfig](arguments, ServiceConfigArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceConfigArgName)
	}
	apiServiceConfig, readyConditions, interpretationErr := validateAndConvertConfigAndReadyConditions(serviceConfig)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	builtin.serviceName = service.ServiceName(serviceName.GoString())

	if isValidServiceName := service.ValidateServiceName(serviceName.GoString()); !isValidServiceName {
		return nil, startosis_errors.NewInterpretationError("Service name '%v' is invalid as it contains disallowed characters. Service names can only contain characters 'a-z', 'A-Z', '0-9', '-' & '_'", serviceName.GoString())
	}

	builtin.serviceConfig = apiServiceConfig
	builtin.readyConditions = readyConditions
	builtin.resultUuid, err = builtin.runtimeValueStore.CreateValue()
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to create runtime value to hold '%v' command return values", AddServiceBuiltinName)
	}

	returnValue, interpretationErr := makeAddServiceInterpretationReturnValue(serviceName, builtin.serviceConfig, builtin.resultUuid)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return returnValue, nil

}

func (builtin *AddServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validationErr := validateSingleService(validatorEnvironment, builtin.serviceName, builtin.serviceConfig); validationErr != nil {
		return validationErr
	}
	return nil
}

func (builtin *AddServiceCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(builtin.runtimeValueStore, builtin.serviceName, builtin.serviceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred replace a magic string in '%s' instruction arguments for service '%s'. Execution cannot proceed", AddServiceBuiltinName, builtin.serviceName)
	}
	startedService, err := builtin.serviceNetwork.StartService(ctx, replacedServiceName, replacedServiceConfig)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unexpected error occurred starting service '%s'", replacedServiceName)
	}

	if err := runServiceReadinessCheck(
		ctx,
		builtin.serviceNetwork,
		builtin.runtimeValueStore,
		replacedServiceName,
		builtin.readyConditions,
	); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while checking if service '%v' is ready", replacedServiceName)
	}

	fillAddServiceReturnValueWithRuntimeValues(startedService, builtin.resultUuid, builtin.runtimeValueStore)
	instructionResult := fmt.Sprintf("Service '%s' added with service UUID '%s'", replacedServiceName, startedService.GetRegistration().GetUUID())
	return instructionResult, nil
}

func validateAndConvertConfigAndReadyConditions(
	rawConfig starlark.Value,
) (*kurtosis_core_rpc_api_bindings.ServiceConfig, *service_config.ReadyConditions, *startosis_errors.InterpretationError) {
	config, ok := rawConfig.(*service_config.ServiceConfig)
	if !ok {
		return nil, nil, startosis_errors.NewInterpretationError("The '%s' argument is not a ServiceConfig (was '%s').", ConfigsArgName, reflect.TypeOf(rawConfig))
	}
	apiServiceConfig, interpretationErr := config.ToKurtosisType()
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	readyConditions, interpretationErr := config.GetReadyConditions()
	if interpretationErr != nil {
		return nil, nil, interpretationErr
	}

	return apiServiceConfig, readyConditions, nil
}
