package resolver

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan"
)

type InstructionsPlanMask struct {
	readIdx                         int
	scheduledPersistableInstruction []*enclave_plan.EnclavePlanInstruction
	isValid                         bool
}

func NewInstructionsPlanMask(size int) *InstructionsPlanMask {
	return &InstructionsPlanMask{
		readIdx:                         0,
		scheduledPersistableInstruction: make([]*enclave_plan.EnclavePlanInstruction, size),
		isValid:                         true, // the mask is considered valid until it's proven to be invalid
	}
}

func (mask *InstructionsPlanMask) InsertAt(idx int, instruction *enclave_plan.EnclavePlanInstruction) {
	mask.scheduledPersistableInstruction[idx] = instruction
}

func (mask *InstructionsPlanMask) HasNext() bool {
	return mask.readIdx < len(mask.scheduledPersistableInstruction)
}

func (mask *InstructionsPlanMask) Next() (int, *enclave_plan.EnclavePlanInstruction) {
	instructionIdx := mask.readIdx
	scheduledInstruction := mask.scheduledPersistableInstruction[instructionIdx]
	mask.readIdx += 1
	return instructionIdx, scheduledInstruction
}

func (mask *InstructionsPlanMask) Size() int {
	return len(mask.scheduledPersistableInstruction)
}

func (mask *InstructionsPlanMask) MarkAsInvalid() {
	mask.isValid = false
}

func (mask *InstructionsPlanMask) IsValid() bool {
	return mask.isValid
}
