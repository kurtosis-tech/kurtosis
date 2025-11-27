package resolver

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
)

type InstructionsPlanMask struct {
	readIdx                 int
	enclavePlanInstructions []*enclave_plan_persistence.EnclavePlanInstruction
	isValid                 bool
}

func NewInstructionsPlanMask(size int) *InstructionsPlanMask {
	return &InstructionsPlanMask{
		readIdx:                 0,
		enclavePlanInstructions: make([]*enclave_plan_persistence.EnclavePlanInstruction, size),
		isValid:                 true, // the mask is considered valid until it's proven to be invalid
	}
}

func (mask *InstructionsPlanMask) InsertAt(idx int, instruction *enclave_plan_persistence.EnclavePlanInstruction) {
	mask.enclavePlanInstructions[idx] = instruction
}

func (mask *InstructionsPlanMask) HasNext() bool {
	return mask.readIdx < len(mask.enclavePlanInstructions)
}

func (mask *InstructionsPlanMask) Next() (int, *enclave_plan_persistence.EnclavePlanInstruction) {
	instructionIdx := mask.readIdx
	scheduledInstruction := mask.enclavePlanInstructions[instructionIdx]
	mask.readIdx += 1
	return instructionIdx, scheduledInstruction
}

func (mask *InstructionsPlanMask) Size() int {
	return len(mask.enclavePlanInstructions)
}

func (mask *InstructionsPlanMask) MarkAsInvalid() {
	mask.isValid = false
}

func (mask *InstructionsPlanMask) IsValid() bool {
	return mask.isValid
}
