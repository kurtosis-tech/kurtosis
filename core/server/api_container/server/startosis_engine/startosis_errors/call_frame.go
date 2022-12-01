package startosis_errors

import (
	"fmt"
)

type CallFrame struct {
	name string

	position *ScriptPosition
}

func NewCallFrame(name string, position *ScriptPosition) *CallFrame {
	return &CallFrame{
		name:     name,
		position: position,
	}
}

func (callFrame *CallFrame) String() string {
	return fmt.Sprintf("%s: %s", callFrame.position.String(), callFrame.name)
}
