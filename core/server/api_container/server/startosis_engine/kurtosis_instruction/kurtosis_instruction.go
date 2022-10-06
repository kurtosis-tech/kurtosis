package kurtosis_instruction

import (
	"context"
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
}

func NewInstructionPosition(line int32, col int32) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
	}
}
