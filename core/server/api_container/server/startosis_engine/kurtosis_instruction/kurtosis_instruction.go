package kurtosis_instruction

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/validator_state"
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

	// Validate check if the current instruction applied to the current 'validatorState' is valid
	// If it is, it might modify validatorState, otherwise will return error
	Validate(validatorState *validator_state.StartosisValidatorState) error
}

func NewInstructionPosition(line int32, col int32) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
	}
}
