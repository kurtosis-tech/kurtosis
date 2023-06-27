package plan_module

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	planModuleName = "plan"
)

func PlanModule(
	instructionsPlan *instructions_plan.InstructionsPlan,
	instructionsPlanMask *resolver.InstructionsPlanMask,
	kurtosisPlanInstructions []*kurtosis_plan_instruction.KurtosisPlanInstruction,
) *starlarkstruct.Module {
	moduleBuiltins := starlark.StringDict{}
	for _, planInstruction := range kurtosisPlanInstructions {
		wrappedPlanInstruction := kurtosis_plan_instruction.NewKurtosisPlanInstructionWrapper(planInstruction, instructionsPlanMask, instructionsPlan)
		moduleBuiltins[planInstruction.GetName()] = starlark.NewBuiltin(planInstruction.GetName(), wrappedPlanInstruction.CreateBuiltin())
	}

	return &starlarkstruct.Module{
		Name:    planModuleName,
		Members: moduleBuiltins,
	}
}
