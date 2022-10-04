package startosis_errors

import "fmt"

// ScriptPosition represents the position of a call in a script
type ScriptPosition struct {
	line int32

	col int32
}

func NewScriptPosition(line int32, col int32) *ScriptPosition {
	return &ScriptPosition{
		line: line,
		col:  col,
	}
}

func (pos *ScriptPosition) String() string {
	return fmt.Sprintf("[%d:%d]", pos.line, pos.col)
}
