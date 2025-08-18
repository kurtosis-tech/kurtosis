package kurtosis_plan_instruction

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstructionCapabilities interface {
	Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError)

	Validate(arguments *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError

	Execute(ctx context.Context, arguments *builtin_argument.ArgumentValuesSet) (string, error)

	TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus

	// FillPersistableAttributes adds to the builder the attributes of the instruction that needs to be persisted to the
	// enclave database to power idempotent runs.
	FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder)

	// Description Brief description of the instruction based on its contents
	Description() string

	// UpdatePlan applies the effect of this instruction capabilities onto the yaml representation of the instruction plan
	UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error

	// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction
	UpdateDependencyGraph(instructionUuid dependency_graph.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionsDependencyGraph) error
}
