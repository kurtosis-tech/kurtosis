package enclave_plan

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
)

type EnclavePlan struct {
	enclavePlanInstructionRepository *EnclavePlanInstructionRepository
}

func NewEnclavePlan() *EnclavePlan {
	return &EnclavePlan{}
}

func (plan *EnclavePlan) Size() int {
	//TODO implement in the repository
	return 0
}

// GeneratePlan unwraps the plan into a list of instructions
func (plan *EnclavePlan) GeneratePlan() ([]*EnclavePlanInstruction, *startosis_errors.InterpretationError) {
	var generatedPlan []*EnclavePlanInstruction
	//TODO implement
	return generatedPlan, nil
}

/*
func (plan *EnclavePlan) AddInstruction(

	scheduledInstruction *instructions_plan.ScheduledInstruction,
	instructionCapabilities kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities,

	) *EnclavePlanInstruction {
		newScheduledInstructionUuid := scheduledInstruction.uuid
		newScheduledInstruction := NewScheduledInstruction(newScheduledInstructionUuid, scheduledInstruction.kurtosisInstruction, scheduledInstruction.returnedValue)
		newScheduledInstruction.Executed(scheduledInstruction.IsExecuted())
		newScheduledInstruction.ImportedFromCurrentEnclavePlan(scheduledInstruction.IsImportedFromCurrentEnclavePlan())

		plan.scheduledInstructionsIndex[newScheduledInstructionUuid] = newScheduledInstruction
		plan.instructionsSequence = append(plan.instructionsSequence, newScheduledInstructionUuid)
		return newScheduledInstruction
	}
*/
func newEnclavePlanInstruction(
	instruction kurtosis_instruction.KurtosisInstruction,
	instructionCapabilities kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities,
) *EnclavePlanInstruction {
	instructionStr := instruction.String()
	capabilities := instructionCapabilities.GetEnclavePlanCapabilities()

	enclavePlanInstruction := NewEnclavePlanInstruction(instructionStr, capabilities)

	return enclavePlanInstruction
}
