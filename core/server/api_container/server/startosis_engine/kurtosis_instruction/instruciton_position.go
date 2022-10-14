package kurtosis_instruction

import "fmt"

type InstructionPosition struct {
	line int32
	col  int32
}

func NewInstructionPosition(line int32, col int32) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
	}
}

func (ip *InstructionPosition) String() string {
	return fmt.Sprintf("%v:%v", ip.line, ip.col)
}
