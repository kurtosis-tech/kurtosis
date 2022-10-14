package kurtosis_instruction

import "fmt"

const (
	placeholderFormat = "{{kurtosis:%v:%v.%v}}"
)

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

func (ip *InstructionPosition) MagicString(suffix string) string {
	return fmt.Sprintf(placeholderFormat, ip.line, ip.col, suffix)
}
