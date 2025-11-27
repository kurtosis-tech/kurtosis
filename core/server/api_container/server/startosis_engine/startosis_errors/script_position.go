package startosis_errors

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
)

const (
	builtInName = "<builtin>"
)

var skipFilenamesValueSet = map[string]bool{
	startosis_constants.PackageIdPlaceholderForStandaloneScript: true,
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
	if _, found := skipFilenamesValueSet[pos.filename]; found {
		return fmt.Sprintf("[%d:%d]", pos.line, pos.col)
	}
	return fmt.Sprintf("[%s:%d:%d]", pos.filename, pos.line, pos.col)
}
