package resolver

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"

type InstructionsPlanMask struct {
	readIdx               int
	scheduledInstructions []*instructions_plan.ScheduledInstruction
}

func NewInstructionsPlanMask(size int) *InstructionsPlanMask {
	return &InstructionsPlanMask{
		readIdx:               0,
		scheduledInstructions: make([]*instructions_plan.ScheduledInstruction, size),
	}
}

func (mask *InstructionsPlanMask) InsertAt(idx int, instruction *instructions_plan.ScheduledInstruction) {
	mask.scheduledInstructions[idx] = instruction
}

func (mask *InstructionsPlanMask) HasNext() bool {
	return mask.readIdx < len(mask.scheduledInstructions)
}

func (mask *InstructionsPlanMask) Next() *instructions_plan.ScheduledInstruction {
	scheduledInstruction := mask.scheduledInstructions[mask.readIdx]
	mask.readIdx += 1
	return scheduledInstruction
}

func (mask *InstructionsPlanMask) Size() int {
	return len(mask.scheduledInstructions)
}
