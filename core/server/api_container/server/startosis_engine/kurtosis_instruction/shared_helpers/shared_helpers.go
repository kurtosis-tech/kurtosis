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

func GetPositionFromThread(thread *starlark.Thread) *kurtosis_instruction.InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	if thread.CallStackDepth() < 2 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// bottom of the stack is <built_in>
	// position 1 is the position of the caller
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
