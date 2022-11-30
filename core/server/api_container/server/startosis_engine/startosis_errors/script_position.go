package startosis_errors

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_const"
)

const (
	builtInName = "<builtin>"
)

var replaceFilenameValuesSet = map[string]bool{
	startosis_const.PackageIdPlaceholderForStandaloneScript: true,
	builtInName: true,
}

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
	if _, found := replaceFilenameValuesSet[pos.filename]; found{
		return fmt.Sprintf("[%d:%d]", pos.line, pos.col)
	}
	return fmt.Sprintf("[%s:%d:%d]", pos.filename, pos.line, pos.col)
}
