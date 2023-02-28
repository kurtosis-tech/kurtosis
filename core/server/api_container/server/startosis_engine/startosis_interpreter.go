package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/plan_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/package_io"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/sirupsen/logrus"
	"go.starlark.net/lib/time"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	"strings"
	"sync"
)

const (
	starlarkGoThreadName = "Startosis interpreter thread"

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Starlark script. Listing each of them below."
	evaluationErrorPrefix          = "Evaluation error: "

	skipImportInstructionInStacktraceValue = "import_module"

	runFunctionName = "run"

	paramsRequiredForArgs              = 2
	minimumParamsRequiredForPlan       = 1
	maximumParamsAllowedForRunFunction = 2

	planParamIndex         = 0
	planParamName          = "plan"
	argsParamIndex         = 1
	argsParamName          = "args"
	unexpectedArgNameError = "Expected argument at index '%v' of run function to be called '%v' got '%v' "
)

var (
	noKwargs []starlark.Tuple
)

type StartosisInterpreter struct {
	// This is mutex protected as interpreting two different scripts in parallel could potentially cause
	// problems with the moduleGlobalsCache & moduleContentProvider. Fixing this is quite complicated, which we decided not to do.
	mutex              *sync.Mutex
	serviceNetwork     service_network.ServiceNetwork
	recipeExecutor     *runtime_value_store.RuntimeValueStore
	moduleGlobalsCache map[string]*startosis_packages.ModuleCacheEntry
	// TODO AUTH there will be a leak here in case people with different repo visibility access a module
	moduleContentProvider startosis_packages.PackageContentProvider
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork service_network.ServiceNetwork, moduleContentProvider startosis_packages.PackageContentProvider, runtimeValueStore *runtime_value_store.RuntimeValueStore) *StartosisInterpreter {
	return &StartosisInterpreter{
		mutex:                 &sync.Mutex{},
		serviceNetwork:        serviceNetwork,
		recipeExecutor:        runtimeValueStore,
		moduleGlobalsCache:    make(map[string]*startosis_packages.ModuleCacheEntry),
		moduleContentProvider: moduleContentProvider,
	}
}

// Interpret interprets the Starlark script and produce different outputs:
//   - A potential interpretation error that the writer of the script should be aware of (syntax error in the Startosis
//     code, inconsistent). Can be nil if the script was successfully interpreted
//   - The list of Kurtosis instructions that was generated based on the interpretation of the script. It can be empty
//     if the interpretation of the script failed
func (interpreter *StartosisInterpreter) Interpret(_ context.Context, packageId string, serializedStarlark string, serializedJsonParams string) (string, []kurtosis_instruction.KurtosisInstruction, *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError) {
	interpreter.mutex.Lock()
	defer interpreter.mutex.Unlock()
	var instructionsQueue []kurtosis_instruction.KurtosisInstruction

	globalVariables, interpretationErr := interpreter.interpretInternal(packageId, serializedStarlark, &instructionsQueue)
	if interpretationErr != nil {
		return startosis_constants.NoOutputObject, nil, interpretationErr.ToAPIType()
	}

	logrus.Debugf("Successfully interpreted Starlark code into instruction queue: \n%s", instructionsQueue)

	if !globalVariables.Has(runFunctionName) {
		return missingRunFunctionReturnValue(packageId)
	}

	runFunction, ok := globalVariables[runFunctionName].(*starlark.Function)
	// if there is a `run` but it isn't a function we have to error as well
	if !ok {
		return missingRunFunctionReturnValue(packageId)
	}

	runFunctionExecutionThread := newStarlarkThread(starlarkGoThreadName)

	if runFunction.NumParams() > maximumParamsAllowedForRunFunction {
		return "", nil, startosis_errors.NewInterpretationError("The 'run' entrypoint function can have at most '%v' argument got '%v'", maximumParamsAllowedForRunFunction, runFunction.NumParams()).ToAPIType()
	}

	var argsTuple starlark.Tuple

	if runFunction.NumParams() >= minimumParamsRequiredForPlan {
		if paramName, _ := runFunction.Param(planParamIndex); paramName != planParamName {
			return "", nil, startosis_errors.NewInterpretationError(unexpectedArgNameError, planParamIndex, planParamName, paramName).ToAPIType()
		}
		oldKurtosisPlanInstructions := OldKurtosisPlanInstructions(&instructionsQueue, interpreter.moduleContentProvider, interpreter.serviceNetwork, interpreter.recipeExecutor)
		kurtosisPlanInstructions := KurtosisPlanInstructions(interpreter.serviceNetwork, interpreter.recipeExecutor, interpreter.moduleContentProvider)
		planModule := plan_module.PlanModule(&instructionsQueue, oldKurtosisPlanInstructions, kurtosisPlanInstructions)
		argsTuple = append(argsTuple, planModule)
	}

	if runFunction.NumParams() == paramsRequiredForArgs {
		if paramName, _ := runFunction.Param(argsParamIndex); paramName != argsParamName {
			return "", nil, startosis_errors.NewInterpretationError(unexpectedArgNameError, argsParamIndex, argsParamName, paramName).ToAPIType()
		}
		// run function has an argument so we parse input args
		inputArgs, interpretationError := interpreter.parseInputArgs(runFunctionExecutionThread, serializedJsonParams)
		if interpretationError != nil {
			return "", nil, interpretationError.ToAPIType()
		}
		argsTuple = append(argsTuple, inputArgs)
	}

	outputObject, err := starlark.Call(runFunctionExecutionThread, runFunction, argsTuple, noKwargs)
	if err != nil {
		return "", nil, generateInterpretationError(err).ToAPIType()
	}

	// Serialize and return the output object. It might contain magic strings that should be resolved post-execution
	if outputObject != starlark.None {
		logrus.Debugf("Starlark output object was: '%s'", outputObject)
		serializedOutputObject, interpretationError := package_io.SerializeOutputObject(runFunctionExecutionThread, outputObject)
		if interpretationError != nil {
			return startosis_constants.NoOutputObject, nil, interpretationError.ToAPIType()
		}
		return serializedOutputObject, instructionsQueue, nil
	}
	return startosis_constants.NoOutputObject, instructionsQueue, nil
}

func (interpreter *StartosisInterpreter) interpretInternal(packageId string, serializedStarlark string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) (starlark.StringDict, *startosis_errors.InterpretationError) {
	// We spin up a new thread for every call to interpreterInternal such that the stacktrace provided by the Starlark
	// Go interpreter is relative to each individual thread, and we don't keep accumulating stacktrace entries from the
	// previous calls inside the same thread
	thread := newStarlarkThread(packageId)
	predeclared := interpreter.buildBindings(instructionsQueue)

	globalVariables, err := starlark.ExecFile(thread, packageId, serializedStarlark, *predeclared)
	if err != nil {
		return nil, generateInterpretationError(err)
	}

	return globalVariables, nil
}

func (interpreter *StartosisInterpreter) buildBindings(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) *starlark.StringDict {
	recursiveInterpretForModuleLoading := func(moduleId string, serializedStartosis string) (starlark.StringDict, *startosis_errors.InterpretationError) {
		result, err := interpreter.interpretInternal(moduleId, serializedStartosis, instructionsQueue)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	predeclared := starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		time.Module.Name:                  time.Module,

		// Kurtosis pre-built module containing Kurtosis constant types
		builtins.KurtosisModuleName: builtins.KurtosisModule(),
	}

	// Add all Kurtosis helpers
	for _, kurtosisHelper := range KurtosisHelpers(recursiveInterpretForModuleLoading, interpreter.moduleContentProvider, interpreter.moduleGlobalsCache) {
		predeclared[kurtosisHelper.Name()] = kurtosisHelper
	}

	// Add all Kurtosis types
	for _, kurtosisTypeConstructors := range KurtosisTypeConstructors() {
		predeclared[kurtosisTypeConstructors.Name()] = kurtosisTypeConstructors
	}
	return &predeclared
}

// This method handles the different cases a Startosis module can be executed.
// - If input args are empty it uses empty JSON ({}) as the input args
// - If input args aren't empty it tries to deserialize them
func (interpreter *StartosisInterpreter) parseInputArgs(thread *starlark.Thread, serializedJsonArgs string) (starlark.Value, *startosis_errors.InterpretationError) {
	if serializedJsonArgs == startosis_constants.EmptyInputArgs {
		return starlark.None, nil
	}
	// it is a module, and it has input args -> deserialize the JSON input and add it as a struct to the predeclared
	deserializedArgs, interpretationError := package_io.DeserializeArgs(thread, serializedJsonArgs)
	if interpretationError != nil {
		return nil, interpretationError
	}
	return deserializedArgs, nil
}

func makeLoadFunction() func(_ *starlark.Thread, packageId string) (starlark.StringDict, error) {
	return func(_ *starlark.Thread, _ string) (starlark.StringDict, error) {
		return nil, startosis_errors.NewInterpretationError("'load(\"path/to/file.star\", var_in_file=\"var_in_file\")' statement is not available in Kurtosis. Please use instead `module = import(\"path/to/file.star\")` and then `module.var_in_file`")
	}
}

func makePrintFunction() func(*starlark.Thread, string) {
	return func(_ *starlark.Thread, msg string) {
		// the `print` function must be overriden with the custom print builtin in the predeclared map
		// which just exists to throw a nice interpretation error as this itself can't
		panic(print_builtin.UsePlanFromKurtosisInstructionError)
	}
}

func generateInterpretationError(err error) *startosis_errors.InterpretationError {
	switch slError := err.(type) {
	case resolve.Error:
		stacktrace := []startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Filename(), slError.Pos.Line, slError.Pos.Col)),
		}
		return startosis_errors.NewInterpretationErrorFromStacktrace(stacktrace)
	case syntax.Error:
		stacktrace := []startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame(slError.Msg, startosis_errors.NewScriptPosition(slError.Pos.Filename(), slError.Pos.Line, slError.Pos.Col)),
		}
		return startosis_errors.NewInterpretationErrorFromStacktrace(stacktrace)
	case resolve.ErrorList:
		// TODO(gb): a bit hacky but it's an acceptable way to wrap multiple errors into a single Interpretation
		//  it's probably not worth adding another level of complexity here to handle InterpretationErrorList
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, slErr := range slError {
			if slErr.Msg == skipImportInstructionInStacktraceValue {
				continue
			}
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(slErr.Msg, startosis_errors.NewScriptPosition(slErr.Pos.Filename(), slErr.Pos.Line, slErr.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(stacktrace, multipleInterpretationErrorMsg)
	case *starlark.EvalError:
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, callStack := range slError.CallStack {
			if callStack.Name == skipImportInstructionInStacktraceValue {
				continue
			}
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(callStack.Name, startosis_errors.NewScriptPosition(callStack.Pos.Filename(), callStack.Pos.Line, callStack.Pos.Col)))
		}
		var errorMsg string
		// no need to add the evaluation error prefix if the wrapped error already has it
		if strings.HasPrefix(slError.Unwrap().Error(), evaluationErrorPrefix) {
			errorMsg = slError.Unwrap().Error()
		} else {
			errorMsg = fmt.Sprintf("%s%s", evaluationErrorPrefix, slError.Unwrap().Error())
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(stacktrace, errorMsg)
	case *startosis_errors.InterpretationError:
		// If it's already an interpretation error -> nothing to convert
		return slError
	}
	return startosis_errors.NewInterpretationError("UnknownError: %s\n", err.Error())
}

func missingRunFunctionReturnValue(packageId string) (string, []kurtosis_instruction.KurtosisInstruction, *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError) {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript {
		return "", nil, startosis_errors.NewInterpretationError("No 'run' function found in the script; a 'run' entrypoint function with the signature `run(args)` or `run()` is required in any Kurtosis script").ToAPIType()
	}
	return "", nil, startosis_errors.NewInterpretationError("No 'run' function found in file '%v/main.star'; a 'run' entrypoint function with the signature `run(args)` or `run()` is required in the main.star file of any Kurtosis package", packageId).ToAPIType()
}

func newStarlarkThread(threadName string) *starlark.Thread {
	return &starlark.Thread{
		Name:       threadName,
		Print:      makePrintFunction(),
		Load:       makeLoadFunction(),
		OnMaxSteps: nil,
		Steps:      0,
	}
}
