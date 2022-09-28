package startosis_engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/sirupsen/logrus"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

type StartosisInterpreter struct {
	serviceNetwork *service_network.ServiceNetwork
}

type SerializedInterpretationOutput struct {
	output string
}

type InterpretationError struct {
	error string
}

func NewStartosisInterpreter(serviceNetwork *service_network.ServiceNetwork) *StartosisInterpreter {
	return &StartosisInterpreter{
		serviceNetwork: serviceNetwork,
	}
}

func (serializedInterpretationOutput *SerializedInterpretationOutput) Get() string {
	return serializedInterpretationOutput.output
}

func (interpretationError *InterpretationError) Get() string {
	return interpretationError.error
}

// Interpret interprets the Startosis script and produce different outputs:
//   - The serialized output of the interpretation (what the Startosis script printed)
//   - A potential interpretation error that the writer of the script should be aware of (syntax error in the Startosis
//     code, inconsistent). Can be nil if the script was successfully interpreted
//   - The list of Kurtosis instructions that was generated based on the interpretation of the script. It can be empty
//     if the interpretation of the script failed
//   - An error if something unexpected happens (crash independent of the Startosis script). This should be as rare as
//     possible
func (interpreter *StartosisInterpreter) Interpret(ctx context.Context, serializedScript string) (*SerializedInterpretationOutput, *InterpretationError, []kurtosis_instruction.KurtosisInstruction, error) {
	var scriptOutputBuffer bytes.Buffer
	var instructionQueue []kurtosis_instruction.KurtosisInstruction
	thread, builtins := interpreter.buildBindings(&scriptOutputBuffer, &instructionQueue)

	_, err := starlark.ExecFile(thread, "FILENAME_NOT_USED", serializedScript, builtins)
	if err != nil {
		interpretationError := generateErrorString(err)
		return nil, &InterpretationError{interpretationError}, nil, nil
	}
	logrus.Debugf("Successfully interpreted startosis script into instruction queue: \n%s", instructionQueue)
	return &SerializedInterpretationOutput{scriptOutputBuffer.String()}, nil, instructionQueue, nil
}

func (interpreter *StartosisInterpreter) buildBindings(scriptOutputBuffer *bytes.Buffer, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) (*starlark.Thread, starlark.StringDict) {
	thread := &starlark.Thread{
		Name: "Startosis interpreter thread",
		Load: func(_ *starlark.Thread, fileToLoad string) (starlark.StringDict, error) {
			return nil, errors.New("Loading external Startosis scripts is not supported yet")
		},
		Print: func(_ *starlark.Thread, msg string) {
			// From the starlark spec, a print statement in starlark is automatically followed by a newline
			scriptOutputBuffer.WriteString(msg + "\n")
		},
	}

	builtins := starlark.StringDict{
		starlarkstruct.Default.GoString():          starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		kurtosis_instruction.AddServiceBuiltinName: starlark.NewBuiltin(kurtosis_instruction.AddServiceBuiltinName, kurtosis_instruction.AddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
	}

	return thread, builtins
}

func generateErrorString(err error) string {
	errorStringBuffer := &bytes.Buffer{}
	errorStringBuffer.WriteString("/!\\ Errors interpreting Startosis script /!\\\n")
	switch err.(type) {
	case resolve.Error:
		slError := err.(resolve.Error)
		writeStarlarkErrorToBuffer(errorStringBuffer, slError.Pos.Line, slError.Pos.Col, slError.Msg)
		break
	case syntax.Error:
		slError := err.(syntax.Error)
		writeStarlarkErrorToBuffer(errorStringBuffer, slError.Pos.Line, slError.Pos.Col, slError.Msg)
		break
	case resolve.ErrorList:
		errorsList := err.(resolve.ErrorList)
		for _, slError := range errorsList {
			writeStarlarkErrorToBuffer(errorStringBuffer, slError.Pos.Line, slError.Pos.Col, slError.Msg)
		}
		break
	case *starlark.EvalError:
		slError := err.(*starlark.EvalError)
		errorStringBuffer.WriteString(fmt.Sprintf("\tEvaluationError: %s\n", slError.Unwrap().Error()))
		for _, callStack := range slError.CallStack {
			errorStringBuffer.WriteString(fmt.Sprintf("\t\tat [%d:%d]: %s\n", callStack.Pos.Line, callStack.Pos.Col, callStack.Name))
		}
		break
	default:
		errorStringBuffer.WriteString(fmt.Sprintf("\tUnkownError: %s\n", err.Error()))
		break
	}
	return errorStringBuffer.String()
}

func writeStarlarkErrorToBuffer(scriptOutputBuffer *bytes.Buffer, line int32, col int32, msg string) {
	scriptOutputBuffer.WriteString(fmt.Sprintf("\t[%d:%d]: %s\n", line, col, msg))
}
