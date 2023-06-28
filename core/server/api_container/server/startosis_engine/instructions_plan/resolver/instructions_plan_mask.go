package resolver

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"

type InstructionsPlanMask struct {
	readIdx               int
	scheduledInstructions []*instructions_plan.ScheduledInstruction
	isValid               bool
}

func NewInstructionsPlanMask(size int) *InstructionsPlanMask {
	return &InstructionsPlanMask{
		readIdx:               0,
		scheduledInstructions: make([]*instructions_plan.ScheduledInstruction, size),
		isValid:               true, // the mask is considered valid until it's proven to be invalid
	}
}

func (mask *InstructionsPlanMask) InsertAt(idx int, instruction *instructions_plan.ScheduledInstruction) {
	mask.scheduledInstructions[idx] = instruction
}

func (mask *InstructionsPlanMask) HasNext() bool {
	return mask.readIdx < len(mask.scheduledInstructions)
}

func (mask *InstructionsPlanMask) Next() (int, *instructions_plan.ScheduledInstruction) {
	instructionIdx := mask.readIdx
	scheduledInstruction := mask.scheduledInstructions[instructionIdx]
	mask.readIdx += 1
	return instructionIdx, scheduledInstruction
}

func (mask *InstructionsPlanMask) Size() int {
	return len(mask.scheduledInstructions)
}

func (mask *InstructionsPlanMask) MarkAsInvalid() {
	mask.isValid = false
}

func (mask *InstructionsPlanMask) IsValid() bool {
	return mask.isValid
}
