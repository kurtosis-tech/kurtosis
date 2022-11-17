package kurtosis_instruction

import (
	"fmt"
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
