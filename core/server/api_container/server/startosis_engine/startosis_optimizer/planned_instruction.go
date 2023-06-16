package startosis_optimizer

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_optimizer/graph"
)

type PlannedInstruction struct {
	isOutOfScope bool

	isSkipped bool

	uuid graph.NodeUuid
	hash graph.InstructionHash

	instruction kurtosis_instruction.KurtosisInstruction
}

func NewPlannedInstruction(uuid graph.NodeUuid, hash graph.InstructionHash, instruction kurtosis_instruction.KurtosisInstruction, isOutOfScope bool, isSkipped bool) *PlannedInstruction {
	return &PlannedInstruction{
		uuid:         uuid,
		hash:         hash,
		isOutOfScope: isOutOfScope,
		isSkipped:    isSkipped,
		instruction:  instruction,
	}
}

func (instruction *PlannedInstruction) GetUuid() graph.NodeUuid {
	return instruction.uuid
}

func (instruction *PlannedInstruction) GetInstruction() kurtosis_instruction.KurtosisInstruction {
	return instruction.instruction
}

func (instruction *PlannedInstruction) GetHash() graph.InstructionHash {
	return instruction.hash
}

func (instruction *PlannedInstruction) IsOutOfScope() bool {
	return instruction.isOutOfScope
}

func (instruction *PlannedInstruction) IsSkipped() bool {
	return instruction.isSkipped
}
