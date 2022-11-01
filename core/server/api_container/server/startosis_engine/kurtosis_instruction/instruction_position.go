package kurtosis_instruction

import (
	"fmt"
)

const (
	// these two are here so that they align
	// the placeholder is the magic string that gets assigned during interpretation
	// the regex gets used during execution to find magic strings to replace with
	// actual values
	magicStringFormat = "{{kurtosis:%v-%v:%v.%v}}"
	// this allows for alphanumeric, casing and underscores, dashes and dots
	// real world example in the test
	regexFormat = "{{kurtosis:[a-zA-Z0-9_./-]+-[0-9]+:[0-9]+.%v}}"
)

type InstructionPosition struct {
	line int32
	col  int32
	name string
}

func NewInstructionPosition(line int32, col int32, name string) *InstructionPosition {
	return &InstructionPosition{
		line: line,
		col:  col,
		name: name,
	}
}

// MagicString the magic string allows us to identify an instruction that doesn't
// have any other obvious identifiers, we take the line & column number of the instruction
// and add a suffix to it to track what is returned by that instruction
// If an instruction returns multiple things, use different suffixes for each object
// This string gets assigned to the object during interpretation time and replaced during
// execution time
func (position *InstructionPosition) MagicString(suffix string) string {
	return fmt.Sprintf(magicStringFormat, position.name, position.line, position.col, suffix)
}

// GetRegularExpressionForInstruction this function allows you to get a regular expression
// that matches an instruction derived magic string, just pass in the same suffix
// that you used to generate the magic string with
func GetRegularExpressionForInstruction(suffix string) string {
	return fmt.Sprintf(regexFormat, suffix)
}
