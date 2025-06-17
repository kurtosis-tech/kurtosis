package run_python

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

const (
	RunPythonBuiltinName = "run_python"
	descriptionStr       = "Running Python script"
)

func NewRunPython() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name:        RunPythonBuiltinName,
			Arguments:   []*builtin_argument.BuiltinArgument{},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunPythonCapabilities{
				description: descriptionStr,
			}
		},
		DefaultDisplayArguments: map[string]bool{},
	}
}

type RunPythonCapabilities struct {
	description string
}

func (builtin *RunPythonCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, descriptionStr)
	return starlark.None, nil
}

func (builtin *RunPythonCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, _ *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *RunPythonCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	return descriptionStr, nil
}

func (builtin *RunPythonCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, _ *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *RunPythonCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(RunPythonBuiltinName)
}

func (builtin *RunPythonCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	return nil
}

func (builtin *RunPythonCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *RunPythonCapabilities) UpdateDependencyGraph(dependencyGraph *dependency_graph.InstructionsDependencyGraph) error {
	// TODO: Implement dependency graph updates for run_python instruction
	return nil
}
