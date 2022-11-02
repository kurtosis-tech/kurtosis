package startosis_engine

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/import_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/read_file"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_files_from_service"
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
	starlarkGoThreadName                 = "Startosis interpreter thread"
	starlarkFilenamePlaceholderAsNotUsed = "FILENAME_NOT_USED"

	nonBreakingSpaceChar = "\u00a0"
	regularSpaceChar     = " "

	multipleInterpretationErrorMsg = "Multiple errors caught interpreting the Startosis script. Listing each of them below."
)

type StartosisInterpreter struct {
	// This is mutex protected as interpreting two different scripts in parallel could potentially cause
	// problems with the moduleGlobalsCache & moduleContentProvider. Fixing this is quite complicated, which we decided not to do.
	mutex              *sync.Mutex
	serviceNetwork     service_network.ServiceNetwork
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
	thread, predeclared := interpreter.buildBindings(starlarkGoThreadName, &instructionsQueue, &scriptOutputBuffer)

	if interpretationError := interpreter.addInputArgsToPredeclared(moduleId, serializedJsonParams, predeclared); interpretationError != nil {
		return "", interpretationError, nil
	}

	_, err := starlark.ExecFile(thread, starlarkFilenamePlaceholderAsNotUsed, serializedStartosis, *predeclared)
	if err != nil {
		return "", generateInterpretationError(err), nil
	}
	logrus.Debugf("Successfully interpreted Startosis code into instruction queue: \n%s", instructionsQueue)
	return SerializedInterpretationOutput(scriptOutputBuffer.String()), nil, instructionsQueue
}

func (interpreter *StartosisInterpreter) buildBindings(threadName string, instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) (*starlark.Thread, *starlark.StringDict) {
	thread := &starlark.Thread{
		Name:  threadName,
		Load:  interpreter.makeLoadFunction(instructionsQueue, scriptOutputBuffer),
		Print: makePrintFunction(scriptOutputBuffer),
	}

	predeclared := &starlark.StringDict{
		starlarkstruct.Default.GoString():                        starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
		add_service.AddServiceBuiltinName:                        starlark.NewBuiltin(add_service.AddServiceBuiltinName, add_service.GenerateAddServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		exec.ExecBuiltinName:                                     starlark.NewBuiltin(exec.ExecBuiltinName, exec.GenerateExecBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		store_files_from_service.StoreFileFromServiceBuiltinName: starlark.NewBuiltin(store_files_from_service.StoreFileFromServiceBuiltinName, store_files_from_service.GenerateStoreFilesFromServiceBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		read_file.ReadFileBuiltinName:                            starlark.NewBuiltin(read_file.ReadFileBuiltinName, read_file.GenerateReadFileBuiltin(instructionsQueue, interpreter.moduleContentProvider)),
		render_templates.RenderTemplatesBuiltinName:              starlark.NewBuiltin(render_templates.RenderTemplatesBuiltinName, render_templates.GenerateRenderTemplatesBuiltin(instructionsQueue, interpreter.serviceNetwork)),
		starlarkjson.Module.Name:                                 starlarkjson.Module,
		import_types.ImportTypesBuiltinName:                      starlark.NewBuiltin(import_types.ImportTypesBuiltinName, import_types.GenerateImportTypesBuiltin(interpreter.protoFileStore)),
		time.Module.Name:                                         time.Module,
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
	fileStore, err := interpreter.protoFileStore.LoadProtoFile(protoTypesFile)
	if err != nil && serializedJsonParams == EmptyInputArgs {
		// If am empty param was passed to the script, then it's valid to not have a types.proto inside the module
		(*predeclared)[MainInputArgName] = starlark.None
		return nil
	}
	if err != nil {
		return startosis_errors.NewInterpretationError("File '%s' not found at the root of module '%s' but a non empty parameter was passed. This is allowed to define a module with no '%s', but it should be always be called with an empty parameter", TypesFileName, moduleId, TypesFileName)
	}
	reflectMessageDescriptor, err := fileStore.FindDescriptorByName(ModuleInputTypeName)
	if err != nil && serializedJsonParams == EmptyInputArgs {
		// If am empty param was passed to the script, then it's valid to not have a ModuleInput type defined in the module's types.proto
		(*predeclared)[MainInputArgName] = starlark.None
		return nil
	}
	if err != nil {
		return startosis_errors.NewInterpretationError("Type '%s' cannot be found in type file '%s' for module '%s' but a non empty parameter was passed. When some parameters are passed to a module, there must be a `%s` type defined in the module's '%s' file", ModuleInputTypeName, TypesFileName, moduleId, MainInputArgName, ModuleInputTypeName)
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
		logrus.Error(stacktrace.Propagate(err, "Unable to marshal proto message '%s' from module ID '%s'", ModuleInputTypeName, moduleId).Error())
		return startosis_errors.NewInterpretationError("Unable to serialize the '%s' type of module '%s'. This is unexpected. More info will be logged to Kurtosis core", ModuleInputTypeName, moduleId)
	}
	starlarkMessage, err := starlarkproto.Unmarshal(message.Descriptor(), protobufMarshalledMessage)
	if err != nil {
		logrus.Error(stacktrace.Propagate(err, "Unable to convert proto message '%s' to a starlark proto message from module ID '%s'", ModuleInputTypeName, moduleId).Error())
		return startosis_errors.NewInterpretationError("Unable to serialize the '%s' type of module '%s'. This is unexpected. More info will be logged to Kurtosis core", moduleId, serializedJsonParams)
	}
	(*predeclared)[MainInputArgName] = starlarkMessage
	return nil
}

/*
	   makeLoadFunction This function returns a sequential (not parallel) implementation of `load` in Starlark
	   This function takes in an instructionsQueue, scriptOutputBuffer & returns a closed function that implements Starlark loading, with the custom provider

		instructionsQueue -> the instructions Queue from the parent thread to add instructions to
		scriptsOutputBuffer -> the scripts output buffer from the parent thread where output form interpreted scripts will be written to

We wrap the function to provide a closure for the above two arguments, as we can't change the signature of the returned function

How does the returned function work?
1. The function first checks whether a module is currently, loading, if so then there's cycle and it errors immediately,
2. The function checks then interpreter.modulesGlobalCache for preloaded symbols or previous interpretation errors, if there is either of them it returns
3. At this point this is a new module for this instance of the interpreter, we set load to in progress (this is useful for cycle detection).
5. We defer undo the loading in case there is a failure loading the contents of the module. We don't want it to be the loading state as the next call to load the module would incorrectly return a cycle error.
6. We then load the contents of the module file using a custom provider which fetches Git repos.
7. After we have the contents of the module, we create a Starlark thread, and call ExecFile on that thread, which interprets the other module and returns its symbols
8. At this point we cache the symbols from the loaded module
9. We now return the contents of the module and any interpretation errors
This function is recursive in the sense, to load a module that loads modules we call the same function
*/
func (interpreter *StartosisInterpreter) makeLoadFunction(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, scriptOutputBuffer *bytes.Buffer) func(_ *starlark.Thread, moduleID string) (starlark.StringDict, error) {
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
		contents, err := interpreter.moduleContentProvider.GetModuleContents(moduleID)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while loading the module '%v'", moduleID)
		}

		thread, predeclared := interpreter.buildBindings(fmt.Sprintf("%v:%v", starlarkGoThreadName, moduleID), instructionsQueue, scriptOutputBuffer)
		globalVariables, err := starlark.ExecFile(thread, moduleID, contents, *predeclared)
		// the above error goes unchecked as it needs to be persisted to the cache and then returned to the parent loader

		// Update the cache.
		entry = startosis_modules.NewModuleCacheEntry(globalVariables, err)
		interpreter.moduleGlobalsCache[moduleID] = entry

		shouldUnsetLoadInProgress = false
		return entry.GetGlobalVariables(), entry.GetError()
		// this error isn't propagated as its returned to the interpreter & persisted in the cache
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
	}
	return startosis_errors.NewInterpretationError("UnknownError: %s\n", err.Error())
}
