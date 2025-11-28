package kurtosis_print

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	PrintBuiltinName = "print"

	PrintArgName = "msg"

	descriptionFormatStr = "Printing a message"
)

func NewPrint(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: PrintBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PrintArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &PrintCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				msg:         nil, // populated at interpretation time
				description: "",  // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			PrintArgName: true,
		},
	}
}

type PrintCapabilities struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	msg         starlark.Value
	description string
}

func (builtin *PrintCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	msg, err := builtin_argument.ExtractArgumentValue[starlark.Value](arguments, PrintArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", PrintArgName)
	}
	builtin.msg = msg
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, descriptionFormatStr)
	return starlark.None, nil
}

func (builtin *PrintCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// nothing to do for now
	// TODO(gb): maybe in the future validate that if we're using a magic string, it points to something real
	return nil
}

func (builtin *PrintCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	var serializedMsg string
	switch arg := builtin.msg.(type) {
	case starlark.String:
		serializedMsg = arg.GoString()
	default:
		serializedMsg = arg.String()
	}
	maybeSerializedArgsWithRuntimeValue, err := magic_string_helper.ReplaceRuntimeValueInString(serializedMsg, builtin.runtimeValueStore)
	if err != nil {
		return "", stacktrace.Propagate(err, "Error replacing runtime value '%v'", serializedMsg)
	}
	return maybeSerializedArgsWithRuntimeValue, nil
}

func (builtin *PrintCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *PrintCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(PrintBuiltinName)
}

func (builitin *PrintCapabilities) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	// print does not affect the plan
	return nil
}

func (builtin *PrintCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *PrintCapabilities) UpdateDependencyGraph(instructionUuid types.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionDependencyGraph) error {
	shortDescriptor := fmt.Sprintf("print(%s)", builtin.msg.String())
	dependencyGraph.UpdateInstructionShortDescriptor(instructionUuid, shortDescriptor)

	dependencyGraph.AddPrintInstruction(instructionUuid)

	dependencyGraph.ConsumesAnyRuntimeValuesInString(instructionUuid, builtin.msg.String())
	return nil
}
