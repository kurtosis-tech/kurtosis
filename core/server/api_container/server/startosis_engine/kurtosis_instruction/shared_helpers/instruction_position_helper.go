package shared_helpers

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"go.starlark.net/starlark"
)

const callerPosition = 1

// GetCallerPositionFromThread gets you the position (line, col, filename) from where this function is called
// We pick the first position on the stack based on this https://github.com/google/starlark-go/blob/eaacdf22efa54ae03ea2ec60e248be80d0cadda0/starlark/eval.go#L136
// As far as I understand, when we call this function from any of the `GenerateXXXBuiltIn` the following occurs
// 1. the 0th position is the builtin, the pos/col on it are 0,0
// 2. the 1st position is the function itself, line is the line of the function, col is the position of the opening parenthesis
// 3. the 2nd position is whatever calls the function, so if its nested in another function its that
// I reckon the stack is built on top of a queue or something, otherwise I'd expect the last item to contain the calling function too
func GetCallerPositionFromThread(thread *starlark.Thread) *kurtosis_instruction.InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	// As the bottom of the stack is guaranteed to be a built in based on above,
	// The 2nd item is the caller, this should always work when called from a GenerateXXXBuiltIn context
	// We panic to eject early in case a bug occurs.
	if thread.CallStackDepth() < 2 {
		panic("Call stack needs to contain at least 2 items for us to get the callers position. This is a Kurtosis Bug.")
	}
	callFrame := thread.CallStack().At(callerPosition)
	return kurtosis_instruction.NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col, callFrame.Pos.Filename())
}
