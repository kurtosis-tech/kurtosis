package instructions_plan

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

// InstructionsPlan is the object to store a sequence of instructions which forms a "plan" for the enclave.
// Right now, the object is fairly simple in the sense of it just stores literally the sequence of instructions, and
// a bit of metadata about each instruction (i.e. whether it has been executed of not, for example)
// The plan is "append-only", i.e. when an instruction is added, it cannot be removed.
// The only read method is GeneratePlan unwraps the plan into an actual list of instructions that can be submitted to
// the executor.
type InstructionsPlan struct {
	indexOfFirstInstruction int

	scheduledInstructionsIndex map[ScheduledInstructionUuid]*ScheduledInstruction

	instructionsSequence []ScheduledInstructionUuid
}

func NewInstructionsPlan() *InstructionsPlan {
	return &InstructionsPlan{
		indexOfFirstInstruction:    0,
		scheduledInstructionsIndex: map[ScheduledInstructionUuid]*ScheduledInstruction{},
		instructionsSequence:       []ScheduledInstructionUuid{},
	}
}

func (plan *InstructionsPlan) PartialClone(indexOfFirstInstructionToNotClone int) (*InstructionsPlan, error) {
	newPlan := NewInstructionsPlan()
	for idx, instructionUuid := range plan.instructionsSequence {
		if idx >= indexOfFirstInstructionToNotClone {
			return newPlan, nil
		}
		instructionToTransfer, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return nil, stacktrace.NewError("Instruction with UUID '%s' was part of the instruction sequence but could not be found in the instruction index. This is a Kurtosis internal issue", instructionUuid)
		}
		newPlan.AddScheduledInstruction(instructionToTransfer)
	}
	return newPlan, nil
}

func (plan *InstructionsPlan) SetIndexOfFirstInstruction(indexOfFirstInstruction int) {
	plan.indexOfFirstInstruction = indexOfFirstInstruction
}

func (plan *InstructionsPlan) GetIndexOfFirstInstruction() int {
	return plan.indexOfFirstInstruction
}

func (plan *InstructionsPlan) AddInstruction(instruction kurtosis_instruction.KurtosisInstruction, returnedValue starlark.Value) (*ScheduledInstruction, error) {
	generatedUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to generate a random UUID for instruction '%s' to add it to the plan", instruction.String())
	}

	scheduledInstructionUuid := ScheduledInstructionUuid(generatedUuid)
	scheduledInstruction := NewScheduledInstruction(scheduledInstructionUuid, instruction, returnedValue)

	plan.scheduledInstructionsIndex[scheduledInstructionUuid] = scheduledInstruction
	plan.instructionsSequence = append(plan.instructionsSequence, scheduledInstructionUuid)
	return scheduledInstruction, nil
}

func (plan *InstructionsPlan) AddScheduledInstruction(scheduledInstruction *ScheduledInstruction) *ScheduledInstruction {
	newScheduledInstructionUuid := scheduledInstruction.uuid
	newScheduledInstruction := NewScheduledInstruction(newScheduledInstructionUuid, scheduledInstruction.kurtosisInstruction, scheduledInstruction.returnedValue)
	newScheduledInstruction.Executed(scheduledInstruction.IsExecuted())

	plan.scheduledInstructionsIndex[newScheduledInstructionUuid] = newScheduledInstruction
	plan.instructionsSequence = append(plan.instructionsSequence, newScheduledInstructionUuid)
	return newScheduledInstruction
}

// GeneratePlan unwraps the plan into a list of instructions
func (plan *InstructionsPlan) GeneratePlan() ([]*ScheduledInstruction, *startosis_errors.InterpretationError) {
	var generatedPlan []*ScheduledInstruction
	for _, instructionUuid := range plan.instructionsSequence {
		instruction, found := plan.scheduledInstructionsIndex[instructionUuid]
		if !found {
			return nil, startosis_errors.NewInterpretationError("Unexpected error generating the Kurtosis Instructions plan. Instruction with UUID '%s' was scheduled but could not be found in Kurtosis instruction index", instructionUuid)
		}
		generatedPlan = append(generatedPlan, instruction)
	}
	return generatedPlan, nil
}

func (plan *InstructionsPlan) GetInstructionsSequence() []ScheduledInstructionUuid {
	return plan.instructionsSequence
}

func (plan *InstructionsPlan) Size() int {
	return len(plan.instructionsSequence)
}
