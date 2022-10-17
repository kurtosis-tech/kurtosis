package kurtosis_instruction

import "fmt"

const (
	placeholderFormat = "{{kurtosis:%v:%v.%v}}"
	regexFormat       = "{{kurtosis:[0-9]+:[0-9]+.%v}}"
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

func GetRegularExpressionForInstruction(suffix string) string {
	return fmt.Sprintf(regexFormat, suffix)
}
