package kurtosis_instruction

import "go.starlark.net/starlark"

const (
	ArtifactUUIDSuffix = "artifact_uuid"
)

func GetPositionFromThread(thread *starlark.Thread) InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return *NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col)
}

func GetFileNameFromThread(thread *starlark.Thread) string {
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return callFrame.Pos.Filename()
}
