package instructions_plan

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/types"
	"go.starlark.net/starlark"
)

// ScheduledInstruction is a wrapper around a KurtosisInstruction to specify that the instruction is part of an
// InstructionPlan. The instruction plan can either be the current enclave plan (which has been executed) or a newly
// generated plan from the latest interpretation.
// In any case, the ScheduledInstructionUuid stores the result object from the interpretation of the instruction,
// as well as a flag to track whether this instruction was already executed or not.
type ScheduledInstruction struct {
	uuid types.ScheduledInstructionUuid

	kurtosisInstruction kurtosis_instruction.KurtosisInstruction

	returnedValue starlark.Value

	executed bool
}

func NewScheduledInstruction(uuid types.ScheduledInstructionUuid, kurtosisInstruction kurtosis_instruction.KurtosisInstruction, returnedValue starlark.Value) *ScheduledInstruction {
	return &ScheduledInstruction{
		uuid:                uuid,
		kurtosisInstruction: kurtosisInstruction,
		returnedValue:       returnedValue,
		executed:            false,
	}
}

func (instruction *ScheduledInstruction) GetUuid() types.ScheduledInstructionUuid {
	return instruction.uuid
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
