package startosis_engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"net"
	"strings"
	"testing"
)

var (
	testServiceNetwork   = service_network.NewMockServiceNetworkCustom(map[service.ServiceName]net.IP{testServiceName: testServiceIpAddress})
	testServiceIpAddress = net.ParseIP("127.0.0.1")
)

const (
	testServiceName        = service.ServiceName("example-datastore-server")
	testContainerImageName = "kurtosistech/example-datastore-server"
	testArtifactName       = "test-artifact"
)

func TestStartosisInterpreter_SimplePrintScript(t *testing.T) {
	testString := "Hello World!"
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	defer packageContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	interpreter := startosisInterpreter
	script := `
def run(plan):
	plan.print("` + testString + `")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Len(t, instructions, 1) // Only the print statement
	require.Nil(t, interpretationError)

	expectedOutput := testString + `
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_DefineFactAndWait(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=output",
		extract = {
			"input": ".query.input"
		}
	)
	response = plan.wait(get_recipe, "code", "==",  200, timeout="5m", interval="5s", service_name="web-server")
	plan.print(response["body"])
`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.NotEmpty(t, instructions)
}

func TestStartosisInterpreter_ScriptFailingSingleError(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

unknownInstruction()
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingMultipleErrors(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")
unknownInstruction()
unknownVariable
unknownInstruction2()
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingSyntaxError(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run():
	plan.print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 5)),
		},
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstruction(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	privateIPAddressPlaceholder := "MAGICAL_PLACEHOLDER_TO_REPLACE"
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

	service_name = "%v"
	plan.print("Adding service " + service_name)

	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP")
		},
		private_ip_address_placeholder = "` + privateIPAddressPlaceholder + `"
	)
	datastore_service = plan.add_service(service_name = service_name, config = config)
	plan.print("The grpc port is " + str(datastore_service.ports["grpc"].number))
	plan.print("The grpc transport protocol is " + datastore_service.ports["grpc"].transport_protocol)
	plan.print("The datastore service ip address is " + datastore_service.ip_address)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 6)

	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 15, 38)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The grpc port is 1323
The grpc transport protocol is TCP
The datastore service ip address is %v
`
	validateScriptOutputFromPrintInstructions(t, instructions, fmt.Sprintf(expectedOutput, testServiceIpAddress))
}

func TestStartosisInterpreter_ValidSimpleScriptWithApplicationProtocol(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	privateIPAddressPlaceholder := "MAGICAL_PLACEHOLDER_TO_REPLACE"
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

	service_name = "%v"
	plan.print("Adding service " + service_name)

	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP", application_protocol = "http")
		},
		private_ip_address_placeholder = "` + privateIPAddressPlaceholder + `"
	)
	datastore_service = plan.add_service(service_name = service_name, config = config)
	plan.print("The port is " + str(datastore_service.ports["grpc"].number))
	plan.print("The transport protocol is " + datastore_service.ports["grpc"].transport_protocol)
	plan.print("The application protocol is " + datastore_service.ports["grpc"].application_protocol)
`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 6)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The port is 1323
The transport protocol is TCP
The application protocol is http
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionMissingContainerName(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")
	
	service_name = "example-datastore-server"
	plan.print("Adding service " + service_name)
	
	config = ServiceConfig(
		# /!\ /!\ missing container name /!\ /!\
		ports = {
			"grpc": struct(number = 1323, transport_protocol = "TCP")
		}
	)
	plan.add_service(service_name = service_name, config = config)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)

	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		errors.New("ServiceConfig: missing argument for image"),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 8, 24)),
			*startosis_errors.NewCallFrame("ServiceConfig", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'ServiceConfig' from the provided arguments.",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionTypoInProtocol(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

	service_name = "example-datastore-server"
	plan.print("Adding service " + service_name)

	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCPK") # typo in protocol
		}
	)
	plan.add_service(service_name = service_name, config = config)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Port protocol should be one of TCP, SCTP, UDP",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionPortNumberAsString(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

	service_name = "example-datastore-server"
	plan.print("Adding service " + service_name)

	config = ServiceConfig(
		image = "` + testContainerImageName + `",
		ports = {
			"grpc": PortSpec(number = "1234", protocol = "TCP") # port number should be an int
		}
	)
	plan.add_service(service_name = service_name, config = config)
`
	expectedErrorStr := `Evaluation error: Cannot construct a PortSpec from the provided arguments. Error was: 
PortSpec: for parameter "number": got string, want int`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		expectedErrorStr,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidScriptWithMultipleInstructions(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
service_name = "example-datastore-server"
ports = [1323, 1324, 1325]	

def deploy_datastore_services(plan):
	for i in range(len(ports)):
		unique_service_name = service_name + "-" + str(i)
		plan.print("Adding service " + unique_service_name)
		config = ServiceConfig(
			image = "` + testContainerImageName + `",
			ports = {
				"grpc": PortSpec(
					number = ports[i],
					transport_protocol = "TCP"
				)
			}
		)

		plan.add_service(service_name = unique_service_name, config = config)

def run(plan):
	plan.print("Starting Startosis script!")
	deploy_datastore_services(plan)
	plan.print("Done!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 8)

	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(t, instructions[4], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(t, instructions[6], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_LoadStatementIsDisallowedInKurtosis(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
load("` + barModulePath + `", "a")
def run(plan):
	plan.print("Hello " + a)
`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 1)),
		},
		"Evaluation error: cannot load github.com/foo/bar/lib.star: 'load(\"path/to/file.star\", var_in_file=\"var_in_file\")' statement is not available in Kurtosis. Please use instead `module = import(\"path/to/file.star\")` and then `module.var_in_file`",
	).ToAPIType()

	require.Equal(t, expectedError, interpretationError)
	require.Empty(t, instructions)
}

func TestStartosisInterpreter_SimpleImport(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("Hello " + my_module.a)
`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Len(t, instructions, 1) // Only the print statement
	require.Nil(t, interpretationError)

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
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
module_doo = import_module("` + moduleDooWhichLoadsModuleBar + `")
def run(plan):
	plan.print(module_doo.b)

`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
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
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `module_doo = import_module("` + moduleDooLoadsModuleBar + `")
def run(plan):
	plan.print(module_doo.b)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions) // No kurtosis instruction
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleBarLoadsModuleDoo, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleDooLoadsModuleBar, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 1, 27)),
		},
		"Evaluation error: There's a cycle in the import_module calls",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_FailsOnNonExistentModule(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	nonExistentModule := "github.com/non/existent/module.star"
	script := `
my_module = import_module("` + nonExistentModule + `")
def run(plan):
	plan.print(my_module.b)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Empty(t, instructions) // No kurtosis instruction

	errorMsg := `Evaluation error: An error occurred while loading the module '` + nonExistentModule + `'
	Caused by: Package '` + nonExistentModule + `' not found`
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 26)),
		},
		errorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_RequestInstruction(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtime_value_store.NewRuntimeValueStore())
	interpreter := startosisInterpreter
	script := `
def run(plan):
	get_recipe = GetHttpRequestRecipe(
		port_id = "http-port",
		endpoint = "?input=output",
		extract = {
			"input": ".query.input"
		}
	)
	response = plan.request(get_recipe, service_name = "web-server")
	plan.print(response["code"])`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 2)
}

func TestStartosisInterpreter_ImportingAValidModuleThatPreviouslyFailedToLoadSucceeds(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	barModulePath := "github.com/foo/bar/lib.star"
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("Hello " + my_module.a)
`

	// assert that first load fails
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, instructions)
	require.NotNil(t, interpretationError)

	barModuleContents := "a=\"World!\""
	require.Nil(t, packageContentProvider.AddFileContent(barModulePath, barModuleContents))
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	_, instructions, interpretationError = interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1) // The print statement
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleScriptWithImportedStruct(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_name = "example-datastore-server"
config = ServiceConfig(
	image = "kurtosistech/example-datastore-server",
	ports = {
		"grpc": PortSpec(number = 1323, transport_protocol = "TCP")
	}
)
`
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
module_bar = import_module("` + moduleBar + `")
def run(plan):
	plan.print("Starting Startosis script!")
	plan.print("Adding service " + module_bar.service_name)
	plan.add_service(service_name = module_bar.service_name, config = module_bar.config)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Len(t, instructions, 3)
	require.Nil(t, interpretationError)

	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 18)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidScriptWithFunctionsImportedFromOtherModule(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_name = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services(plan):
    for i in range(len(ports)):
        unique_service_name = service_name + "-" + str(i)
        plan.print("Adding service " + unique_service_name)
        config = ServiceConfig(
			image = "kurtosistech/example-datastore-server",
			ports = {
				"grpc": PortSpec(
					number = ports[i],
					transport_protocol = "TCP"
				)
			}
		)
        plan.add_service(service_name = unique_service_name, config = config)
`
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
datastore_module = import_module("` + moduleBar + `")

def run(plan):
	plan.print("Starting Startosis script!")
	datastore_module.deploy_datastore_services(plan)
	plan.print("Done!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 8)

	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(t, instructions[4], add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(t, instructions[6], add_service.AddServiceBuiltinName, moduleBar, 18, 25)

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
		barModulePath: "",
	}
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("World!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Len(t, instructions, 1)
	require.Nil(t, interpretationError)

	expectedOutput := `World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_TestInstructionQueueAndOutputBufferDontHaveDupesInterpretingAnotherScript(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
def deploy_service(plan):
	service_name = "example-datastore-server"
	plan.print("Constructing config")
	config = ServiceConfig(
		image = "kurtosistech/example-datastore-server",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP")
		}
	)
	plan.print("Adding service " + service_name)
	plan.add_service(service_name = service_name, config = config)
`
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seedModules))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	scriptA := `
deployer = import_module("` + moduleBar + `")
def run(plan):
	deployer.deploy_service(plan)
	plan.print("Starting Startosis script!")
`

	expectedOutputFromScriptA := `Constructing config
Adding service example-datastore-server
Starting Startosis script!
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, scriptA, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 4)
	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, moduleBar, 12, 18)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutputFromScriptA)

	scriptB := `
def run(plan):
	plan.print("Starting Startosis script!")

	service_name = "example-datastore-server"
	plan.print("Adding service " + service_name)
	
	config = ServiceConfig(
		image = "kurtosistech/example-datastore-server",
		ports = {
			"grpc": PortSpec(number = 1323, transport_protocol = "TCP")
		}
	)
	plan.add_service(service_name = service_name, config = config)
`
	expectedOutputFromScriptB := `Starting Startosis script!
Adding service example-datastore-server
`

	_, instructions, interpretationError = interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, scriptB, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 3)
	assertInstructionTypeAndPosition(t, instructions[2], add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 14, 18)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutputFromScriptB)
}

func TestStartosisInterpreter_ValidExecScript(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	testRuntimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, testRuntimeValueStore)
	script := `
def run(plan):
	plan.print("Executing mkdir!")
	recipe = ExecRecipe(
		command = ["mkdir", "/tmp/foo"]
	)
	plan.exec(recipe = recipe, service_name = "web-server")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 2)
}

func TestStartosisInterpreter_InvalidExecRecipeMissingRequiredCommand(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	testRuntimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, testRuntimeValueStore)
	script := `
def run(plan):
	plan.print("Executing mkdir!")
	recipe = ExecRecipe()
	plan.exec(recipe = recipe)
`

	_, _, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 4, 21)),
			*startosis_errors.NewCallFrame("ExecRecipe", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 0, 0)),
		},
		"Evaluation error: ExecRecipe: missing argument for command",
	).ToAPIType()

	require.NotNil(t, interpretationError)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ReadFileFromGithub(t *testing.T) {
	src := "github.com/foo/bar/static_files/main.txt"
	seed := map[string]string{
		src: "this is a test string",
	}
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	require.Nil(t, packageContentProvider.BulkAddFileContent(seed))
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Reading file from GitHub!")
	file_contents=read_file("` + src + `")
	plan.print(file_contents)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 2)

	expectedOutput := `Reading file from GitHub!
this is a test string
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_RenderTemplates(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
template_data = {
			"Name" : "Stranger",
			"Answer": 6,
			"Numbers": [1, 2, 3],
			"UnixTimeStamp": 1257894000,
			"LargeFloat": 1231231243.43,
			"Alive": True
}

data = {
	"/foo/bar/test.txt" : struct(
		template="Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}",
		data=template_data
	)
}

def run(plan):
	plan.print("Rendering template to disk!")
	artifact_name = plan.render_templates(config = data, name = "` + testArtifactName + `")
	plan.print(artifact_name)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 3)

	assertInstructionTypeAndPosition(t, instructions[1], render_templates.RenderTemplatesBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 20, 39)

	expectedOutput := fmt.Sprintf(`Rendering template to disk!
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ThreeLevelNestedInstructionPositionTest(t *testing.T) {
	storeFileDefinitionPath := "github.com/kurtosis/store.star"
	storeFileContent := `
def store_for_me(plan):
	plan.print("In the store files instruction")
	artifact_name=plan.store_service_files(service_name="example-datastore-server", src="/foo/bar", name = "` + string(testArtifactName) + `")
	return artifact_name
`

	moduleThatCallsStoreFile := "github.com/kurtosis/foo.star"
	moduleThatCallsStoreFileContent := `
store_for_me_module = import_module("github.com/kurtosis/store.star")
def call_store_for_me(plan):
	plan.print("In the module that calls store.star")
	return store_for_me_module.store_for_me(plan)
	`

	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	err := packageContentProvider.AddFileContent(storeFileDefinitionPath, storeFileContent)
	require.Nil(t, err)

	err = packageContentProvider.AddFileContent(moduleThatCallsStoreFile, moduleThatCallsStoreFileContent)
	require.Nil(t, err)

	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	call_store_for_me_module = import_module("github.com/kurtosis/foo.star")
	uuid = call_store_for_me_module.call_store_for_me(plan)
	plan.print(uuid)
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 4)

	assertInstructionTypeAndPosition(t, instructions[2], store_service_files.StoreServiceFilesBuiltinName, storeFileDefinitionPath, 4, 40)

	expectedOutput := fmt.Sprintf(`In the module that calls store.star
In the store files instruction
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_ValidSimpleRemoveService(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Starting Startosis script!")
	service_name = "example-datastore-server"
	plan.remove_service(service_name=service_name)
	plan.print("The service example-datastore-server has been removed")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Len(t, instructions, 3)
	require.Nil(t, interpretationError)

	assertInstructionTypeAndPosition(t, instructions[1], remove_service.RemoveServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 21)

	expectedOutput := `Starting Startosis script!
The service example-datastore-server has been removed
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_NoPanicIfUploadIsPassedAPathNotOnDisk(t *testing.T) {
	filePath := "github.com/kurtosis/module/lib/lib.star"
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.upload_files("` + filePath + `")
`
	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, instructions)
	require.NotNil(t, interpretationError)
}

func TestStartosisInterpreter_RunWithoutArgsNoArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1)

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_RunWithoutArgsArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, `{"number": 4}`)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1)

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_RunWithArgsArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("My favorite number is {0}".format(args.number))
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, `{"number": 4}`)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1)

	expectedOutput := `My favorite number is 4
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_RunWithArgsNoArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	if hasattr(args, "number"):
		plan.print("My favorite number is {0}".format(args.number))
	else:
		plan.print("Sorry no args!")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.Nil(t, interpretationError)
	require.Len(t, instructions, 1)

	expectedOutput := `Sorry no args!
`
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_RunWithMoreThanExpectedParams(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args, invalid_arg):
	plan.print("this wouldn't interpret so the text here doesnt matter")
`

	_, instructions, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.NotNil(t, interpretationError)
	expectedError := fmt.Sprintf("The 'run' entrypoint function can have at most '%v' argument got '%v'", maximumParamsAllowedForRunFunction, 3)
	require.Equal(t, expectedError, interpretationError.GetErrorMessage())

	expectedOutput := ``
	validateScriptOutputFromPrintInstructions(t, instructions, expectedOutput)
}

func TestStartosisInterpreter_PrintWithoutPlanErrorsNicely(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	print("this doesnt matter")
`

	_, _, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, script, startosis_constants.EmptyInputArgs)
	require.NotNil(t, interpretationError)
	require.Equal(t, fmt.Sprintf("Evaluation error: %v\n\tat [3:7]: run\n\tat [0:0]: print", print_builtin.UsePlanFromKurtosisInstructionError), interpretationError.GetErrorMessage())
}

// #####################################################################################################################
//
//	TEST HELPERS
//
// #####################################################################################################################
func validateScriptOutputFromPrintInstructions(t *testing.T, instructions []kurtosis_instruction.KurtosisInstruction, expectedOutput string) {
	scriptOutput := strings.Builder{}
	for _, instruction := range instructions {
		switch instruction.(type) {
		case *kurtosis_print.PrintInstruction:
			instructionOutput, err := instruction.Execute(context.Background())
			require.Nil(t, err, "Error running the print statements")
			if instructionOutput != nil {
				scriptOutput.WriteString(*instructionOutput)
				scriptOutput.WriteString("\n")
			}
		}
	}
	require.Equal(t, expectedOutput, scriptOutput.String())
}

func assertInstructionTypeAndPosition(t *testing.T, instruction kurtosis_instruction.KurtosisInstruction, expectedInstructionName string, filename string, expectedLine int32, expectedCol int32) {
	canonicalInstruction := instruction.GetCanonicalInstruction()
	require.Equal(t, expectedInstructionName, canonicalInstruction.GetInstructionName())
	expectedPosition := binding_constructors.NewStarlarkInstructionPosition(filename, expectedLine, expectedCol)
	require.Equal(t, expectedPosition, canonicalInstruction.GetPosition())
}
