package instructions_plan

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type InstructionsPlan struct {
	scheduledInstructionsIndex map[ScheduledInstructionUuid]*ScheduledInstruction

	instructionsQueue []ScheduledInstructionUuid
}

func NewInstructionsPlan() *InstructionsPlan {
	return &InstructionsPlan{
		scheduledInstructionsIndex: map[ScheduledInstructionUuid]*ScheduledInstruction{},
	}
}

func (plan *InstructionsPlan) AddInstruction(instruction kurtosis_instruction.KurtosisInstruction, returnedValue starlark.Value) error {
	generatedUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to generate a random UUID for instruction '%s' to add it to the plan", instruction.String())
	}

	scheduledInstructionUuid := ScheduledInstructionUuid(generatedUuid)
	scheduledInstruction := NewScheduledInstruction(scheduledInstructionUuid, instruction, returnedValue)

	plan.scheduledInstructionsIndex[scheduledInstructionUuid] = scheduledInstruction
	plan.instructionsQueue = append(plan.instructionsQueue, scheduledInstructionUuid)
	return nil
}

func (plan *InstructionsPlan) AddScheduledInstruction(scheduledInstruction *ScheduledInstruction) *ScheduledInstruction {
	newScheduledInstructionUuid := scheduledInstruction.uuid
	newScheduledInstruction := NewScheduledInstruction(newScheduledInstructionUuid, scheduledInstruction.kurtosisInstruction, scheduledInstruction.returnedValue)
	newScheduledInstruction.Executed(scheduledInstruction.IsExecuted())
	newScheduledInstruction.ImportedFromPreviousEnclavePlan(scheduledInstruction.IsImportedFromPreviousEnclavePlan())

	plan.scheduledInstructionsIndex[newScheduledInstructionUuid] = newScheduledInstruction
	plan.instructionsQueue = append(plan.instructionsQueue, newScheduledInstructionUuid)
	return newScheduledInstruction
}

func (plan *InstructionsPlan) GeneratePlan() ([]*ScheduledInstruction, *startosis_errors.InterpretationError) {
	var generatedPlan []*ScheduledInstruction
	for _, instructionUuid := range plan.instructionsQueue {
		instruction, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return nil, startosis_errors.NewInterpretationError("Unexpected error generating the Kurtosis Instructions plan. Instruction with UUID '%s' was scheduled but could not be found in Kurtosis instruction index", instructionUuid)
		}
		generatedPlan = append(generatedPlan, instruction)
	}
	return generatedPlan, nil
}

func (plan *InstructionsPlan) Size() int {
	return len(plan.instructionsQueue)
}
