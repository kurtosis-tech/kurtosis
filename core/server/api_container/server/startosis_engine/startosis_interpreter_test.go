package startosis_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/upload_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/mock_module_content_provider"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"path/filepath"
	"strings"
	"testing"
)

var testServiceNetwork service_network.ServiceNetwork = service_network.NewEmptyMockServiceNetwork()

const (
	testContainerImageName = "kurtosistech/example-datastore-server"
)

var (
	defaultEntryPointArgs              []string          = nil
	defaultCmdArgs                     []string          = nil
	defaultEnvVars                     map[string]string = nil
	defaultPrivateIPAddressPlaceholder                   = ""
)

func TestStartosisInterpreter_SimplePrintScript(t *testing.T) {
	testString := "Hello World!"
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	interpreter := startosisInterpreter
	script := `
print("` + testString + `")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 1) // Only the print statement
	require.Nil(t, interpretationError)

	expectedOutput := testString + `
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ScriptFailingSingleError(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

unknownInstruction()
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
		},
		"Multiple errors caught interpreting the Startosis script. Listing each of them below.",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingMultipleErrors(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

unknownInstruction()
print(unknownVariable)

unknownInstruction2()
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(5, 7)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(7, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingSyntaxError(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(4, 5)),
		},
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstruction(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	privateIPAddressPlaceholder := "MAGICAL_PLACEHOLDER_TO_REPLACE"
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	image = "` + testContainerImageName + `",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	},
	private_ip_address_placeholder = "` + privateIPAddressPlaceholder + `"
)
datastore_service = add_service(service_id = service_id, config = config)
print("The grpc port is " + str(datastore_service.ports["grpc"].number))
print("The grpc port protocol is " + datastore_service.ports["grpc"].protocol)
print("The datastore service ip address is " + datastore_service.ip_address)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 6)
	require.Nil(t, interpretationError)

	addServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 14, 32, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, privateIPAddressPlaceholder)
	require.Equal(t, addServiceInstruction, instructions[2])

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The grpc port is 1323
The grpc port protocol is TCP
The datastore service ip address is {{kurtosis:example-datastore-server.ip_address}}
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionMissingContainerName(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	# /!\ /!\ missing container name /!\ /!\
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, config = config)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
		"Evaluation error: Missing value 'image' as element of the struct object 'config'",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionTypoInProtocol(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	image = "` + testContainerImageName + `",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCPK") # typo in protocol
	}
)
add_service(service_id = service_id, config = config)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
		"Evaluation error: Port protocol should be one of TCP, SCTP, UDP",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionPortNumberAsString(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	image = "` + testContainerImageName + `",
	ports = {
		"grpc": struct(number = "1234", protocol = "TCP") # port number should be an int
	}
)
add_service(service_id = service_id, config = config)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Empty(t, instructions)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
		"Evaluation error: Argument 'number' is expected to be an integer. Got starlark.String",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidScriptWithMultipleInstructions(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services():
    for i in range(len(ports)):
        unique_service_id = service_id + "-" + str(i)
        print("Adding service " + unique_service_id)
        config = struct(
			image = "` + testContainerImageName + `",
			ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, config = config)

deploy_datastore_services()
print("Done!")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 8)
	require.Nil(t, interpretationError)

	addServiceInstruction0 := createSimpleAddServiceInstruction(t, "example-datastore-server-0", testContainerImageName, 1323, 20, 20, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	addServiceInstruction1 := createSimpleAddServiceInstruction(t, "example-datastore-server-1", testContainerImageName, 1324, 20, 20, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	addServiceInstruction2 := createSimpleAddServiceInstruction(t, "example-datastore-server-2", testContainerImageName, 1325, 20, 20, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	require.Equal(t, addServiceInstruction0, instructions[2])
	require.Equal(t, addServiceInstruction1, instructions[4])
	require.Equal(t, addServiceInstruction2, instructions[6])

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_SimpleLoading_DeprecatedForBackwardCompatibility(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + barModulePath + `", "a")
print("Hello " + a)
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Len(t, instructions, 1) // Only the print statement
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_SimpleImport(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
my_module = import_module("` + barModulePath + `")
print("Hello " + my_module.a)
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Len(t, instructions, 1) // Only the print statement
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_TransitiveLoading(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `a="World!"`
	moduleDooWhichLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleDooWhichLoadsModuleBar] = `module_bar = import_module("` + moduleBar + `")
b = "Hello " + module_bar.a
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
module_doo = import_module("` + moduleDooWhichLoadsModuleBar + `")
print(module_doo.b)

`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 1) // Only the print statement
	require.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_FailsOnCycle(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBarLoadsModuleDoo := "github.com/foo/bar/lib.star"
	moduleDooLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleBarLoadsModuleDoo] = `module_doo = import_module("` + moduleDooLoadsModuleBar + `")
a = "Hello" + module_doo.b`
	seedModules[moduleDooLoadsModuleBar] = `module_bar = import_module("` + moduleBarLoadsModuleDoo + `")
b = "Hello " + module_bar.a
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
module_doo = import_module("` + moduleDooLoadsModuleBar + `")
print(module_doo.b)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Empty(t, instructions) // No kurtosis instruction
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(1, 27)),
			*startosis_errors.NewCallFrame("import_module", startosis_errors.NewScriptPosition(0, 0)),
		},
		"Evaluation error: There's a cycle in the import_module calls",
	).ToAPIType()
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_FailsOnNonExistentModule(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	nonExistentModule := "github.com/non/existent/module.star"
	script := `
my_module = import_module("` + nonExistentModule + `")
print(my_module.b)
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Empty(t, instructions) // No kurtosis instruction

	errorMsg := `Evaluation error: An error occurred while loading the module '` + nonExistentModule + `'
	Caused by: Module '` + nonExistentModule + `' not found`
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 26)),
			*startosis_errors.NewCallFrame("import_module", startosis_errors.NewScriptPosition(0, 0)),
		},
		errorMsg,
	).ToAPIType()
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ImportingAValidModuleThatPreviouslyFailedToLoadSucceeds(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	barModulePath := "github.com/foo/bar/lib.star"
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
my_module = import_module("` + barModulePath + `")
print("Hello " + my_module.a)
`

	// assert that first load fails
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Nil(t, instructions)
	assert.NotNil(t, interpretationError)

	barModuleContents := "a=\"World!\""
	require.Nil(t, moduleContentProvider.AddFileContent(barModulePath, barModuleContents))
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	instructions, interpretationError = interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Nil(t, interpretationError)
	assert.Len(t, instructions, 1) // The print statement
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleScriptWithImportedStruct(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing config")
config = struct(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
module_bar = import_module("` + moduleBar + `")
print("Starting Startosis script!")

print("Adding service " + module_bar.service_id)
add_service(service_id = module_bar.service_id, config = module_bar.config)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 4)
	require.Nil(t, interpretationError)

	addServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 6, 12, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	require.Equal(t, addServiceInstruction, instructions[3])

	expectedOutput := `Constructing config
Starting Startosis script!
Adding service example-datastore-server
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidScriptWithFunctionsImportedFromOtherModule(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services():
    for i in range(len(ports)):
        unique_service_id = service_id + "-" + str(i)
        print("Adding service " + unique_service_id)
        config = struct(
			image = "kurtosistech/example-datastore-server",
			ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, config = config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
datastore_module = import_module("` + moduleBar + `")
print("Starting Startosis script!")

datastore_module.deploy_datastore_services()
print("Done!")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 8)
	require.Nil(t, interpretationError)

	addServiceInstruction0 := createSimpleAddServiceInstruction(t, "example-datastore-server-0", testContainerImageName, 1323, 18, 20, moduleBar, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	addServiceInstruction1 := createSimpleAddServiceInstruction(t, "example-datastore-server-1", testContainerImageName, 1324, 18, 20, moduleBar, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	addServiceInstruction2 := createSimpleAddServiceInstruction(t, "example-datastore-server-2", testContainerImageName, 1325, 18, 20, moduleBar, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	require.Equal(t, addServiceInstruction0, instructions[2])
	require.Equal(t, addServiceInstruction1, instructions[4])
	require.Equal(t, addServiceInstruction2, instructions[6])

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ImportModuleWithNoGlobalVariables(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "print(\"Hello\")",
	}
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
my_module = import_module("` + barModulePath + `")
print("World!")
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	assert.Len(t, instructions, 2)
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello
World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_AddServiceInOtherModulePopulatesQueue(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing config")
config = struct(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, config = config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
import_module("` + moduleBar + `")
print("Starting Startosis script!")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 4)
	require.Nil(t, interpretationError)

	addServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 11, 12, moduleBar, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	require.Equal(t, addServiceInstruction, instructions[2])

	expectedOutput := `Constructing config
Adding service example-datastore-server
Starting Startosis script!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_TestInstructionQueueAndOutputBufferDontHaveDupesInterpretingAnotherScript(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing config")
config = struct(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, config = config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seedModules))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	scriptA := `
import_module("` + moduleBar + `")
print("Starting Startosis script!")
`
	addServiceInstructionFromScriptA := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 11, 12, moduleBar, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	expectedOutputFromScriptA := `Constructing config
Adding service example-datastore-server
Starting Startosis script!
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, scriptA, EmptyInputArgs)
	require.Len(t, instructions, 4)
	require.Nil(t, interpretationError)
	require.Equal(t, addServiceInstructionFromScriptA, instructions[2])
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutputFromScriptA)

	scriptB := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, config = config)
`
	addServiceInstructionFromScriptB := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 13, 12, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	expectedOutputFromScriptB := `Starting Startosis script!
Adding service example-datastore-server
`

	instructions, interpretationError = interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, scriptB, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 3)
	require.Equal(t, addServiceInstructionFromScriptB, instructions[2])
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutputFromScriptB)
}

func TestStartosisInterpreter_AddServiceWithEnvVarsCmdArgsAndEntryPointArgs(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")
service_id = "example-datastore-server"
print("Adding service " + service_id)
store_config = struct(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
datastore_service = add_service(service_id = service_id, config = store_config)
client_service_id = "example-datastore-client"
print("Adding service " + client_service_id)
client_config = struct(
	image = "kurtosistech/example-datastore-client",
	ports = {
		"grpc": struct(number = 1337, protocol = "TCP")
	},
	entrypoint = ["--store-port " + str(datastore_service.ports["grpc"].number), "--store-ip " + datastore_service.ip_address],
	cmd = ["ping", datastore_service.ip_address],
	env_vars = {"STORE_IP": datastore_service.ip_address}
)
add_service(service_id = client_service_id, config = client_config)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 5)

	dataSourceAddServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 1323, 11, 32, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)

	entryPointArgs := []string{"--store-port 1323", "--store-ip {{kurtosis:example-datastore-server.ip_address}}"}
	cmdArgs := []string{"ping", "{{kurtosis:example-datastore-server.ip_address}}"}
	envVars := map[string]string{"STORE_IP": "{{kurtosis:example-datastore-server.ip_address}}"}
	clientAddServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-client", "kurtosistech/example-datastore-client", 1337, 23, 12, ModuleIdPlaceholderForStandaloneScripts, entryPointArgs, cmdArgs, envVars, defaultPrivateIPAddressPlaceholder)

	require.Equal(t, dataSourceAddServiceInstruction, instructions[2])
	require.Equal(t, clientAddServiceInstruction, instructions[4])

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
Adding service example-datastore-client
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidExecScriptWithoutExitCodeDefaultsTo0(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Executing mkdir!")
exec(service_id = "example-datastore-server", command = ["mkdir", "/tmp/foo"])
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 2)
	require.Nil(t, interpretationError)

	execInstruction := exec.NewExecInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(3, 5, ModuleIdPlaceholderForStandaloneScripts),
		"example-datastore-server",
		[]string{"mkdir", "/tmp/foo"},
		0,
	)

	require.Equal(t, execInstruction, instructions[1])

	expectedOutput := `Executing mkdir!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_PassedExitCodeIsInterpretedCorrectly(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Executing mkdir!")
exec(service_id = "example-datastore-server", command = ["mkdir", "/tmp/foo"], expected_exit_code = -7)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 2)
	require.Nil(t, interpretationError)

	execInstruction := exec.NewExecInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(3, 5, ModuleIdPlaceholderForStandaloneScripts),
		"example-datastore-server",
		[]string{"mkdir", "/tmp/foo"},
		-7,
	)

	require.Equal(t, execInstruction, instructions[1])

	expectedOutput := `Executing mkdir!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_StoreFileFromService(t *testing.T) {
	testArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Storing file from service!")
artifact_uuid=store_service_files(service_id="example-datastore-server", src_path="/foo/bar", artifact_uuid="` + string(testArtifactUuid) + `")
print(artifact_uuid)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 3)

	storeInstruction := store_service_files.NewStoreServiceFilesInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(3, 38, ModuleIdPlaceholderForStandaloneScripts),
		"example-datastore-server",
		"/foo/bar",
		testArtifactUuid,
	)

	require.Equal(t, storeInstruction, instructions[1])

	expectedOutput := fmt.Sprintf(`Storing file from service!
%v
`, testArtifactUuid)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ReadFileFromGithub(t *testing.T) {
	srcPath := "github.com/foo/bar/static_files/main.txt"
	seed := map[string]string{
		srcPath: "this is a test string",
	}
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.BulkAddFileContent(seed))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Reading file from GitHub!")
file_contents=read_file("` + srcPath + `")
print(file_contents)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 2)

	expectedOutput := `Reading file from GitHub!
this is a test string
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_DefineFactAndWait(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreterWithFacts(testServiceNetwork, nil, moduleContentProvider)
	scriptFormatStr := `
define_fact(service_id="%v", fact_name="%v", fact_recipe=struct(method="GET", endpoint="/", port_id="http"))
wait(service_id="%v", fact_name="%v")
`
	serviceId := "service"
	factName := "fact"
	script := fmt.Sprintf(scriptFormatStr, serviceId, factName, serviceId, factName)
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.NotEmpty(t, instructions)
	validateScriptOutputFromPrintInstructions(t, instructions, "")
}

func TestStartosisInterpreter_RenderTemplates(t *testing.T) {
	testArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Rendering template to disk!")
template_data = {
			"Name" : "Stranger",
			"Answer": 6,
			"Numbers": [1, 2, 3],
			"UnixTimeStamp": 1257894000,
			"LargeFloat": 1231231243.43,
			"Alive": True
}
encoded_json = json.encode(template_data)
data = {
	"/foo/bar/test.txt" : {
		"template": "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}",
		"template_data_json": encoded_json
    }
}
artifact_uuid = render_templates(template_and_data_by_dest_rel_filepath = data, artifact_uuid = "` + string(testArtifactUuid) + `")
print(artifact_uuid)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 3)

	template := "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}"
	templateData := map[string]interface{}{"Name": "Stranger", "Answer": 6, "Numbers": []int{1, 2, 3}, "UnixTimeStamp": 1257894000, "LargeFloat": 1231231243.43, "Alive": true}
	serializedTemplateData := `{"Alive":true,"Answer":6,"LargeFloat":1.23123124343e+09,"Name":"Stranger","Numbers":[1,2,3],"UnixTimeStamp":1257894000}`
	templateDataAsJson, err := json.Marshal(templateData)
	require.Nil(t, err)
	templateAndData := binding_constructors.NewTemplateAndData(template, string(templateDataAsJson))
	templateAndDataByDestFilepath := map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData{
		"/foo/bar/test.txt": templateAndData,
	}

	templateAndDataValues := starlark.NewDict(1)
	fooBarTestValuesValues := starlark.NewDict(2)
	require.Nil(t, fooBarTestValuesValues.SetKey(starlark.String("template"), starlark.String("Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}")))
	require.Nil(t, fooBarTestValuesValues.SetKey(starlark.String("template_data_json"), starlark.String(serializedTemplateData)))
	fooBarTestValuesValues.Freeze()
	require.Nil(t, templateAndDataValues.SetKey(starlark.String("/foo/bar/test.txt"), fooBarTestValuesValues))
	templateAndDataValues.Freeze()

	renderInstruction := render_templates.NewRenderTemplatesInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(18, 33, ModuleIdPlaceholderForStandaloneScripts),
		templateAndDataByDestFilepath,
		starlark.StringDict{
			"template_and_data_by_dest_rel_filepath": templateAndDataValues,
			"artifact_uuid":                          starlark.String(testArtifactUuid),
		},
		testArtifactUuid,
	)

	require.Equal(t, renderInstruction, instructions[1])

	expectedOutput := fmt.Sprintf(`Rendering template to disk!
%v
`, testArtifactUuid)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ReadTypesFromProtoFileInScript(t *testing.T) {
	typesFilePath := "github.com/kurtosis/module/types.proto"
	typesFileContent := `
syntax = "proto3";
message TestType {
  string greetings = 1;
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
types = import_types(types_file = "github.com/kurtosis/module/types.proto")
test_type = types.TestType({
    "greetings": "Hello World!"
})
print(test_type)
print(test_type.greetings)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 2) // the print statement

	expectedOutput := `TestType(greetings="Hello World!")
Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ReadTypesFromProtoFile_FailuresWrongArgument(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
types = import_types(proto_types_file_bad_argument = "github.com/kurtosis/module/types.proto")
print("Hello world!")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedErrorString := "Evaluation error: Unable to parse arguments of command 'import_types'. It should be a non empty string argument pointing to the fully qualified .proto types file (i.e. \"github.com/kurtosis/module/types.proto\")"
	require.Contains(t, interpretationError.GetErrorMessage(), expectedErrorString)
}

func TestStartosisInterpreter_ReadTypesFromProtoFile_FailuresNoTypesFile(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
types = import_types("github.com/kurtosis/module/types.proto")
print("Hello world!")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 21)),
			*startosis_errors.NewCallFrame("import_types", startosis_errors.NewScriptPosition(0, 0)),
		},
		"Evaluation error: Unable to load types file github.com/kurtosis/module/types.proto. Is the corresponding type file present in the module?",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_InjectValidInputArgsToModule(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	typesFilePath := moduleId + "/types.proto"
	typesFileContent := `
syntax = "proto3";
message ModuleInput {
  string greetings = 1;
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print(input_args.greetings)
`
	serializedArgs := `{"greetings": "Hello World!"}`
	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, serializedArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1) // the print statement

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_InjectValidInputArgsToNonModuleScript(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	typesFilePath := moduleId + "/types.proto"
	typesFileContent := `
syntax = "proto3";
message ModuleInput {
  string greetings = 1;
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print(input_args.greetings)
`
	serializedArgs := `{"greetings": "Hello World!"}`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, serializedArgs)
	require.Empty(t, instructions)

	expectedError := binding_constructors.NewKurtosisInterpretationError("Passing parameter to a standalone script is not yet supported in Kurtosis.")
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_InvalidProtoFile(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	typesFilePath := moduleId + "/types.proto"
	typesFileContent := `
syntax "proto3"; // Missing '=' between 'syntax' and '"proto3"''
message ModuleInput {
  string greetings = 1
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	absFilePath, err := moduleContentProvider.GetOnDiskAbsoluteFilePath(typesFilePath)
	require.Nil(t, err)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
def main(input_args):
	print(input_args.greetings)
`
	serializedArgs := `{"greetings": "Hello World!"}`
	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, serializedArgs)
	require.Empty(t, instructions)

	expectedErrorMsg := fmt.Sprintf(`A non empty parameter was passed to the module 'github.com/kurtosis/module' but the module doesn't contain a valid 'types.proto' file (it is either absent of invalid). To be able to pass a parameter to a Kurtosis module, please define a 'ModuleInput' type in the module's 'types.proto' file
	Caused by: Unable to compile .proto file 'github.com/kurtosis/module/types.proto' (checked out at '%s'). Proto compiler output was: 
%s:2:8: Expected "=".
`, absFilePath, filepath.Base(absFilePath))
	require.Equal(t, expectedErrorMsg, interpretationError.GetErrorMessage())
}

func TestStartosisInterpreter_InjectValidInvalidInputArgsToModule_InvalidJson(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	typesFilePath := moduleId + "/types.proto"
	typesFileContent := `
syntax = "proto3";
message ModuleInput {
  string greetings = 1;
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print(input_args.greetings)
`
	serializedArgs := `"greetings": "Hello World!"` // Invalid JSON
	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, serializedArgs)
	require.Empty(t, instructions)

	expectedError := binding_constructors.NewKurtosisInterpretationError(`Module parameter shape does not fit the module expected input type (module: 'github.com/kurtosis/module'). Parameter was: 
"greetings": "Hello World!"
Error was: 
proto: syntax error (line 1:1): unexpected token "greetings"`)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_InjectValidInvalidInputArgsToModule_ValidJsonButWrongType(t *testing.T) {
	moduleId := "github.com/kurtosis/module"
	typesFilePath := moduleId + "/types.proto"
	typesFileContent := `
syntax = "proto3";
message ModuleInput {
  string greetings = 1;
}
`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	require.Nil(t, moduleContentProvider.AddFileContent(typesFilePath, typesFileContent))
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print(input_args.greetings)
`
	serializedArgs := `{"greetings": 3}` // greeting should be a string here
	instructions, interpretationError := interpreter.Interpret(context.Background(), moduleId, script, serializedArgs)
	require.Empty(t, instructions)

	expectedError := binding_constructors.NewKurtosisInterpretationError(`Module parameter shape does not fit the module expected input type (module: 'github.com/kurtosis/module'). Parameter was: 
{"greetings": 3}
Error was: 
proto: (line 1:15): invalid value for string type: 3`)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ThreeLevelNestedInstructionPositionTest(t *testing.T) {
	testArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	storeFileDefinitionPath := "github.com/kurtosis/store.star"
	storeFileContent := `
def store_for_me():
	print("In the store files instruction")
	artifact_uuid=store_service_files(service_id="example-datastore-server", src_path="/foo/bar", artifact_uuid = "` + string(testArtifactUuid) + `")
	return artifact_uuid
`

	moduleThatCallsStoreFile := "github.com/kurtosis/foo.star"
	moduleThatCallsStoreFileContent := `
store_for_me_module = import_module("github.com/kurtosis/store.star")
def call_store_for_me():
	print("In the module that calls store.star")
	return store_for_me_module.store_for_me()
	`

	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	err = moduleContentProvider.AddFileContent(storeFileDefinitionPath, storeFileContent)
	require.Nil(t, err)

	err = moduleContentProvider.AddFileContent(moduleThatCallsStoreFile, moduleThatCallsStoreFileContent)
	require.Nil(t, err)

	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
call_store_for_me_module = import_module("github.com/kurtosis/foo.star")
uuid = call_store_for_me_module.call_store_for_me()
print(uuid)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 4)

	storeInstruction := store_service_files.NewStoreServiceFilesInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(4, 39, storeFileDefinitionPath),
		"example-datastore-server",
		"/foo/bar",
		testArtifactUuid,
	)

	require.Equal(t, storeInstruction, instructions[2])

	expectedOutput := fmt.Sprintf(`In the module that calls store.star
In the store files instruction
%v
`, testArtifactUuid)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleRemoveService(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")
service_id = "example-datastore-server"
remove_service(service_id=service_id)
print("The service example-datastore-server has been removed")
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Len(t, instructions, 3)
	require.Nil(t, interpretationError)

	removeInstruction := remove_service.NewRemoveServiceInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(4, 15, ModuleIdPlaceholderForStandaloneScripts),
		"example-datastore-server",
	)

	require.Equal(t, removeInstruction, instructions[1])

	expectedOutput := `Starting Startosis script!
The service example-datastore-server has been removed
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_UploadGetsInterpretedCorrectly(t *testing.T) {
	filePath := "github.com/kurtosis/module/lib/lib.star"
	artifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
	require.Nil(t, err)
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	err = moduleContentProvider.AddFileContent(filePath, "fooBar")
	require.Nil(t, err)
	filePathOnDisk, err := moduleContentProvider.GetOnDiskAbsoluteFilePath(filePath)
	require.Nil(t, err)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `upload_files("` + filePath + `","` + string(artifactUuid) + `")
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1)
	validateScriptOutputFromPrintInstructions(t, instructions, "")

	expectedUploadInstruction := upload_files.NewUploadFilesInstruction(
		kurtosis_instruction.NewInstructionPosition(1, 13, ModuleIdPlaceholderForStandaloneScripts),
		testServiceNetwork, moduleContentProvider, filePath, filePathOnDisk, artifactUuid,
	)

	require.Equal(t, expectedUploadInstruction, instructions[0])
}

func TestStartosisInterpreter_NoPanicIfUploadIsPassedAPathNotOnDisk(t *testing.T) {
	filePath := "github.com/kurtosis/module/lib/lib.star"
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `upload_files("` + filePath + `")
`
	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, instructions)
	require.NotNil(t, interpretationError)
}

func TestStartosisInterpreter_NoPortsIsOkayForAddServiceInstruction(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider()
	defer moduleContentProvider.RemoveAll()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

config = struct(
	image = "` + testContainerImageName + `",
)
datastore_service = add_service(service_id = service_id, config = config)
print("The datastore service ip address is " + datastore_service.ip_address)
`

	instructions, interpretationError := interpreter.Interpret(context.Background(), ModuleIdPlaceholderForStandaloneScripts, script, EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Equal(t, 4, len(instructions))

	addServiceInstruction := createSimpleAddServiceInstruction(t, "example-datastore-server", testContainerImageName, 0, 10, 32, ModuleIdPlaceholderForStandaloneScripts, defaultEntryPointArgs, defaultCmdArgs, defaultEnvVars, defaultPrivateIPAddressPlaceholder)
	require.Equal(t, instructions[2], addServiceInstruction)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The datastore service ip address is {{kurtosis:example-datastore-server.ip_address}}
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

// #####################################################################################################################
//                                                  TEST HELPERS
// #####################################################################################################################

func createSimpleAddServiceInstruction(t *testing.T, serviceId service.ServiceID, imageName string, portNumber uint32, lineNumber int32, colNumber int32, fileName string, entryPointArgs []string, cmdArgs []string, envVars map[string]string, privateIPAddressPlaceholder string) *add_service.AddServiceInstruction {
	serviceConfigStringDict := starlark.StringDict{}
	serviceConfigStringDict["image"] = starlark.String(imageName)

	if portNumber != 0 {
		usedPortDict := starlark.NewDict(1)
		require.Nil(t, usedPortDict.SetKey(
			starlark.String("grpc"),
			starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
				"number":   starlark.MakeInt(int(portNumber)),
				"protocol": starlark.String("TCP"),
			})))
		serviceConfigStringDict["ports"] = usedPortDict
	}

	if entryPointArgs != nil {
		entryPointArgsValues := make([]starlark.Value, 0)
		for _, entryPointArg := range entryPointArgs {
			entryPointArgsValues = append(entryPointArgsValues, starlark.String(entryPointArg))
		}
		serviceConfigStringDict["entrypoint"] = starlark.NewList(entryPointArgsValues)
	}

	if cmdArgs != nil {
		cmdArgsValues := make([]starlark.Value, 0)
		for _, cmdArg := range cmdArgs {
			cmdArgsValues = append(cmdArgsValues, starlark.String(cmdArg))
		}
		serviceConfigStringDict["cmd"] = starlark.NewList(cmdArgsValues)
	}

	if envVars != nil {
		envVarsValues := starlark.NewDict(len(envVars))
		for key, value := range envVars {
			require.Nil(t, envVarsValues.SetKey(starlark.String(key), starlark.String(value)))
		}
		serviceConfigStringDict["env_vars"] = envVarsValues
	}

	if privateIPAddressPlaceholder != "" {
		privateIPAddressPlaceholderStarlarkValue := starlark.String(privateIPAddressPlaceholder)
		serviceConfigStringDict["private_ip_address_placeholder"] = privateIPAddressPlaceholderStarlarkValue
	}

	serviceConfigStruct := starlarkstruct.FromStringDict(starlarkstruct.Default, serviceConfigStringDict)
	serviceConfigStruct.Freeze()

	serviceConfigBuilder := services.NewServiceConfigBuilder(
		imageName,
	)

	if portNumber != 0 {
		serviceConfigBuilder.WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   portNumber,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		)
	}

	if entryPointArgs != nil {
		serviceConfigBuilder.WithEntryPointArgs(entryPointArgs)
	}
	if cmdArgs != nil {
		serviceConfigBuilder.WithCmdArgs(cmdArgs)
	}
	if envVars != nil {
		serviceConfigBuilder.WithEnvVars(envVars)
	}

	if privateIPAddressPlaceholder != "" {
		serviceConfigBuilder.WithPrivateIPAddressPlaceholder(privateIPAddressPlaceholder)
	}

	return add_service.NewAddServiceInstruction(
		testServiceNetwork,
		kurtosis_instruction.NewInstructionPosition(lineNumber, colNumber, fileName),
		serviceId,
		serviceConfigBuilder.Build(),
		starlark.StringDict{
			"service_id": starlark.String(serviceId),
			"config":     serviceConfigStruct,
		},
	)
}

func validateScriptOutputFromPrintInstructions(t *testing.T, instructions []kurtosis_instruction.KurtosisInstruction, expectedOutput string) {
	scriptOutput := strings.Builder{}
	for _, instruction := range instructions {
		switch instruction.(type) {
		case *kurtosis_print.PrintInstruction:
			instructionOutput, err := instruction.Execute(context.Background())
			require.Nil(t, err, "Error running the print statements")
			if instructionOutput != nil {
				scriptOutput.WriteString(*instructionOutput)
			}
		}
	}
	require.Equal(t, expectedOutput, scriptOutput.String())
}
