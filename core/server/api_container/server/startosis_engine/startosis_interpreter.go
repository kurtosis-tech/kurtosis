package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
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
func (interpreter *StartosisInterpreter) Interpret(
	_ context.Context,
	packageId string,
	mainFunctionName string,
	serializedStarlark string,
	serializedJsonParams string,
) (string, *instructions_plan.InstructionsPlan, *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError) {
	interpreter.mutex.Lock()
	defer interpreter.mutex.Unlock()
	newInstructionsPlan := instructions_plan.NewInstructionsPlan()
	logrus.Debugf("Interpreting package '%v' with contents '%v' and params '%v'", packageId, serializedStarlark, serializedJsonParams)
	globalVariables, interpretationErr := interpreter.interpretInternal(packageId, serializedStarlark, newInstructionsPlan)
	if interpretationErr != nil {
		return startosis_constants.NoOutputObject, nil, interpretationErr.ToAPIType()
	}

	logrus.Debugf("Successfully interpreted Starlark code into %d instructions", newInstructionsPlan.Size())

	var isUsingDefaultMainFunction bool
	// if the user sends "" or "run" we isUsingDefaultMainFunction to true
	if mainFunctionName == "" || mainFunctionName == runFunctionName {
		mainFunctionName = runFunctionName
		isUsingDefaultMainFunction = true
	}

	if !globalVariables.Has(mainFunctionName) {
		return "", nil, missingMainFunctionError(packageId, mainFunctionName)
	}

	mainFunction, ok := globalVariables[mainFunctionName].(*starlark.Function)
	// if there is an element with the `mainFunctionName` but it isn't a function we have to error as well
	if !ok {
		return "", nil, missingMainFunctionError(packageId, mainFunctionName)
	}

	runFunctionExecutionThread := newStarlarkThread(starlarkGoThreadName)

	if isUsingDefaultMainFunction && mainFunction.NumParams() > maximumParamsAllowedForRunFunction {
		return "", nil, startosis_errors.NewInterpretationError("The 'run' entrypoint function can have at most '%v' argument got '%v'", maximumParamsAllowedForRunFunction, mainFunction.NumParams()).ToAPIType()
	}

	var argsTuple starlark.Tuple
	var kwArgs []starlark.Tuple

	mainFuncParamsNum := mainFunction.NumParams()

	if mainFuncParamsNum >= minimumParamsRequiredForPlan {
		// the plan object will always be injected if the first argument name is 'plan'
		firstParamName, _ := mainFunction.Param(planParamIndex)
		if firstParamName == planParamName {
			kurtosisPlanInstructions := KurtosisPlanInstructions(interpreter.serviceNetwork, interpreter.recipeExecutor, interpreter.moduleContentProvider)
			planModule := plan_module.PlanModule(newInstructionsPlan, kurtosisPlanInstructions)
			argsTuple = append(argsTuple, planModule)
		}

		if firstParamName != planParamName && isUsingDefaultMainFunction {
			return "", nil, startosis_errors.NewInterpretationError(unexpectedArgNameError, planParamIndex, planParamName, firstParamName).ToAPIType()
		}
	}

	if (isUsingDefaultMainFunction && mainFuncParamsNum == paramsRequiredForArgs) ||
		(!isUsingDefaultMainFunction && mainFuncParamsNum > 0) {
		if isUsingDefaultMainFunction {
			if paramName, _ := mainFunction.Param(argsParamIndex); paramName != argsParamName {
				return "", nil, startosis_errors.NewInterpretationError(unexpectedArgNameError, argsParamIndex, argsParamName, paramName).ToAPIType()
			}
		}
		// run function has an argument so we parse input args
		inputArgs, interpretationError := interpreter.parseInputArgs(runFunctionExecutionThread, serializedJsonParams)
		if interpretationError != nil {
			return "", nil, interpretationError.ToAPIType()
		}
		if isUsingDefaultMainFunction {
			argsTuple = append(argsTuple, inputArgs)
			kwArgs = noKwargs
		} else {
			argsDict, ok := inputArgs.(*starlark.Dict)
			if !ok {
				return "", nil, startosis_errors.NewInterpretationError("An error occurred casting input args '%s' to Starlark Dict", inputArgs).ToAPIType()
			}
			kwArgs = append(kwArgs, argsDict.Items()...)
		}
	}

	outputObject, err := starlark.Call(runFunctionExecutionThread, mainFunction, argsTuple, kwArgs)
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
		return serializedOutputObject, newInstructionsPlan, nil
	}
	return startosis_constants.NoOutputObject, newInstructionsPlan, nil
}

func (interpreter *StartosisInterpreter) interpretInternal(packageId string, serializedStarlark string, instructionPlan *instructions_plan.InstructionsPlan) (starlark.StringDict, *startosis_errors.InterpretationError) {
	// We spin up a new thread for every call to interpreterInternal such that the stacktrace provided by the Starlark
	// Go interpreter is relative to each individual thread, and we don't keep accumulating stacktrace entries from the
	// previous calls inside the same thread
	thread := newStarlarkThread(packageId)
	predeclared, interpretationErr := interpreter.buildBindings(instructionPlan)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	globalVariables, err := starlark.ExecFile(thread, packageId, serializedStarlark, *predeclared)
	if err != nil {
		return nil, generateInterpretationError(err)
	}

	return globalVariables, nil
}

func (interpreter *StartosisInterpreter) buildBindings(instructionPlan *instructions_plan.InstructionsPlan) (*starlark.StringDict, *startosis_errors.InterpretationError) {
	recursiveInterpretForModuleLoading := func(moduleId string, serializedStartosis string) (starlark.StringDict, *startosis_errors.InterpretationError) {
		result, err := interpreter.interpretInternal(moduleId, serializedStartosis, instructionPlan)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	kurtosisModule, interpretationErr := builtins.KurtosisModule()
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	predeclared := starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		time.Module.Name:                  time.Module,

		// Kurtosis pre-built module containing Kurtosis constant types
		builtins.KurtosisModuleName: kurtosisModule,
	}

	// Add all Kurtosis helpers
	for _, kurtosisHelper := range KurtosisHelpers(recursiveInterpretForModuleLoading, interpreter.moduleContentProvider, interpreter.moduleGlobalsCache) {
		predeclared[kurtosisHelper.Name()] = kurtosisHelper
	}

	// Add all Kurtosis types
	for _, kurtosisTypeConstructors := range KurtosisTypeConstructors() {
		predeclared[kurtosisTypeConstructors.Name()] = kurtosisTypeConstructors
	}
	return &predeclared, nil
}

// This method handles the different cases a Startosis module can be executed.
// - If input args are empty it uses empty JSON ({}) as the input args
// - If input args aren't empty it tries to deserialize them
func (interpreter *StartosisInterpreter) parseInputArgs(thread *starlark.Thread, serializedJsonArgs string) (starlark.Value, *startosis_errors.InterpretationError) {
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

func missingMainFunctionError(packageId string, mainFunctionName string) *kurtosis_core_rpc_api_bindings.StarlarkInterpretationError {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript {
		return startosis_errors.NewInterpretationError(
			"No '%s' function found in the script; a '%s' entrypoint function with the signature `%s(plan, args)` or `%s()` is required in the Kurtosis script",
			mainFunctionName,
			mainFunctionName,
			mainFunctionName,
			mainFunctionName,
		).ToAPIType()
	}

	return startosis_errors.NewInterpretationError(
		"No '%s' function found in the main file of package '%s'; a '%s' entrypoint function with the signature `%s(plan, args)` or `%s()` is required in the main file of the Kurtosis package",
		mainFunctionName,
		packageId,
		mainFunctionName,
		mainFunctionName,
		mainFunctionName,
	).ToAPIType()
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
