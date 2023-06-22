package instructions_plan

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"go.starlark.net/starlark"
)

type ScheduledInstructionUuid string

type ScheduledInstruction struct {
	uuid ScheduledInstructionUuid

	kurtosisInstruction kurtosis_instruction.KurtosisInstruction

	returnedValue starlark.Value

	executed bool
}

func NewScheduledInstruction(uuid ScheduledInstructionUuid, kurtosisInstruction kurtosis_instruction.KurtosisInstruction, returnedValue starlark.Value) *ScheduledInstruction {
	return &ScheduledInstruction{
		uuid:                uuid,
		kurtosisInstruction: kurtosisInstruction,
		returnedValue:       returnedValue,
		executed:            false,
	}
}

func (instruction *ScheduledInstruction) GetInstruction() kurtosis_instruction.KurtosisInstruction {
	return instruction.kurtosisInstruction
}

func (instruction *ScheduledInstruction) GetReturnedValue() starlark.Value {
	return instruction.returnedValue
}

func (instruction *ScheduledInstruction) Executed(isExecuted bool) *ScheduledInstruction {
	instruction.executed = isExecuted
	return instruction
}

func (instruction *ScheduledInstruction) IsExecuted() bool {
	return instruction.executed
}
