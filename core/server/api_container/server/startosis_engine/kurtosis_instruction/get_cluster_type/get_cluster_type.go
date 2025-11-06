package get_cluster_type

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"go.starlark.net/starlark"
)

const (
	GetClusterTypeBuiltinName = "get_cluster_type"

	descriptionFormatStr = "Fetching cluster type '%v'"
)

func NewGetClusterType(kurtosisBackendType args.KurtosisBackendType) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name:        GetClusterTypeBuiltinName,
			Arguments:   []*builtin_argument.BuiltinArgument{},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &GetClusterTypeCapabilities{
				clusterType: kurtosisBackendType.String(),
				description: fmt.Sprintf(descriptionFormatStr, kurtosisBackendType.String()),
			}
		},
		DefaultDisplayArguments: map[string]bool{},
	}
}

type GetClusterTypeCapabilities struct {
	clusterType string
	description string
}

func (builtin *GetClusterTypeCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	return starlark.String(builtin.clusterType), nil
}

func (builtin *GetClusterTypeCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *GetClusterTypeCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// Note this is a no-op
	return fmt.Sprintf("Fetched cluster type '%v'", builtin.clusterType), nil
}

func (builtin *GetClusterTypeCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *GetClusterTypeCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(GetClusterTypeBuiltinName)
}

func (builtin *GetClusterTypeCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	// get cluster type does not affect the planYaml
	return nil
}

func (builtin *GetClusterTypeCapabilities) UpdateDependencyGraph(instructionUuid types.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionDependencyGraph) error {
	// has no effect no dependencies
	return nil
}

func (builtin *GetClusterTypeCapabilities) Description() string {
	return builtin.description
}
