package shared_helpers

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"regexp"
	"strings"
)

const (
	ArtifactUUIDSuffix = "artifact_uuid"

	unlimitedMatches = -1
	singleMatch      = 1

	callerPosition = 1
)

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

// ReplaceArtifactUuidMagicStringWithValue This function gets used to replace artifact uuid magic strings generated during interpretation time with actual values during execution time
// TODO extend this in the future to be generic, instead of environment one could pass in a func(string) -> string maybe
func ReplaceArtifactUuidMagicStringWithValue(originalString string, serviceIdForLogging string, environment *startosis_executor.ExecutionEnvironment) (string, error) {
	compiledArtifactUuidRegex := regexp.MustCompile(kurtosis_instruction.GetRegularExpressionForInstruction(ArtifactUUIDSuffix))
	matches := compiledArtifactUuidRegex.FindAllString(originalString, unlimitedMatches)
	replacedString := originalString
	for _, match := range matches {
		artifactUuid, found := environment.GetArtifactUuid(match)
		if !found {
			return "", stacktrace.NewError("Couldn't find '%v' in the execution environment which is required by service '%v'", originalString, serviceIdForLogging)
		}
		replacedString = strings.Replace(replacedString, match, artifactUuid, singleMatch)
	}
	return replacedString, nil
}
