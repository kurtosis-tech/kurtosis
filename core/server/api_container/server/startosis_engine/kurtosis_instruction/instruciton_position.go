package kurtosis_instruction

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
