package startosis_errors

import "fmt"

const (
	builtInName = "<builtin>"
)

// ScriptPosition represents the position of a call in a script
type ScriptPosition struct {
	filename string
	line     int32
	col      int32
}

func NewScriptPosition(filename string, line int32, col int32) *ScriptPosition {
	return &ScriptPosition{
		line:     line,
		col:      col,
		filename: filename,
	}
}

func (pos *ScriptPosition) String() string {
	if pos.filename == builtInName {
		return fmt.Sprintf("[%d:%d]", pos.line, pos.col)
	}
	return fmt.Sprintf("[%s:%d:%d]", pos.filename, pos.line, pos.col)
}
