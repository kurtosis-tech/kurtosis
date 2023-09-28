package plan_module

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	planModuleName = "plan"
)

func PlanModule(
	instructionsPlan *instructions_plan.InstructionsPlan,
	enclaveComponents *enclave_structure.EnclaveComponents,
	starlarkValueSerde *kurtosis_types.StarlarkValueSerde,
	instructionsPlanMask *resolver.InstructionsPlanMask,
	kurtosisPlanInstructions []*kurtosis_plan_instruction.KurtosisPlanInstruction,
) *starlarkstruct.Module {
	moduleBuiltins := starlark.StringDict{}
	for _, planInstruction := range kurtosisPlanInstructions {
		logrus.Infof("[LEO-DEBUG-3] planInstruction: %+v", planInstruction)
		wrappedPlanInstruction := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(planInstruction, enclaveComponents, starlarkValueSerde, instructionsPlanMask, instructionsPlan)
		logrus.Infof("[LEO-DEBUG-3] wrappedPlanInstruction: %+v", wrappedPlanInstruction)
		moduleBuiltins[planInstruction.GetName()] = starlark.NewBuiltin(planInstruction.GetName(), wrappedPlanInstruction.CreateBuiltin())
	}

	return &starlarkstruct.Module{
		Name:    planModuleName,
		Members: moduleBuiltins,
	}
}
