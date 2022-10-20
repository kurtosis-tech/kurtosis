package kurtosis_instruction

import (
	"fmt"
)

const (
	// these two are here so that they align
	// the placeholder is the magic string that gets assigned during interpretation
	// the regex gets used during execution to find magic strings to replace with
	// actual values
	magicStringFormat = "{{kurtosis:%v:%v.%v}}"
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

// MagicString the magic string allows us to identify an instruction that doesn't
// have any other obvious identifiers, we take the line & column number of the instruction
// and add a suffix to it to track what is returned by that instruction
// If an instruction returns multiple things, use different suffixes for each object
// This string gets assigned to the object during interpretation time and replaced during
// execution time
func (position *InstructionPosition) MagicString(suffix string) string {
	return fmt.Sprintf(magicStringFormat, position.line, position.col, suffix)
}

// GetRegularExpressionForInstruction this function allows you to get a regular expression
// that matches an instruction derived magic string, just pass in the same suffix
// that you used to generate the magic string with
func GetRegularExpressionForInstruction(suffix string) string {
	return fmt.Sprintf(regexFormat, suffix)
}
