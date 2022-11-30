package startosis_errors

import "fmt"

const(
	skipTopLevelCallFrameName = "<toplevel>"
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
	if callFrame.name == skipTopLevelCallFrameName{
		return callFrame.position.String()
	}
	return fmt.Sprintf("%s: %s", callFrame.position.String(), callFrame.name)
}
