package startosis_engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/print_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/time_now_builtin"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/interpretation_time_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/kurtosis_print"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/remove_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_service_files"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/mock_package_content_provider"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"strings"
	"testing"
)

var (
	testServiceIpAddress = net.ParseIP("127.0.0.1")
)

var (
	emptyEnclaveComponents    = enclave_structure.NewEnclaveComponents()
	emptyInstructionsPlanMask = resolver.NewInstructionsPlanMask(0)
)

const (
	testServiceName        = service.ServiceName("example-datastore-server")
	testContainerImageName = "kurtosistech/example-datastore-server"
	testArtifactName       = "test-artifact"

	useDefaultMainFunctionName = ""

	mockEnclaveUuid      = "enclave-uuid"
	serviceUuidSuffix    = "uuid"
	mockFileArtifactName = "mock-artifact-id"

	defaultImageDownloadMode = image_download_mode.ImageDownloadMode_Missing
)

type StartosisInterpreterTestSuite struct {
	suite.Suite
	serviceNetwork               *service_network.MockServiceNetwork
	packageContentProvider       *mock_package_content_provider.MockPackageContentProvider
	runtimeValueStore            *runtime_value_store.RuntimeValueStore
	interpretationTimeValueStore *interpretation_time_value_store.InterpretationTimeValueStore

	interpreter *StartosisInterpreter
}

func (suite *StartosisInterpreterTestSuite) SetupTest() {
	suite.packageContentProvider = mock_package_content_provider.NewMockPackageContentProvider()
	enclaveDb := getEnclaveDBForTest(suite.T())

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(suite.T(), err)
	suite.runtimeValueStore = runtimeValueStore
	suite.serviceNetwork = service_network.NewMockServiceNetwork(suite.T())

	interpretationTimeValueStore, err := interpretation_time_value_store.CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(suite.T(), err)
	suite.interpretationTimeValueStore = interpretationTimeValueStore
	require.NotNil(suite.T(), interpretationTimeValueStore)

	suite.interpreter = NewStartosisInterpreter(suite.serviceNetwork, suite.packageContentProvider, suite.runtimeValueStore, nil, "", suite.interpretationTimeValueStore)

	service.NewServiceRegistration(
		testServiceName,
		service.ServiceUUID(fmt.Sprintf("%s-%s", testServiceName, serviceUuidSuffix)),
		mockEnclaveUuid,
		testServiceIpAddress,
		string(testServiceName),
	)
	suite.serviceNetwork.EXPECT().GetUniqueNameForFileArtifact().Maybe().Return(mockFileArtifactName, nil)
	suite.serviceNetwork.EXPECT().GetEnclaveUuid().Maybe().Return(enclave.EnclaveUUID(mockEnclaveUuid))
	suite.serviceNetwork.EXPECT().ExistServiceRegistration(testServiceName).Maybe().Return(true, nil)
}

func TestRunStartosisInterpreterTestSuite(t *testing.T) {
	suite.Run(t, new(StartosisInterpreterTestSuite))
}

func (suite *StartosisInterpreterTestSuite) TearDownTest() {
	suite.packageContentProvider.RemoveAll()
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_SimplePrintScript() {
	testString := "Hello World!"
	script := `
def run(plan):
	plan.print("` + testString + `")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size()) // Only the print statement

	expectedOutput := testString + `
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RandomMainFunctionAndParamsWithPlan() {
	script := `
def deploy_contract(plan,service_name,contract_name,init_message,args):
	plan.print("Service name: service_name")
	plan.print("Contract name: contract_name")
	plan.print("Init message: init_message")

	return args["arg1"] + ":" + args["arg2"]
`
	mainFunctionName := "deploy_contract"
	inputArgs := `{"service_name": "my-service", "contract_name": "my-contract", "init_message": "Init message", "args": {"arg1": "arg1-value", "arg2": "arg2-value"}}`

	result, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, mainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, inputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 3, instructionsPlan.Size()) // The three print functions
	require.NotNil(suite.T(), result)
	expectedResult := "\"arg1-value:arg2-value\""
	require.Equal(suite.T(), expectedResult, result)
	expectedOutput := "Service name: service_name\nContract name: contract_name\nInit message: init_message\n"
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RandomMainFunctionAndParams() {
	script := `
def my_func(my_arg1, my_arg2, args):
	
	all_arg_values = my_arg1 + "--" + my_arg2 +  "--" + args["arg1"] + ":" + args["arg2"]

	return all_arg_values
`
	mainFunctionName := "my_func"
	inputArgs := `{"my_arg1": "foo", "my_arg2": "bar", "args": {"arg1": "arg1-value", "arg2": "arg2-value"}}`

	result, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, mainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, inputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 0, instructionsPlan.Size()) // There are no instructions to execute
	require.NotNil(suite.T(), result)
	expectedResult := "\"foo--bar--arg1-value:arg2-value\""
	require.Equal(suite.T(), expectedResult, result)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RandomMainFunctionAndParamsErrDueToDeprecatedArgsObject() {
	script := `
def run(plan, args):
	all_arg_values = args["arg1"] + ":" + args["arg2"]
	return all_arg_values
`
	mainFunctionName := "run"
	inputArgs := `{"args": {"arg1": "arg1-value", "arg2": "arg2-value"}}`

	_, _, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, mainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, inputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Equal(suite.T(), "Evaluation error: key \"arg1\" not in dict\n\tat [3:23]: run", interpretationError.GetErrorMessage())
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_Test() {
	script := `
def run(plan):
	my_dict = {}
	plan.print(my_dict)
	my_dict["hello"] = "world"
	plan.print(my_dict)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	expectedOutput := `{}
{"hello": "world"}
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ScriptFailingSingleError() {
	script := `
def run(plan):
	plan.print("Starting Startosis script!")

unknownInstruction()
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ScriptFailingMultipleErrors() {
	script := `
def run(plan):
	plan.print("Starting Startosis script!")
unknownInstruction()
unknownVariable
unknownInstruction2()
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 1)),
		},
		multipleInterpretationErrorMsg,
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ScriptFailingSyntaxError() {
	script := `
def run():
	plan.print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 5)),
		},
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithInstruction() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 5, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 15, 38)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The grpc port is 1323
The grpc transport protocol is TCP
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithApplicationProtocol() {
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
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, fmt.Sprintf(script, testServiceName), startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 6, instructionsPlan.Size())

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The port is 1323
The transport protocol is TCP
The application protocol is http
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithInstructionMissingContainerName() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)

	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		errors.New("ServiceConfig: missing argument for image"),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 8, 24)),
			*startosis_errors.NewCallFrame("ServiceConfig", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'ServiceConfig' from the provided arguments.",
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithInstructionTypoInProtocol() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		startosis_errors.NewInterpretationError(`The following argument(s) could not be parsed or did not pass validation: {"transport_protocol":"Invalid argument value for 'transport_protocol': 'TCPK'. Valid values are TCP, SCTP, UDP"}`),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'PortSpec' from the provided arguments.",
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithInstructionPortNumberAsString() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	expectedError := startosis_errors.NewInterpretationErrorWithCauseAndCustomMsg(
		startosis_errors.NewInterpretationError(`The following argument(s) could not be parsed or did not pass validation: {"number":"Value for 'number' was expected to be an integer between 1 and 65535, but it was 'starlark.String'"}`),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("run", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 11, 20)),
			*startosis_errors.NewCallFrame("PortSpec", startosis_errors.NewScriptPosition("<builtin>", 0, 0)),
		},
		"Evaluation error: Cannot construct 'PortSpec' from the provided arguments.",
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidScriptWithMultipleInstructions() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 8, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 4, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 6, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 19, 19)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_LoadStatementIsDisallowedInKurtosis() {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
load("` + barModulePath + `", "a")
def run(plan):
	plan.print("Hello " + a)
`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 1)),
		},
		"Evaluation error: cannot load github.com/foo/bar/lib.star: 'load(\"path/to/file.star\", var_in_file=\"var_in_file\")' statement is not available in Kurtosis. Please use instead `module = import(\"path/to/file.star\")` and then `module.var_in_file`",
	).ToAPIType()

	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_SimpleImport() {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("Hello " + my_module.a)
`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size()) // Only the print statement

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_TransitiveLoading() {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/doo/crib.star"
	seedModules[moduleBar] = `a="World!"`
	moduleDooWhichLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleDooWhichLoadsModuleBar] = `module_bar = import_module("` + moduleBar + `")
b = "Hello " + module_bar.a
`
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
module_doo = import_module("` + moduleDooWhichLoadsModuleBar + `")
def run(plan):
	plan.print(module_doo.b)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_FailsOnCycle() {
	seedModules := make(map[string]string)
	moduleBarLoadsModuleDoo := "github.com/foo/bar/lib.star"
	moduleDooLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleBarLoadsModuleDoo] = `module_doo = import_module("` + moduleDooLoadsModuleBar + `")
a = "Hello" + module_doo.b`
	seedModules[moduleDooLoadsModuleBar] = `module_bar = import_module("` + moduleBarLoadsModuleDoo + `")
b = "Hello " + module_bar.a
`
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `module_doo = import_module("` + moduleDooLoadsModuleBar + `")
def run(plan):
	plan.print(module_doo.b)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleBarLoadsModuleDoo, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(moduleDooLoadsModuleBar, 1, 27)),
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 1, 27)),
		},
		"Evaluation error: There's a cycle in the import_module calls",
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_FailsOnNonExistentModule() {
	nonExistentModule := "github.com/non/existent/module.star"
	script := `
my_module = import_module("` + nonExistentModule + `")
def run(plan):
	plan.print(my_module.b)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)

	errorMsg := `Evaluation error: An error occurred while loading the module '` + nonExistentModule + `'
	Caused by: Package '` + nonExistentModule + `' not found`
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(startosis_constants.PackageIdPlaceholderForStandaloneScript, 2, 26)),
		},
		errorMsg,
	).ToAPIType()
	require.Equal(suite.T(), expectedError, interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ImportingAValidModuleThatPreviouslyFailedToLoadSucceeds() {
	barModulePath := "github.com/foo/bar/lib.star"
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("Hello " + my_module.a)
`

	// assert that first load fails
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Nil(suite.T(), instructionsPlan)

	barModuleContents := "a=\"World!\""
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(barModulePath, barModuleContents))
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	_, instructionsPlan, interpretationError = suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleScriptWithImportedStruct() {
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
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
module_bar = import_module("` + moduleBar + `")
def run(plan):
	plan.print("Starting Startosis script!")
	plan.print("Adding service " + module_bar.service_name)
	plan.add_service(name = module_bar.service_name, config = module_bar.config)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 6, 18)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidScriptWithFunctionsImportedFromOtherModule() {
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
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
datastore_module = import_module("` + moduleBar + `")

def run(plan):
	plan.print("Starting Startosis script!")
	datastore_module.deploy_datastore_services(plan)
	plan.print("Done!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 8, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 4, add_service.AddServiceBuiltinName, moduleBar, 18, 25)
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 6, add_service.AddServiceBuiltinName, moduleBar, 18, 25)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ImportModuleWithNoGlobalVariables() {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "",
	}
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
	script := `
my_module = import_module("` + barModulePath + `")
def run(plan):
	plan.print("World!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	expectedOutput := `World!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_TestInstructionQueueAndOutputBufferDontHaveDupesInterpretingAnotherScript() {
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
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seedModules))
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, scriptA, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 4, instructionsPlan.Size())
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, moduleBar, 12, 18)
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutputFromScriptA)

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

	_, instructionsPlan, interpretationError = suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, scriptB, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 3, instructionsPlan.Size())
	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, add_service.AddServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 14, 18)
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutputFromScriptB)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ReadFileFromGithub() {
	src := "github.com/foo/bar/static_files/main.txt"
	seed := map[string]string{
		src: "this is a test string",
	}
	require.Nil(suite.T(), suite.packageContentProvider.BulkAddFileContent(seed))
	script := `
def run(plan):
	plan.print("Reading file from GitHub!")
	file_contents=read_file("` + src + `")
	plan.print(file_contents)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 2, instructionsPlan.Size())

	expectedOutput := `Reading file from GitHub!
this is a test string
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RenderTemplates() {
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

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 1, render_templates.RenderTemplatesBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 20, 39)

	expectedOutput := fmt.Sprintf(`Rendering template to disk!
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ThreeLevelNestedInstructionPositionTest() {
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

	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(storeFileDefinitionPath, storeFileContent))
	require.Nil(suite.T(), suite.packageContentProvider.AddFileContent(moduleThatCallsStoreFile, moduleThatCallsStoreFileContent))

	script := `
def run(plan):
	call_store_for_me_module = import_module("github.com/kurtosis/foo.star")
	uuid = call_store_for_me_module.call_store_for_me(plan)
	plan.print(uuid)
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 4, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 2, store_service_files.StoreServiceFilesBuiltinName, storeFileDefinitionPath, 4, 40)

	expectedOutput := fmt.Sprintf(`In the module that calls store.star
In the store files instruction
%v
`, testArtifactName)
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_ValidSimpleRemoveService() {
	script := `
def run(plan):
	plan.print("Starting Startosis script!")
	service_name = "example-datastore-server"
	plan.remove_service(name=service_name)
	plan.print("The service example-datastore-server has been removed")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 3, instructionsPlan.Size())

	assertInstructionTypeAndPosition(suite.T(), instructionsPlan, 1, remove_service.RemoveServiceBuiltinName, startosis_constants.PackageIdPlaceholderForStandaloneScript, 5, 21)

	expectedOutput := `Starting Startosis script!
The service example-datastore-server has been removed
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_NoPanicIfUploadIsPassedAPathNotOnDisk() {
	filePath := "github.com/kurtosis/module/lib/lib.star"
	script := `
def run(plan):
	plan.upload_files("` + filePath + `")
`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithoutArgsAndNoArgsPassed() {
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	expectedOutput := `Hello World!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithoutArgsAndArgsPassed() {
	script := `
def run(plan):
	plan.print("Hello World!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"number": 4}`, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithArgsAndArgsPassed() {
	script := `
def run(plan, args):
	plan.print("My favorite number is {0}".format(args["number"]))
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"number": 4}`, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	expectedOutput := `My favorite number is 4
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithArgsAndNoArgsPassed() {
	script := `
def run(plan, args):
	if hasattr(args, "number"):
		plan.print("My favorite number is {0}".format(args.number))
	else:
		plan.print("Sorry no args!")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())

	expectedOutput := `Sorry no args!
`
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithMoreThanExpectedParams() {
	script := `
def run(plan, args, invalid_arg):
	plan.print("this wouldn't interpret so the text here doesnt matter")
`

	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	expectedError := "Evaluation error: function run missing 2 arguments (args, invalid_arg)"
	require.Contains(suite.T(), interpretationError.GetErrorMessage(), expectedError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithUnpackedDictButMissingArgs() {
	script := `
def run(plan, a, b):
	plan.print("this wouldn't interpret so the text here doesnt matter")
`
	missingArgumentCount := 1
	missingArgument := "b"
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"a": "x"}`, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)

	expectedError := fmt.Sprintf("Evaluation error: function run missing %d argument (%v)", missingArgumentCount, missingArgument)
	require.Contains(suite.T(), interpretationError.GetErrorMessage(), expectedError)
	require.Nil(suite.T(), instructionsPlan)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_RunWithUnpackedDict() {
	script := `
def run(plan, a, b=1):
	plan.print("My favorite number is {0}, but my favorite letter is {1}".format(b, a))
`
	_, instructionsPlan, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, `{"a": "x"}`, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
	require.Equal(suite.T(), 1, instructionsPlan.Size())
	expectedOutput := "My favorite number is 1, but my favorite letter is x\n"
	validateScriptOutputFromPrintInstructions(suite.T(), instructionsPlan, expectedOutput)
}

func (suite *StartosisInterpreterTestSuite) TestStartosisInterpreter_PrintWithoutPlanErrorsNicely() {
	script := `
def run(plan):
	print("this doesnt matter")
`

	_, _, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Equal(suite.T(), fmt.Sprintf("Evaluation error: %v\n\tat [3:7]: run\n\tat [0:0]: print", print_builtin.UsePlanFromKurtosisInstructionError), interpretationError.GetErrorMessage())
}

func (suite *StartosisInterpreterTestSuite) TestStarlarkInterpreter_TimeNowFailsWithInterpretationErr() {
	script := `
def run(plan):
	time.now()
`

	_, _, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.NotNil(suite.T(), interpretationError)
	require.Equal(suite.T(), fmt.Sprintf("Evaluation error: %v\n\tat [3:10]: run\n\tat [0:0]: now", time_now_builtin.UseRunPythonInsteadOfTimeNowError), interpretationError.GetErrorMessage())
}

func (suite *StartosisInterpreterTestSuite) TestStarlarkInterpreter_ParseDurationContinuesToWork() {
	script := `
def run(plan):
	time.parse_duration("5s")
`

	_, _, interpretationError := suite.interpreter.Interpret(context.Background(), startosis_constants.PackageIdPlaceholderForStandaloneScript, useDefaultMainFunctionName, noPackageReplaceOptions, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, script, startosis_constants.EmptyInputArgs, defaultNonBlockingMode, emptyEnclaveComponents, emptyInstructionsPlanMask, defaultImageDownloadMode)
	require.Nil(suite.T(), interpretationError)
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
		switch instruction.GetCanonicalInstruction(isSkipped).InstructionName {
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

	canonicalInstruction := instruction.GetCanonicalInstruction(isSkipped)
	require.Equal(t, expectedInstructionName, canonicalInstruction.GetInstructionName())
	expectedPosition := binding_constructors.NewStarlarkInstructionPosition(filename, expectedLine, expectedCol)
	require.Equal(t, expectedPosition, canonicalInstruction.GetPosition())
}
