package kurtosis_instruction

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
)

type InstructionPosition struct {
	line     int32
	col      int32
	filename string
}

func NewInstructionPosition(line int32, col int32, filename string) *InstructionPosition {
	return &InstructionPosition{
		line:     line,
		col:      col,
		filename: filename,
	}
}

func (position *InstructionPosition) String() string {
	return fmt.Sprintf("%s[%d:%d]", position.filename, position.line, position.col)
}

func (position *InstructionPosition) ToAPIType() *kurtosis_core_rpc_api_bindings.StarlarkInstructionPosition {
	return binding_constructors.NewStarlarkInstructionPosition(position.filename, position.line, position.col)
}
