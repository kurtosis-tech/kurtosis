package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/add_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/exec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/store_files_from_service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules/mock_module_content_provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var testServiceNetwork service_network.ServiceNetwork = service_network.NewEmptyMockServiceNetwork()

const testContainerImageName = "kurtosistech/example-datastore-server"

func TestStartosisInterpreter_SimplePrintScript(t *testing.T) {
	testString := "Hello World!"
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	startosisInterpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	interpreter := startosisInterpreter
	script := `
print("` + testString + `")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions)) // No kurtosis instruction
	require.Nil(t, interpretationError)

	expectedOutput := testString + `
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ScriptFailingSingleError(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

unknownInstruction()
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Multiple errors caught interpreting the Startosis script. Listing each of them below.",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingMultipleErrors(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

unknownInstruction()
print(unknownVariable)

unknownInstruction2()
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		multipleInterpretationErrorMsg,
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("undefined: unknownInstruction", startosis_errors.NewScriptPosition(4, 1)),
			*startosis_errors.NewCallFrame("undefined: unknownVariable", startosis_errors.NewScriptPosition(5, 7)),
			*startosis_errors.NewCallFrame("undefined: unknownInstruction2", startosis_errors.NewScriptPosition(7, 1)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ScriptFailingSyntaxError(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorFromStacktrace(
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("load statement must import at least 1 symbol", startosis_errors.NewScriptPosition(4, 5)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstruction(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "` + testContainerImageName + `",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
datastore_service = add_service(service_id = service_id, service_config = service_config)
print("The grpc port is " + str(datastore_service.ports["grpc"].number))
print("The grpc port protocol is " + datastore_service.ports["grpc"].protocol)
print("The datastore service ip address is " + datastore_service.ip_address)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(13, 32),
		"example-datastore-server",
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
The grpc port is 1323
The grpc port protocol is TCP
The datastore service ip address is {{kurtosis:example-datastore-server.ip_address}}
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionMissingContainerName(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	# /!\ /!\ missing container name /!\ /!\
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Missing value 'container_image_name' as element of the struct object 'service_config'",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionTypoInProtocol(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "` + testContainerImageName + `",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCPK") # typo in protocol
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Port protocol should be one of TCP, SCTP, UDP",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidSimpleScriptWithInstructionPortNumberAsString(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "` + testContainerImageName + `",
	used_ports = {
		"grpc": struct(number = "1234", protocol = "TCP") # port number should be an int
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 0, len(instructions))
	require.Empty(t, scriptOutput)
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		"Evaluation error: Argument 'number' is expected to be an integer. Got starlark.String",
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(13, 12)),
			*startosis_errors.NewCallFrame("add_service", startosis_errors.NewScriptPosition(0, 0)),
		},
	)
	require.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_ValidScriptWithMultipleInstructions(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
ports = [1323, 1324, 1325]

def deploy_datastore_services():
    for i in range(len(ports)):
        unique_service_id = service_id + "-" + str(i)
        print("Adding service " + unique_service_id)
        service_config = struct(
			container_image_name = "` + testContainerImageName + `",
			used_ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, service_config = service_config)

deploy_datastore_services()
print("Done!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 3, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction0 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-0"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)
	addServiceInstruction1 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-1"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1324,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)
	addServiceInstruction2 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(22, 26),
		service.ServiceID("example-datastore-server-2"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	require.Equal(t, instructions[0], addServiceInstruction0)
	require.Equal(t, instructions[1], addServiceInstruction1)
	require.Equal(t, instructions[2], addServiceInstruction2)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_SimpleLoading(t *testing.T) {
	barModulePath := "github.com/foo/bar/lib.star"
	seedModules := map[string]string{
		barModulePath: "a=\"World!\"",
	}
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + barModulePath + `", "a")
print("Hello " + a)
`
	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_TransitiveLoading(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `a="World!"`
	moduleDooWhichLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleDooWhichLoadsModuleBar] = `load("` + moduleBar + `", "a")
b = "Hello " + a
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + moduleDooWhichLoadsModuleBar + `", "b")
print(b)

`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError)

	expectedOutput := `Hello World!
`
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_FailsOnCycle(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBarLoadsModuleDoo := "github.com/foo/bar/lib.star"
	moduleDooLoadsModuleBar := "github.com/foo/doo/lib.star"
	seedModules[moduleBarLoadsModuleDoo] = `load("` + moduleDooLoadsModuleBar + `", "b")
a = "Hello" + b`
	seedModules[moduleDooLoadsModuleBar] = `load("` + moduleBarLoadsModuleDoo + `", "a")
b = "Hello " + a
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + moduleDooLoadsModuleBar + `", "b")
print(b)
`

	_, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		fmt.Sprintf("Evaluation error: cannot load %v: cannot load %v: cannot load %v: There is a cycle in the load graph", moduleDooLoadsModuleBar, moduleBarLoadsModuleDoo, moduleDooLoadsModuleBar),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 1)),
		},
	)
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_FailsOnNonExistentModule(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	nonExistentModule := "github.com/non/existent/module.star"
	script := `
load("` + nonExistentModule + `", "b")
print(b)
`
	_, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction

	expectedError := startosis_errors.NewInterpretationErrorWithCustomMsg(
		fmt.Sprintf("Evaluation error: cannot load %v: An error occurred while loading the module '%v'", nonExistentModule, nonExistentModule),
		[]startosis_errors.CallFrame{
			*startosis_errors.NewCallFrame("<toplevel>", startosis_errors.NewScriptPosition(2, 1)),
		},
	)
	assert.Equal(t, expectedError, interpretationError)
}

func TestStartosisInterpreter_LoadingAValidModuleThatPreviouslyFailedToLoadSucceeds(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	barModulePath := "github.com/foo/bar/lib.star"
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + barModulePath + `", "a")
print("Hello " + a)
`

	// assert that first load fails
	_, interpretationError, _ := interpreter.Interpret(context.Background(), script)
	assert.NotNil(t, interpretationError)

	barModuleContents := "a=\"World!\""
	moduleContentProvider.Add(barModulePath, barModuleContents)
	expectedOutput := `Hello World!
`
	// assert that second load succeeds
	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	assert.Nil(t, interpretationError)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidSimpleScriptWithImportedStruct(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")

print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(6, 12),
		service.ServiceID("example-datastore-server"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Constructing service_config
Starting Startosis script!
Adding service example-datastore-server
`
	require.Equal(t, expectedOutput, string(scriptOutput))
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
        service_config = struct(
			container_image_name = "kurtosistech/example-datastore-server",
			used_ports = {
				"grpc": struct(
					number = ports[i],
					protocol = "TCP"
				)
			}
		)
        add_service(service_id = unique_service_id, service_config = service_config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + moduleBar + `", "deploy_datastore_services")
print("Starting Startosis script!")

deploy_datastore_services()
print("Done!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 3, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction0 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-0"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)
	addServiceInstruction1 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-1"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1324,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)
	addServiceInstruction2 := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(5, 26),
		service.ServiceID("example-datastore-server-2"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	require.Equal(t, instructions[0], addServiceInstruction0)
	require.Equal(t, instructions[1], addServiceInstruction1)
	require.Equal(t, instructions[2], addServiceInstruction2)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_AddServiceInOtherModulePopulatesQueue(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	addServiceInstruction := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(11, 12),
		service.ServiceID("example-datastore-server"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	require.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Constructing service_config
Adding service example-datastore-server
Starting Startosis script!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_TestInstructionQueueAndOutputBufferDontHaveDupesInterpretingAnotherScript(t *testing.T) {
	seedModules := make(map[string]string)
	moduleBar := "github.com/foo/bar/lib.star"
	seedModules[moduleBar] = `
service_id = "example-datastore-server"
print("Constructing service_config")
service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
print("Adding service " + service_id)
add_service(service_id = service_id, service_config = service_config)
`
	moduleContentProvider := mock_module_content_provider.NewMockModuleContentProvider(seedModules)
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	scriptA := `
load("` + moduleBar + `", "service_id", "service_config")
print("Starting Startosis script!")
`
	addServiceInstructionFromScriptA := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(11, 12),
		service.ServiceID("example-datastore-server"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	expectedOutputFromScriptA := `Constructing service_config
Adding service example-datastore-server
Starting Startosis script!
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), scriptA)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)
	require.Equal(t, instructions[0], addServiceInstructionFromScriptA)
	require.Equal(t, expectedOutputFromScriptA, string(scriptOutput))

	scriptB := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
add_service(service_id = service_id, service_config = service_config)
`
	addServiceInstructionFromScriptB := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(13, 12),
		service.ServiceID("example-datastore-server"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)
	expectedOutputFromScriptB := `Starting Startosis script!
Adding service example-datastore-server
`

	scriptOutput, interpretationError, instructions = interpreter.Interpret(context.Background(), scriptB)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, len(instructions))
	require.Equal(t, instructions[0], addServiceInstructionFromScriptB)
	require.Equal(t, expectedOutputFromScriptB, string(scriptOutput))
}

func TestStartosisInterpreter_AddServiceWithEnvVarsCmdArgsAndEntryPointArgs(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Starting Startosis script!")
service_id = "example-datastore-server"
print("Adding service " + service_id)
store_service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCP")
	}
)
datastore_service = add_service(service_id = service_id, service_config = store_service_config)
client_service_id = "example-datastore-client"
print("Adding service " + client_service_id)
client_service_config = struct(
	container_image_name = "kurtosistech/example-datastore-client",
	used_ports = {
		"grpc": struct(number = 1337, protocol = "TCP")
	},
	entry_point_args = ["--store-port " + str(datastore_service.ports["grpc"].number), "--store-ip " + datastore_service.ip_address],
	cmd_args = ["ping", datastore_service.ip_address],
	env_vars = {"STORE_IP": datastore_service.ip_address}
)
add_service(service_id = client_service_id, service_config = client_service_config)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Nil(t, interpretationError)
	require.Equal(t, 2, len(instructions))

	dataSourceAddServiceInstruction := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(11, 32),
		service.ServiceID("example-datastore-server"),
		services.NewServiceConfigBuilder(
			testContainerImageName,
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).Build(),
	)

	clientAddServiceInstruction := add_service.NewAddServiceInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(23, 12),
		service.ServiceID("example-datastore-client"),
		services.NewServiceConfigBuilder(
			"kurtosistech/example-datastore-client",
		).WithPrivatePorts(
			map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1337,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		).WithEntryPointArgs(
			[]string{"--store-port 1323", "--store-ip {{kurtosis:example-datastore-server.ip_address}}"},
		).WithCmdArgs(
			[]string{"ping", "{{kurtosis:example-datastore-server.ip_address}}"},
		).WithEnvVars(
			map[string]string{"STORE_IP": "{{kurtosis:example-datastore-server.ip_address}}"},
		).Build(),
	)

	require.Equal(t, instructions[0], dataSourceAddServiceInstruction)
	require.Equal(t, instructions[1], clientAddServiceInstruction)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
Adding service example-datastore-client
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_ValidExecScriptWithoutExitCodeDefaultsTo0(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Executing mkdir!")
exec(service_id = "example-datastore-server", command = ["mkdir", "/tmp/foo"])
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	execInstruction := exec.NewExecInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(3, 5),
		"example-datastore-server",
		[]string{"mkdir", "/tmp/foo"},
		0,
	)

	require.Equal(t, instructions[0], execInstruction)

	expectedOutput := `Executing mkdir!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_PassedExitCodeIsInterpretedCorrectly(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Executing mkdir!")
exec(service_id = "example-datastore-server", command = ["mkdir", "/tmp/foo"], expected_exit_code = -7)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Equal(t, 1, len(instructions))
	require.Nil(t, interpretationError)

	execInstruction := exec.NewExecInstruction(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(3, 5),
		"example-datastore-server",
		[]string{"mkdir", "/tmp/foo"},
		-7,
	)

	require.Equal(t, instructions[0], execInstruction)

	expectedOutput := `Executing mkdir!
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}

func TestStartosisInterpreter_StoreFileFromService(t *testing.T) {
	moduleContentProvider := mock_module_content_provider.NewEmptyMockModuleContentProvider()
	interpreter := NewStartosisInterpreter(testServiceNetwork, moduleContentProvider)
	script := `
print("Storing file from service!")
artifact_uuid=store_file_from_service(service_id="example-datastore-server", src_path="/foo/bar")
print(artifact_uuid)
`

	scriptOutput, interpretationError, instructions := interpreter.Interpret(context.Background(), script)
	require.Nil(t, interpretationError)
	require.Equal(t, 1, len(instructions))

	execInstruction := store_files_from_service.NewStoreFilesFromServicePosition(
		testServiceNetwork,
		*kurtosis_instruction.NewInstructionPosition(3, 38),
		"example-datastore-server",
		"/foo/bar",
	)

	require.Equal(t, instructions[0], execInstruction)

	expectedOutput := `Storing file from service!
{{kurtosis:3:38.artifact_uuid}}
`
	require.Equal(t, expectedOutput, string(scriptOutput))
}
