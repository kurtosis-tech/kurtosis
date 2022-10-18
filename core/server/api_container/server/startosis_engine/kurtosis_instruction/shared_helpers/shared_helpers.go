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
)

func GetPositionFromThread(thread *starlark.Thread) *kurtosis_instruction.InstructionPosition {
	// TODO(gb): can do better by returning the entire callstack positions, but it's a good start
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return kurtosis_instruction.NewInstructionPosition(callFrame.Pos.Line, callFrame.Pos.Col)
}

func GetFileNameFromThread(thread *starlark.Thread) string {
	if len(thread.CallStack()) == 0 {
		panic("empty call stack is unexpected, this should not happen")
	}
	// position of current instruction is  store at the bottom of the call stack
	callFrame := thread.CallStack().At(len(thread.CallStack()) - 1)
	return callFrame.Pos.Filename()
}

func ReplaceMagicStringWithValue(suffixToReplace string, originalString string, serviceIdForLogging string, environment *startosis_executor.ExecutionEnvironment) (string, error) {
	compiledArtifactUuidRegex := regexp.MustCompile(kurtosis_instruction.GetRegularExpressionForInstruction(suffixToReplace))
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
