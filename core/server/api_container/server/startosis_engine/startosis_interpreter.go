package startosis_engine

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
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
	serviceNetwork     *service_network.ServiceNetwork
	moduleCache        *startosis_modules.ModuleCache
	moduleManager      startosis_modules.ModuleManager
	scriptOutputBuffer bytes.Buffer
	instructionsQueue  []kurtosis_instruction.KurtosisInstruction
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork *service_network.ServiceNetwork, moduleManager startosis_modules.ModuleManager) *StartosisInterpreter {
	return &StartosisInterpreter{
		serviceNetwork: serviceNetwork,
		moduleManager:  moduleManager,
		moduleCache:    startosis_modules.NewModuleCache(),
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
	thread, builtins := interpreter.buildBindings(starlarkGoThreadName)

	_, err := starlark.ExecFile(thread, starlarkFilenamePlaceholderAsNotUsed, serializedScript, builtins)
	if err != nil {
		return "", generateInterpretationError(err), nil
	}
	logrus.Debugf("Successfully interpreted Startosis script into instruction queue: \n%s", interpreter.instructionsQueue)
	return SerializedInterpretationOutput(interpreter.scriptOutputBuffer.String()), nil, interpreter.instructionsQueue
}

func (interpreter *StartosisInterpreter) buildBindings(threadName string) (*starlark.Thread, starlark.StringDict) {
	thread := &starlark.Thread{
		Name: threadName,
		Load: interpreter.Load,
		Print: func(_ *starlark.Thread, msg string) {
			// From the Starlark spec, a print statement in Starlark is automatically followed by a newline
			interpreter.scriptOutputBuffer.WriteString(msg + "\n")
		},
	}

	builtins := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		add_service.AddServiceBuiltinName: starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(&interpreter.instructionsQueue, interpreter.serviceNetwork)),
	}

	return thread, builtins
}

func (interpreter *StartosisInterpreter) Load(_ *starlark.Thread, moduleID string) (starlark.StringDict, error) {
	if interpreter.moduleCache.IsLoadInProgress(moduleID) {
		return nil, startosis_errors.NewInterpretationError("There is a cycle in the load graph")
	}

	entry, found := interpreter.moduleCache.Get(moduleID)
	if found {
		return entry.GetGlobalVariables(), entry.GetError()
	}

	interpreter.moduleCache.SetLoadInProgress(moduleID)

	// Load it.
	contents, err := interpreter.moduleManager.GetModule(moduleID)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while loading the module '%v'", moduleID))
	}

	thread, bindings := interpreter.buildBindings(fmt.Sprintf("%v:%v", starlarkGoThreadName, moduleID))
	globalVariables, err := starlark.ExecFile(thread, moduleID, contents, bindings)

	// Update the cache.
	entry = startosis_modules.NewModuleCacheEntry(globalVariables, err)
	interpreter.moduleCache.Add(moduleID, entry)

	return entry.GetGlobalVariables(), entry.GetError()
}

func generateInterpretationError(err error) *startosis_errors.InterpretationError {
	switch slError := err.(type) {
	case resolve.Error:
		stacktrace := []startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)),
		}
		return startosis_errors.NewInterpretationErrorFromStacktrace(stacktrace)
	case syntax.Error:
		stacktrace := []startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)),
		}
		return startosis_errors.NewInterpretationErrorFromStacktrace(stacktrace)
	case resolve.ErrorList:
		// TODO(gb): a bit hacky but it's an acceptable way to wrap multiple errors into a single Interpretation
		//  it's probably not worth adding another level of complexity here to handle InterpretationErrorList
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, slError := range slError {
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Line, slError.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(multipleInterpretationErrorMsg, stacktrace)
	case *starlark.EvalError:
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, callStack := range slError.CallStack {
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(callStack.Name, startosis_errors.NewScriptPosition(callStack.Pos.Line, callStack.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(
			fmt.Sprintf("Evaluation error: %s", slError.Unwrap().Error()),
			stacktrace,
		)
	}
	return startosis_errors.NewInterpretationError(fmt.Sprintf("UnknownError: %s\n", err.Error()))
}
