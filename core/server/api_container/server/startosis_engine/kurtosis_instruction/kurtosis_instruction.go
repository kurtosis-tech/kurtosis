package kurtosis_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
)

type InstructionPosition struct {
	line int32
	col  int32
}

type KurtosisInstruction interface {
	GetPositionInOriginalScript() *InstructionPosition

	GetCanonicalInstruction() string

	Execute(ctx context.Context) error

	// String is only for easy printing in logs and error messages.
	// Most of the time it will just call GetCanonicalInstruction()
	String() string

	// ValidateAndUpdateEnvironment validates if the instruction can be applied to an environment, and mutates that
	// environment to reflect how Kurtosis would look like after this instruction is successfully executed.
	ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error
}

func NewInstructionPosition(line int32, col int32) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
	}
}
