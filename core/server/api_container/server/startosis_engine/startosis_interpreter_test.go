package startosis_engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
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

var (
	emptyInstructionsPlanMask = resolver.NewInstructionsPlanMask(0)
)

const (
	testServiceName        = service.ServiceName("example-datastore-server")
	testContainerImageName = "kurtosistech/example-datastore-server"
	testArtifactName       = "test-artifact"

	useDefaultMainFunctionName = ""
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size()) // Only the print statement

	expectedOutput := testString + `
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
}

func TestStartosisInterpreter_RandomMainFunctionAndParamsWithPlan(t *testing.T) {

	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	defer packageContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	interpreter := startosisInterpreter
	script := `
def deploy_contract(plan,service_name,contract_name,init_message,args):
	plan.print("Service name: service_name")
	plan.print("Contract name: contract_name")
	plan.print("Init message: init_message")

	return args["arg1"] + ":" + args["arg2"]
`
	mainFunctionName := "deploy_contract"
	inputArgs := `{"service_name": "my-service", "contract_name": "my-contract", "init_message": "Init message", "args": {"arg1": "arg1-value", "arg2": "arg2-value"}}`

	result, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, mainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, inputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 3, instructionsPlan.Size()) // The three print functions
	require.NotNil(t, result)
	expectedResult := "\"arg1-value:arg2-value\""
	require.Equal(t, expectedResult, result)
	expectedOutput := "Service name: service_name\nContract name: contract_name\nInit message: init_message\n"
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
}

func TestStartosisInterpreter_RandomMainFunctionAndParams(t *testing.T) {

	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	defer packageContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	interpreter := startosisInterpreter
	script := `
def my_func(my_arg1, my_arg2, args):
	
	all_arg_values = my_arg1 + "--" + my_arg2 +  "--" + args["arg1"] + ":" + args["arg2"]

	return all_arg_values
`
	mainFunctionName := "my_func"
	inputArgs := `{"my_arg1": "foo", "my_arg2": "bar", "args": {"arg1": "arg1-value", "arg2": "arg2-value"}}`

	result, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, mainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, inputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 0, instructionsPlan.Size()) // There are no instructions to execute
	require.NotNil(t, result)
	expectedResult := "\"foo--bar--arg1-value:arg2-value\""
	require.Equal(t, expectedResult, result)
}

func TestStartosisInterpreter_Test(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	defer packageContentProvider.RemoveAll()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	interpreter := startosisInterpreter
	script := `
def run(plan):
	my_dict = {}
	plan.print(my_dict)
	my_dict["hello"] = "world"
	plan.print(my_dict)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 2, instructionsPlan.Size())

	expectedOutput := `{}
{"hello": "world"}
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 5)),
		},
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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
	datastore_service = plan.add_service(name = service_name, config = config)
	plan.print("The grpc port is " + str(datastore_service.ports["grpc"].number))
	plan.print("The grpc transport protocol is " + datastore_service.ports["grpc"].transport_protocol)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 5, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 15, 38)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The grpc port is 1323
The grpc transport protocol is TCP
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	datastore_service = plan.add_service(name = service_name, config = config)
	plan.print("The port is " + str(datastore_service.ports["grpc"].number))
	plan.print("The transport protocol is " + datastore_service.ports["grpc"].transport_protocol)
	plan.print("The application protocol is " + datastore_service.ports["grpc"].application_protocol)
`
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 6, instructionsPlan.Size())

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The port is 1323
The transport protocol is TCP
The application protocol is http
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)

	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		errors.New("ServiceConfig: missing argument for image"),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 8, 24)),
			*startosis_errors.NewCallFrame("ServiceConfig", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'ServiceConfig' from the provided arguments.",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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
	plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		startosis_errors.NewInterpretationError(`The following argument(s) could not be parsed or did not pass validation: {"transport_protocol":"Invalid argument value for 'transport_protocol': 'TCPK'. Valid values are TCP, SCTP, UDP"}`),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'PortSpec' from the provided arguments.",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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
			"grpc": PortSpec(number = "1234", transport_protocol = "TCP") # port number should be an int
		}
	)
	plan.add_service(name = service_name, config = config)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		startosis_errors.NewInterpretationError(`The following argument(s) could not be parsed or did not pass validation: {"number":"Value for 'number' was expected to be an integer between 1 and 65535, but it was 'starlark.String'"}`),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'PortSpec' from the provided arguments.",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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

		plan.add_service(name = unique_service_name, config = config)

def run(plan):
	plan.print("Starting Startosis script!")
	deploy_datastore_services(plan)
	plan.print("Done!")
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 8, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(t, instructionsPlan, 4, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(t, instructionsPlan, 6, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 1)),
		},
		"Evaluation error: cannot load github.com/foo/bar/lib.star: 'load(\"path/to/file.star\", var_in_file=\"var_in_file\")' statement is not available in Kurtosis. Please use instead `module = import(\"path/to/file.star\")` and then `module.var_in_file`",
	).ToAPIType()

	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size()) // Only the print statement

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
}

func TestStartosisInterpreter_TransitiveLoading(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/doo/crib.star"
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleBarLoadsModuleDoo, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleDooLoadsModuleBar, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 1, 27)),
		},
		"Evaluation error: There's a cycle in the import_module calls",
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)

	errorMsg := `Evaluation error: An error occurred while loading the module '` + nonExistentModule + `'
	Caused by: Package '` + nonExistentModule + `' not found`
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 26)),
		},
		errorMsg,
	).ToAPIType()
	require.Equal(t, expectedError, interpretationError)
	require.Nil(t, instructionsPlan)
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
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)
	require.Nil(t, instructionsPlan)

	barModuleContents := "a=\"World!\""
	require.Nil(t, packageContentProvider.AddFileContent(barModulePath, barModuleContents))
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	_, instructionsPlan, interpretationError = interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	plan.add_service(name = module_bar.service_name, config = module_bar.config)
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 18)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
        plan.add_service(name = unique_service_name, config = config)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 8, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(t, instructionsPlan, 4, add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(t, instructionsPlan, 6, add_service.AddServiceBuiltinName, moduleBar, 18, 25)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())

	expectedOutput := `World!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	plan.add_service(name = service_name, config = config)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, scriptA, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 4, instructionsPlan.Size())
	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, moduleBar, 12, 18)
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutputFromScriptA)

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
	plan.add_service(name = service_name, config = config)
`
	expectedOutputFromScriptB := `Starting Startosis script!
Adding service example-datastore-server
`

	_, instructionsPlan, interpretationError = interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, scriptB, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 3, instructionsPlan.Size())
	assertInstructionTypeAndPosition(t, instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 14, 18)
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutputFromScriptB)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 2, instructionsPlan.Size())

	expectedOutput := `Reading file from GitHub!
this is a test string
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 1, render_templates.RenderTemplatesBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 20, 39)

	expectedOutput := fmt.Sprintf(`Rendering template to disk!
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 4, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 2, store_service_files.StoreServiceFilesBuiltinName, storeFileDefinitionPath, 4, 40)

	expectedOutput := fmt.Sprintf(`In the module that calls store.star
In the store files instruction
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	plan.remove_service(name=service_name)
	plan.print("The service example-datastore-server has been removed")
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(t, instructionsPlan, 1, remove_service.RemoveServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 21)

	expectedOutput := `Starting Startosis script!
The service example-datastore-server has been removed
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)
	require.Nil(t, instructionsPlan)
}

func TestStartosisInterpreter_RunWithoutArgsAndNoArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
}

func TestStartosisInterpreter_RunWithoutArgsAndArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"number": 4}`, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)
	require.Nil(t, instructionsPlan)
}

func TestStartosisInterpreter_RunWithArgsAndArgsPassed(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, args):
	plan.print("My favorite number is {0}".format(args["number"]))
`

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"number": 4}`, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())

	expectedOutput := `My favorite number is 4
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
}

func TestStartosisInterpreter_RunWithArgsAndNoArgsPassed(t *testing.T) {
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())

	expectedOutput := `Sorry no args!
`
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)
	expectedError := "Evaluation error: function run missing 2 arguments (args, invalid_arg)"
	require.Contains(t, interpretationError.GetErrorMessage(), expectedError)
	require.Nil(t, instructionsPlan)
}

func TestStartosisInterpreter_RunWithUnpackedDictButMissingArgs(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, a, b):
	plan.print("this wouldn't interpret so the text here doesnt matter")
`
	missingArgumentCount := 1
	missingArgument := "b"
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"a": "x"}`, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)

	expectedError := fmt.Sprintf("Evaluation error: function run missing %d argument (%v)", missingArgumentCount, missingArgument)
	require.Contains(t, interpretationError.GetErrorMessage(), expectedError)
	require.Nil(t, instructionsPlan)
}

func TestStartosisInterpreter_RunWithUnpackedDict(t *testing.T) {
	packageContentProvider := mock_package_content_provider.NewMockPackageContentProvider()
	defer packageContentProvider.RemoveAll()
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	interpreter := NewStartosisInterpreter(testServiceNetwork, packageContentProvider, runtimeValueStore)
	script := `
def run(plan, a, b=1):
	plan.print("My favorite number is {0}, but my favorite letter is {1}".format(b, a))
`
	_, instructionsPlan, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"a": "x"}`, emptyInstructionsPlanMask)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, instructionsPlan.Size())
	expectedOutput := "My favorite number is 1, but my favorite letter is x\n"
	validateScriptOutputFromPrintInstructions(t, instructionsPlan, expectedOutput)
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

	_, _, interpretationError := interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, emptyInstructionsPlanMask)
	require.NotNil(t, interpretationError)
	require.Equal(t, fmt.Sprintf("Evaluation error: %v\n\tat [3:7]: run\n\tat [0:0]: print", print_builtin.UsePlanFromKurtosisInstructionError), interpretationError.GetErrorMessage())
}

// #####################################################################################################################
//
//	TEST HELPERS
//
// #####################################################################################################################
func validateScriptOutputFromPrintInstructions(t *testing.T, instructionsPlan *instructions_plan.InstructionsPlan, expectedOutput string) {
	scriptOutput := strings.Builder{}
	scheduledInstructions, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	for _, scheduledInstruction := range scheduledInstructions {
		if scheduledInstruction.IsExecuted() {
			continue
		}
		instruction := scheduledInstruction.GetInstruction()
		switch instruction.GetCanonicalInstruction().InstructionName {
		case kurtosis_print.PrintBuiltinName:
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

func assertInstructionTypeAndPosition(t *testing.T, instructionsPlan *instructions_plan.InstructionsPlan, idxInPlan int, expectedInstructionName string, filename string, expectedLine int32, expectedCol int32) {
	scheduledInstructions, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)
	instruction := scheduledInstructions[idxInPlan].GetInstruction()

	canonicalInstruction := instruction.GetCanonicalInstruction()
	require.Equal(t, expectedInstructionName, canonicalInstruction.GetInstructionName())
	expectedPosition := binding_constructors.NewStarlarkInstructionPosition(filename, expectedLine, expectedCol)
	require.Equal(t, expectedPosition, canonicalInstruction.GetPosition())
}
