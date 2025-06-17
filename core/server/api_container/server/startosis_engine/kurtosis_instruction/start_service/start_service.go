package start_service

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	StartServiceBuiltinName = "start_service"

	ServiceNameArgName = "name"
)

const (
	descriptionFormatStr = "Starting service '%v'"
)

func NewStartService(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: StartServiceBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						// TODO: when #903 is merged, validate service name are non emtpy string
						return nil
					},
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &StartServiceCapabilities{
				serviceNetwork: serviceNetwork,

				serviceName: "", // populated at interpretation time
				description: "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName: true,
		},
	}
}

type StartServiceCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	serviceName service.ServiceName

	description string
}

func (builtin *StartServiceCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	builtin.serviceName = service.ServiceName(serviceName.GoString())
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName))
	return starlark.None, nil
}

func (builtin *StartServiceCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesServiceNameExist(builtin.serviceName) == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service name '%v' doesn't exist", StartServiceBuiltinName, builtin.serviceName)
	}
	return nil
}

func (builtin *StartServiceCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	err := builtin.serviceNetwork.StartService(ctx, string(builtin.serviceName))
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed starting service with unexpected error")
	}
	instructionResult := fmt.Sprintf("Service '%s' started", builtin.serviceName)
	return instructionResult, nil
}

func (builtin *StartServiceCapabilities) TryResolveWith(_ bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	return enclave_structure.InstructionIsNotResolvableAbort
}

func (builtin *StartServiceCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(
		StartServiceBuiltinName,
	).AddServiceName(
		builtin.serviceName,
	)
}

func (builtin *StartServiceCapabilities) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	// start services doesn't affect the plan
	return nil
}

func (builtin *StartServiceCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *StartServiceCapabilities) UpdateDependencyGraph(dependencyGraph *dependency_graph.InstructionsDependencyGraph) error {
	// TODO: Implement dependency graph updates for start_service instruction
	return nil
}
