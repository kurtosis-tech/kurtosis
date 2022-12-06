package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/assert"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/define_fact"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/extract"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/get_value"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
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
	"sync"
)

const (
	starlarkGoThreadName = "Startosis interpreter thread"

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Starlark script. Listing each of them below."

	skipImportInstructionInStacktraceValue = "import_module"
)

type StartosisInterpreter struct {
	// This is mutex protected as interpreting two different scripts in parallel could potentially cause
	// problems with the moduleGlobalsCache & moduleContentProvider. Fixing this is quite complicated, which we decided not to do.
	mutex              *sync.Mutex
	serviceNetwork     service_network.ServiceNetwork
	factsEngine        *facts_engine.FactsEngine
	recipeExecutor     *runtime_value_store.RuntimeValueStore
	moduleGlobalsCache map[string]*startosis_packages.ModuleCacheEntry
	// TODO AUTH there will be a leak here in case people with different repo visibility access a module
	moduleContentProvider startosis_packages.PackageContentProvider
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork service_network.ServiceNetwork, moduleContentProvider startosis_packages.PackageContentProvider) *StartosisInterpreter {
	return &StartosisInterpreter{
		mutex:                 &sync.Mutex{},
		serviceNetwork:        serviceNetwork,
		factsEngine:           nil,
		recipeExecutor:        nil,
		moduleGlobalsCache:    make(map[string]*startosis_packages.ModuleCacheEntry),
		moduleContentProvider: moduleContentProvider,
	}
}

func NewStartosisInterpreterWithFacts(serviceNetwork service_network.ServiceNetwork, factsEngine *facts_engine.FactsEngine, moduleContentProvider startosis_packages.PackageContentProvider, recipeExecutor *runtime_value_store.RuntimeValueStore) *StartosisInterpreter {
	return &StartosisInterpreter{
		mutex:                 &sync.Mutex{},
		serviceNetwork:        serviceNetwork,
		moduleContentProvider: moduleContentProvider,
		recipeExecutor:        recipeExecutor,
		moduleGlobalsCache:    make(map[string]*startosis_packages.ModuleCacheEntry),
		factsEngine:           factsEngine,
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

	thread := &starlark.Thread{
		Name:       starlarkGoThreadName,
		Print:      makePrintFunction(),
		Load:       makeLoadFunction(),
		OnMaxSteps: nil,
	}

	globalVariables, err := interpreter.interpretInternal(thread, packageId, serializedStarlark, serializedJsonParams, &instructionsQueue)
	if err != nil {
		return startosis_constants.NoOutputObject, nil, generateInterpretationError(err).ToAPIType()
	}

	logrus.Debugf("Successfully interpreted Starlark code into instruction queue: \n%s", instructionsQueue)

	// Serialize and return the output object. It might contain magic strings that should be resolved post-execution
	if globalVariables.Has(startosis_constants.MainOutputObjectName) && globalVariables[startosis_constants.MainOutputObjectName] != starlark.None {
		logrus.Debugf("Starlark output object was: '%s'", globalVariables[startosis_constants.MainOutputObjectName])
		serializedOutputObject, interpretationError := package_io.SerializeOutputObject(thread, globalVariables[startosis_constants.MainOutputObjectName])
		if interpretationError != nil {
			return startosis_constants.NoOutputObject, nil, interpretationError.ToAPIType()
		}
		return serializedOutputObject, instructionsQueue, nil
	}
	return startosis_constants.NoOutputObject, instructionsQueue, nil
}

func (interpreter *StartosisInterpreter) interpretInternal(thread *starlark.Thread, packageId string, serializedStarlark string, serializedJsonParams string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) (starlark.StringDict, error) {
	predeclared := interpreter.buildBindings(thread, instructionsQueue)

	if interpretationError := interpreter.addInputArgsToPredeclared(thread, packageId, serializedJsonParams, predeclared); interpretationError != nil {
		return nil, interpretationError
	}

	globalVariables, err := starlark.ExecFile(thread, packageId, serializedStarlark, *predeclared)
	if err != nil {
		return nil, err
	}
	return globalVariables, nil
}

func (interpreter *StartosisInterpreter) buildBindings(thread *starlark.Thread, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction) *starlark.StringDict {
	recursiveInterpretForModuleLoading := func(moduleId string, serializedStartosis string) (starlark.StringDict, error) {
		return interpreter.interpretInternal(thread, moduleId, serializedStartosis, startosis_constants.EmptyInputArgs, instructionsQueue)
	}

	predeclared := &starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		time.Module.Name:                  time.Module,

		// Kurtosis instructions - will push instructions to the queue that will affect the enclave state at execution
		add_service.AddServiceBuiltinName:                starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork, interpreter.factsEngine)),
		assert.AssertBuiltinName:                         starlark.NewBuiltin(assert.AssertBuiltinName, assert.GenerateAssertBuiltin(instructionsQueue, interpreter.recipeExecutor, interpreter.serviceNetwork)),
		exec.ExecBuiltinName:                             starlark.NewBuiltin(exec.ExecBuiltinName, exec.GenerateExecBuiltin(instructionsQueue, interpreter.serviceNetwork, interpreter.recipeExecutor)),
		extract.ExtractBuiltinName:                       starlark.NewBuiltin(extract.ExtractBuiltinName, extract.GenerateExtractInstructionBuiltin(instructionsQueue, interpreter.recipeExecutor, interpreter.serviceNetwork)),
		get_value.GetValueBuiltinName:                    starlark.NewBuiltin(get_value.GetValueBuiltinName, get_value.GenerateGetValueBuiltin(instructionsQueue, interpreter.recipeExecutor, interpreter.serviceNetwork)),
		kurtosis_print.PrintBuiltinName:                  starlark.NewBuiltin(kurtosis_print.PrintBuiltinName, kurtosis_print.GeneratePrintBuiltin(instructionsQueue, interpreter.recipeExecutor, interpreter.serviceNetwork)),
		remove_service.RemoveServiceBuiltinName:          starlark.NewBuiltin(remove_service.RemoveServiceBuiltinName, remove_service.GenerateRemoveServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		render_templates.RenderTemplatesBuiltinName:      starlark.NewBuiltin(render_templates.RenderTemplatesBuiltinName, render_templates.GenerateRenderTemplatesBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		store_service_files.StoreServiceFilesBuiltinName: starlark.NewBuiltin(store_service_files.StoreServiceFilesBuiltinName, store_service_files.GenerateStoreServiceFilesBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		define_fact.DefineFactBuiltinName:                starlark.NewBuiltin(define_fact.DefineFactBuiltinName, define_fact.GenerateDefineFactBuiltin(instructionsQueue, interpreter.factsEngine)),
		upload_files.UploadFilesBuiltinName:              starlark.NewBuiltin(upload_files.UploadFilesBuiltinName, upload_files.GenerateUploadFilesBuiltin(instructionsQueue, interpreter.moduleContentProvider, interpreter.serviceNetwork)),
		wait.WaitBuiltinName:                             starlark.NewBuiltin(wait.WaitBuiltinName, wait.GenerateWaitBuiltin(instructionsQueue, interpreter.factsEngine)),

		// Kurtosis custom builtins - pure interpretation-time helpers. Do not add any instructions to the queue
		import_module.ImportModuleBuiltinName: starlark.NewBuiltin(import_module.ImportModuleBuiltinName, import_module.GenerateImportBuiltin(recursiveInterpretForModuleLoading, interpreter.moduleContentProvider, interpreter.moduleGlobalsCache)),
		read_file.ReadFileBuiltinName:         starlark.NewBuiltin(read_file.ReadFileBuiltinName, read_file.GenerateReadFileBuiltin(interpreter.moduleContentProvider)),
	}
	return predeclared
}

// This method handles the different cases a Startosis module can be executed. Here are the different cases handled:
// - For a Kurtosis Package, the run method will always receive input args. If none were passed through the CLI params, empty JSON will be used
// - For a standalone Kurtosis script however, no params can be passed. It will fail if it is the case
func (interpreter *StartosisInterpreter) addInputArgsToPredeclared(thread *starlark.Thread, packageId string, serializedJsonArgs string, predeclared *starlark.StringDict) *startosis_errors.InterpretationError {
	if packageId == startosis_constants.PackageIdPlaceholderForStandaloneScript && serializedJsonArgs == startosis_constants.EmptyInputArgs {
		(*predeclared)[startosis_constants.MainInputArgName] = starlark.None
		return nil
	}
	// it is a module, and it has input args -> deserialize the JSON input and add it as a struct to the predeclared
	deserializedArgs, interpretationError := package_io.DeserializeArgs(thread, serializedJsonArgs)
	if interpretationError != nil {
		return interpretationError
	}
	(*predeclared)[startosis_constants.MainInputArgName] = deserializedArgs
	return nil
}

func makeLoadFunction() func(_ *starlark.Thread, packageId string) (starlark.StringDict, error) {
	return func(_ *starlark.Thread, _ string) (starlark.StringDict, error) {
		return nil, startosis_errors.NewInterpretationError("'load(\"path/to/file.star\", var_in_file=\"var_in_file\")' statement is not available in Kurtosis. Please use instead `module = import(\"path/to/file.star\")` and then `module.var_in_file`")
	}
}

func makePrintFunction() func(*starlark.Thread, string) {
	return func(_ *starlark.Thread, msg string) {
		// the `print` function must be overriden with the custom kurtosis_print instruction in the predeclared map
		panic("The print function does not function correctly. This is a Kurtosis bug")
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
		return startosis_errors.NewInterpretationErrorWithCustomMsg(
			stacktrace,
			"Evaluation error: %s",
			slError.Unwrap().Error(),
		)
	case *startosis_errors.InterpretationError:
		// If it's already an interpretation error -> nothing to convert
		return slError
	}
	return startosis_errors.NewInterpretationError("UnknownError: %s\n", err.Error())
}
