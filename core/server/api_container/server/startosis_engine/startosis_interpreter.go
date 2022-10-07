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
	"sync"
)

const (
	starlarkGoThreadName                 = "Startosis interpreter thread"
	starlarkFilenamePlaceholderAsNotUsed = "FILENAME_NOT_USED"

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Startosis script. Listing each of them below."
)

type StartosisInterpreter struct {
	mutex              *sync.Mutex
	serviceNetwork     service_network.ServiceNetwork
	moduleGlobalsCache map[string]*startosis_modules.ModuleCacheEntry
	// TODO AUTH there will be a leak here in case people with different repo visibility access a module
	moduleManager startosis_modules.ModuleContentProvider
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork service_network.ServiceNetwork, moduleManager startosis_modules.ModuleContentProvider) *StartosisInterpreter {
	return &StartosisInterpreter{
		serviceNetwork:     serviceNetwork,
		moduleManager:      moduleManager,
		moduleGlobalsCache: make(map[string]*startosis_modules.ModuleCacheEntry),
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
	var instructionsQueue []kurtosis_instruction.KurtosisInstruction
	thread, builtins := interpreter.buildBindings(starlarkGoThreadName, &instructionsQueue, &scriptOutputBuffer)

	_, err := starlark.ExecFile(thread, starlarkFilenamePlaceholderAsNotUsed, serializedScript, builtins)
	if err != nil {
		return "", generateInterpretationError(err), nil
	}
	logrus.Debugf("Successfully interpreted Startosis script into instruction queue: \n%s", instructionsQueue)
	return SerializedInterpretationOutput(scriptOutputBuffer.String()), nil, instructionsQueue
}

func (interpreter *StartosisInterpreter) buildBindings(threadName string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) (*starlark.Thread, starlark.StringDict) {
	thread := &starlark.Thread{
		Name:  threadName,
		Load:  interpreter.makeLoad(instructionsQueue, scriptOutputBuffer),
		Print: makePrint(scriptOutputBuffer),
	}

	builtins := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		add_service.AddServiceBuiltinName: starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
	}

	return thread, builtins
}

func (interpreter *StartosisInterpreter) makeLoad(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) func(_ *starlark.Thread, moduleID string) (starlark.StringDict, error) {
	// A nil entry to indicate that a load is in progress
	return func(_ *starlark.Thread, moduleID string) (starlark.StringDict, error) {
		var loadInProgress *startosis_modules.ModuleCacheEntry
		entry, found := interpreter.moduleGlobalsCache[moduleID]
		if found && entry == loadInProgress {
			return nil, startosis_errors.NewInterpretationError("There is a cycle in the load graph")
		} else if found {
			return entry.GetGlobalVariables(), entry.GetError()
		}

		interpreter.moduleGlobalsCache[moduleID] = loadInProgress
		shouldUnsetLoadInProgress := true
		defer func() {
			if shouldUnsetLoadInProgress {
				delete(interpreter.moduleGlobalsCache, moduleID)
			}
		}()

		// Load it.
		contents, err := interpreter.moduleManager.GetModuleContents(moduleID)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while loading the module '%v'", moduleID))
		}

		thread, bindings := interpreter.buildBindings(fmt.Sprintf("%v:%v", starlarkGoThreadName, moduleID), instructionsQueue, scriptOutputBuffer)
		globalVariables, err := starlark.ExecFile(thread, moduleID, contents, bindings)

		// Update the cache.
		entry = startosis_modules.NewModuleCacheEntry(globalVariables, err)
		interpreter.moduleGlobalsCache[moduleID] = entry

		shouldUnsetLoadInProgress = false
		return entry.GetGlobalVariables(), entry.GetError()
	}
}

func makePrint(scriptOutputBuffer *bytes.Buffer) func(*starlark.Thread, string) {
	return func(_ *starlark.Thread, msg string) {
		// From the Starlark spec, a print statement in Starlark is automatically followed by a newline
		scriptOutputBuffer.WriteString(msg + "\n")
	}
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
