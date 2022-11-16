package startosis_engine

import (
	"bytes"
	"context"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/facts_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_module"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/define_fact"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_files_from_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/wait"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/proto_compiler"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	starlarkproto "go.starlark.net/lib/proto"
	"go.starlark.net/lib/time"
	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"path/filepath"
	"strings"
	"sync"
)

const (
	starlarkGoThreadName = "Startosis interpreter thread"

	nonBreakingSpaceChar = "\u00a0"
	regularSpaceChar     = " "

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Startosis script. Listing each of them below."
)

type StartosisInterpreter struct {
	// This is mutex protected as interpreting two different scripts in parallel could potentially cause
	// problems with the moduleGlobalsCache & moduleContentProvider. Fixing this is quite complicated, which we decided not to do.
	mutex              *sync.Mutex
	serviceNetwork     service_network.ServiceNetwork
	factsEngine        *facts_engine.FactsEngine
	moduleGlobalsCache map[string]*startosis_modules.ModuleCacheEntry
	// TODO AUTH there will be a leak here in case people with different repo visibility access a module
	moduleContentProvider startosis_modules.ModuleContentProvider
	protoFileStore        *proto_compiler.ProtoFileStore
}

type SerializedInterpretationOutput string

func NewStartosisInterpreter(serviceNetwork service_network.ServiceNetwork, moduleContentProvider startosis_modules.ModuleContentProvider) *StartosisInterpreter {
	return &StartosisInterpreter{
		mutex:                 &sync.Mutex{},
		serviceNetwork:        serviceNetwork,
		moduleContentProvider: moduleContentProvider,
		moduleGlobalsCache:    make(map[string]*startosis_modules.ModuleCacheEntry),
		protoFileStore:        proto_compiler.NewProtoFileStore(moduleContentProvider),
	}
}

func NewStartosisInterpreterWithFacts(serviceNetwork service_network.ServiceNetwork, factsEngine *facts_engine.FactsEngine, moduleContentProvider startosis_modules.ModuleContentProvider) *StartosisInterpreter {
	return &StartosisInterpreter{
		mutex:                 &sync.Mutex{},
		serviceNetwork:        serviceNetwork,
		moduleContentProvider: moduleContentProvider,
		moduleGlobalsCache:    make(map[string]*startosis_modules.ModuleCacheEntry),
		factsEngine:           factsEngine,
		protoFileStore:        proto_compiler.NewProtoFileStore(moduleContentProvider),
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
func (interpreter *StartosisInterpreter) Interpret(ctx context.Context, moduleId string, serializedStartosis string, serializedJsonParams string) (SerializedInterpretationOutput, *startosis_errors.InterpretationError, []kurtosis_instruction.KurtosisInstruction) {
	interpreter.mutex.Lock()
	defer interpreter.mutex.Unlock()
	var scriptOutputBuffer bytes.Buffer
	var instructionsQueue []kurtosis_instruction.KurtosisInstruction

	_, err := interpreter.interpretInternal(moduleId, serializedStartosis, serializedJsonParams, &instructionsQueue, &scriptOutputBuffer)
	if err != nil {
		return "", generateInterpretationError(err), nil
	}

	logrus.Debugf("Successfully interpreted Startosis code into instruction queue: \n%s", instructionsQueue)
	return SerializedInterpretationOutput(scriptOutputBuffer.String()), nil, instructionsQueue
}

func (interpreter *StartosisInterpreter) interpretInternal(moduleId string, serializedStartosis string, serializedJsonParams string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) (starlark.StringDict, error) {
	thread, predeclared := interpreter.buildBindings(starlarkGoThreadName, instructionsQueue, scriptOutputBuffer)

	if interpretationError := interpreter.addInputArgsToPredeclared(moduleId, serializedJsonParams, predeclared); interpretationError != nil {
		return nil, interpretationError
	}

	globalVariables, err := starlark.ExecFile(thread, moduleId, serializedStartosis, *predeclared)
	if err != nil {
		return nil, err
	}
	return globalVariables, nil
}

func (interpreter *StartosisInterpreter) buildBindings(threadName string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) (*starlark.Thread, *starlark.StringDict) {
	thread := &starlark.Thread{
		Name:  threadName,
		Load:  interpreter.makeLoadFunction(instructionsQueue, scriptOutputBuffer),
		Print: makePrintFunction(scriptOutputBuffer),
	}

	recursiveInterpretForModuleLoading := func(moduleId string, serializedStartosis string) (starlark.StringDict, error) {
		return interpreter.interpretInternal(moduleId, serializedStartosis, EmptyInputArgs, instructionsQueue, scriptOutputBuffer)
	}

	predeclared := &starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkproto.Module.Name:         starlarkproto.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		time.Module.Name:                  time.Module,

		// Kurtosis instructions - will push instructions to the queue that will affect the enclave state at execution
		add_service.AddServiceBuiltinName:                        starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		define_fact.DefineFactBuiltinName:                        starlark.NewBuiltin(define_fact.DefineFactBuiltinName, define_fact.GenerateDefineFactBuiltin(instructionsQueue, interpreter.factsEngine)),
		exec.ExecBuiltinName:                                     starlark.NewBuiltin(exec.ExecBuiltinName, exec.GenerateExecBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		remove_service.RemoveServiceBuiltinName:                  starlark.NewBuiltin(remove_service.RemoveServiceBuiltinName, remove_service.GenerateRemoveServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		render_templates.RenderTemplatesBuiltinName:              starlark.NewBuiltin(render_templates.RenderTemplatesBuiltinName, render_templates.GenerateRenderTemplatesBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		store_files_from_service.StoreFileFromServiceBuiltinName: starlark.NewBuiltin(store_files_from_service.StoreFileFromServiceBuiltinName, store_files_from_service.GenerateStoreFilesFromServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		upload_files.UploadFilesBuiltinName:                      starlark.NewBuiltin(upload_files.UploadFilesBuiltinName, upload_files.GenerateUploadFilesBuiltin(instructionsQueue, interpreter.moduleContentProvider, interpreter.serviceNetwork)),
		wait.WaitBuiltinName:                                     starlark.NewBuiltin(wait.WaitBuiltinName, wait.GenerateWaitBuiltin(instructionsQueue, interpreter.factsEngine)),

		// Kurtosis custom builtins - pure interpretation-time helpers. Do not add any instructions to the queue
		import_module.ImportModuleBuiltinName: starlark.NewBuiltin(import_module.ImportModuleBuiltinName, import_module.GenerateImportScriptBuiltin(recursiveInterpretForModuleLoading, interpreter.moduleContentProvider, interpreter.moduleGlobalsCache)),
		import_types.ImportTypesBuiltinName:   starlark.NewBuiltin(import_types.ImportTypesBuiltinName, import_types.GenerateImportTypesBuiltin(interpreter.protoFileStore)),
		read_file.ReadFileBuiltinName:         starlark.NewBuiltin(read_file.ReadFileBuiltinName, read_file.GenerateReadFileBuiltin(interpreter.moduleContentProvider)),
	}
	return thread, predeclared
}

// This method handles the different cases a Startosis module can be executed. Here are the different cases handled:
// - If there's no `types.proto` in the module, then the CLI should be called with no execute params (or default empty `{}`). The main function will not receive any arg. It should implement a simple `def main()`.
// - If there's a `types.proto` in the module and this `types.proto` does not define a `ModuleInput` type, then the CLI should be called with no execute params (or default empty `{}`). The main function must have an `input_args` param and this `input_args` will be an empty object.
// - If there's a `types.proto` in the module and this `types.proto` defines a `ModuleInput` type, then the CLI should be called with execute params (otherwise default values will be used). The main function must have an `input_args` param and `input_args` will be built from the execute params passed to the CLI
func (interpreter *StartosisInterpreter) addInputArgsToPredeclared(moduleId string, serializedJsonParams string, predeclared *starlark.StringDict) *startosis_errors.InterpretationError {
	if moduleId == ModuleIdPlaceholderForStandaloneScripts && serializedJsonParams == EmptyInputArgs {
		(*predeclared)[MainInputArgName] = starlark.None
		return nil
	}
	if moduleId == ModuleIdPlaceholderForStandaloneScripts && serializedJsonParams != EmptyInputArgs {
		// This _cannot_ be reached right now. Just being nice in the logs in case it can be hit in the future
		return startosis_errors.NewInterpretationError("Passing parameter to a standalone script is not yet supported in Kurtosis.")
	}

	// Get descriptor for type "ModuleInput" in the module types.proto file
	protoTypesFile := strings.Join([]string{moduleId, TypesFileName}, string(filepath.Separator))
	fileStore, interpretationError := interpreter.protoFileStore.LoadProtoFile(protoTypesFile)
	if interpretationError != nil && serializedJsonParams == EmptyInputArgs {
		// If am empty param was passed to the script, then it's valid to not have a types.proto inside the module
		(*predeclared)[MainInputArgName] = starlark.None
		return nil
	}
	if interpretationError != nil {
		// TODO(gb): rework the error piping here to include the compiler error message (see https://github.com/kurtosis-tech/kurtosis/issues/270)
		return startosis_errors.WrapWithInterpretationError(interpretationError, "A non empty parameter was passed to the module '%s' but the module doesn't contain a valid '%s' file (it is either absent of invalid). To be able to pass a parameter to a Kurtosis module, please define a '%s' type in the module's '%s' file", moduleId, TypesFileName, ModuleInputTypeName, TypesFileName)
	}
	reflectMessageDescriptor, err := fileStore.FindDescriptorByName(ModuleInputTypeName)
	if err != nil && serializedJsonParams == EmptyInputArgs {
		// If am empty param was passed to the script, then it's valid to not have a ModuleInput type defined in the module's types.proto
		(*predeclared)[MainInputArgName] = starlark.None
		return nil
	}
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "A non empty parameter was passed to the module '%s' but '%s' type is not defined in the module's '%s' file. To be able to pass a parameter to a Kurtosis module, please define a '%s' type in the module's '%s' file", moduleId, ModuleInputTypeName, TypesFileName, ModuleInputTypeName, TypesFileName)
	}
	messageDescriptor, ok := reflectMessageDescriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return startosis_errors.NewInterpretationError("Cannot cast protoreflect.Descriptor to protoreflect.MessageDescriptor. This is a very unexpected product bug (module ID: '%s', type: '%s')", moduleId, ModuleInputTypeName)
	}

	// Unmarshall the serialized params into this ModuleInput type
	message := dynamicpb.NewMessage(messageDescriptor)
	if err = protojson.Unmarshal([]byte(serializedJsonParams), message); err != nil {
		// protoc compiler error can contain non-breaking spaces. For simplicity, convert them to regular spaces
		protocErrorMsg := strings.ReplaceAll(err.Error(), nonBreakingSpaceChar, regularSpaceChar)
		return startosis_errors.NewInterpretationError("Module parameter shape does not fit the module expected input type (module: '%s'). Parameter was: \n%v\nError was: \n%v", moduleId, serializedJsonParams, protocErrorMsg)
	}

	// Convert the proto.Message into a starlarkproto.Message
	protobufMarshalledMessage, err := proto.Marshal(message)
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Unable to serialize the '%s' type of module '%s'. This is unexpected.", ModuleInputTypeName, moduleId)
	}
	starlarkMessage, err := starlarkproto.Unmarshal(message.Descriptor(), protobufMarshalledMessage)
	if err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Unable to serialize the '%s' type of module '%s'. This is unexpected.", moduleId, serializedJsonParams)
	}
	(*predeclared)[MainInputArgName] = starlarkMessage
	return nil
}

// this will be removed soon in favor of import_module()
// When it is, replace it with a nice InterpretationError throwing method like:
//
//		func makeLoadFunction(_ *starlark.Thread, _ string) (starlark.StringDict, error) {
//	 		return nil, startosis_errors.NewInterpretationError("'load(\"path/to/file.star\", module=\"module\")' statement is not available in Kurtosis. Please use instead `module = import_module(\"path/to/file.star\")`")
//		}
func (interpreter *StartosisInterpreter) makeLoadFunction(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) func(_ *starlark.Thread, moduleID string) (starlark.StringDict, error) {
	logrus.Warnf("`load()` statement is deprecated and will be removed soon. Please migrate to `import_module()`")
	return func(thread *starlark.Thread, moduleID string) (starlark.StringDict, error) {
		module, err := import_module.GenerateImportScriptBuiltin(
			func(moduleId string, serializedStartosis string) (starlark.StringDict, error) {
				return interpreter.interpretInternal(moduleId, serializedStartosis, EmptyInputArgs, instructionsQueue, scriptOutputBuffer)
			},
			interpreter.moduleContentProvider,
			interpreter.moduleGlobalsCache,
		)(thread, nil, starlark.Tuple{starlark.String(moduleID)}, []starlark.Tuple{})
		if err != nil {
			return nil, err
		}
		starlarkModule, ok := module.(*starlarkstruct.Module)
		if !ok {
			return nil, stacktrace.NewError("Unable to cast output of import_module builtin to a Starlark Module object. This is unexpected")
		}
		return starlarkModule.Members, nil
	}
}

func makePrintFunction(scriptOutputBuffer *bytes.Buffer) func(*starlark.Thread, string) {
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
		return startosis_errors.NewInterpretationErrorWithCustomMsg(stacktrace, multipleInterpretationErrorMsg)
	case *starlark.EvalError:
		stacktrace := make([]startosis_errors.CallFrame, 0)
		for _, callStack := range slError.CallStack {
			stacktrace = append(stacktrace, *startosis_errors.NewCallFrame(callStack.Name, startosis_errors.NewScriptPosition(callStack.Pos.Line, callStack.Pos.Col)))
		}
		return startosis_errors.NewInterpretationErrorWithCustomMsg(
			stacktrace,
			"Evaluation error: %s",
			slError.Unwrap().Error(),
		)
	case *startosis_errors.InterpretationError:
		// TODO(gb): This is because interpretInternal returns an InterpretationError when adding the input_args
		//  This won't be the case anymore when we remove protobuf, so we will be able to remove it if we want to
		return slError
	}
	return startosis_errors.NewInterpretationError("UnknownError: %s\n", err.Error())
}
