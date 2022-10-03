package startosis_engine

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

const (
	starlarkGoThreadName                 = "Startosis interpreter thread"
	starlarkFilenamePlaceholderAsNotUsed = "FILENAME_NOT_USED"

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Startosis script. Listing each of them below."
)

type StartosisInterpreter struct {
	serviceNetwork *service_network.ServiceNetwork
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork *service_network.ServiceNetwork) *StartosisInterpreter {
	return &StartosisInterpreter{
		serviceNetwork: serviceNetwork,
	}
}

// Interpret interprets the Startosis script and produce different outputs:
//   - The serialized output of the interpretation (what the Startosis script printed)
//   - A potential interpretation error that the writer of the script should be aware of (syntax error in the Startosis
//     code, inconsistent). Can be nil if the script was successfully interpreted
//   - The list of Kurtosis instructions that was generated based on the interpretation of the script. It can be empty
//     if the interpretation of the script failed
//   - An error if something unexpected happens (crash independent of the Startosis script). This should be as rare as
//     possible
func (interpreter *StartosisInterpreter) Interpret(ctx context.Context, serializedScript string) (SerializedInterpretationOutput, *startosis_errors.InterpretationError, []kurtosis_instruction.KurtosisInstruction) {
	var scriptOutputBuffer bytes.Buffer
	var instructionQueue []kurtosis_instruction.KurtosisInstruction
	thread, builtins := interpreter.buildBindings(&scriptOutputBuffer, &instructionQueue)

	_, err := starlark.ExecFile(thread, starlarkFilenamePlaceholderAsNotUsed, serializedScript, builtins)
	if err != nil {
		return "", generateInterpretationError(err), nil
	}
	logrus.Debugf("Successfully interpreted startosis script into instruction queue: \n%s", instructionQueue)
	return SerializedInterpretationOutput(scriptOutputBuffer.String()), nil, instructionQueue
}

func (interpreter *StartosisInterpreter) buildBindings(scriptOutputBuffer *bytes.Buffer, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) (*starlark.Thread, starlark.StringDict) {
	thread := &starlark.Thread{
		Name: starlarkGoThreadName,
		Load: func(_ *starlark.Thread, fileToLoad string) (starlark.StringDict, error) {
			// TODO(gb): remove when we implement the load feature
			//  Also note we could return the position here by analysing the callstack in the thread object, but not worth it as this will go away soon
			return nil, startosis_errors.NewInterpretationErrorWithCustomMsg("Loading external Startosis scripts is not supported yet", nil)
		},
		Print: func(_ *starlark.Thread, msg string) {
			// From the Starlark spec, a print statement in Starlark is automatically followed by a newline
			scriptOutputBuffer.WriteString(msg + "\n")
		},
	}

	builtins := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		add_service.AddServiceBuiltinName: starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
	}

	return thread, builtins
}

func generateInterpretationError(err error) *startosis_errors.InterpretationError {
	switch err.(type) {
	case resolve.Error:
		slError := err.(resolve.Error)
		return startosis_errors.NewInterpretationErrorFromStacktrace(
			[]startosis_errors.CallFrame{
				*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)),
			},
		)
	case syntax.Error:
		slError := err.(syntax.Error)
		return startosis_errors.NewInterpretationErrorFromStacktrace(
			[]startosis_errors.CallFrame{
				*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)),
			},
		)
	case resolve.ErrorList:
		errorsList := err.(resolve.ErrorList)
		// TODO(gb): a bit hacky but it's an acceptable way to wrap multiple errors into a single Interpretation
		//  it's probably not worth adding another level of complexity here to handle InterpretationErrorList
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, slError := range errorsList {
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(
			multipleInterpretationErrorMsg,
			stacktrace,
		)
	case *starlark.EvalError:
		slError := err.(*starlark.EvalError)
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, callStack := range slError.CallStack {
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(callStack.Name, startosis_errors.NewScriptPosition(callStack.Pos.Line, callStack.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(
			fmt.Sprintf("Evaluation error: %s", slError.Unwrap().Error()),
			stacktrace,
		)
	}
	return startosis_errors.NewInterpretationErrorWithCustomMsg(fmt.Sprintf("UnknownError: %s\n", err.Error()), nil)
}
