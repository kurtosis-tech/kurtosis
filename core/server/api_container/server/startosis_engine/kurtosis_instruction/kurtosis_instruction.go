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

	// UpdateEnvironment mutates environment to reflect how Kurtosis would look like after this instruction
	UpdateEnvironment(environment *startosis_validator.ValidatorEnvironment)
}

func NewInstructionPosition(line int32, col int32) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
	}
}
