package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartosisCompiler_SimplePrintScript(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Hello World!")
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions)) // No kurtosis instruction
	assert.Nil(t, interpretationError)
	assert.Nil(t, err)

	expectedOutput := `Hello World!
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", scriptOutput.Get()))
}

func TestStartosisCompiler_ScriptFailingSingleError(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

unknownInstruction()
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)

	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	[4:1]: undefined: unknownInstruction
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ScriptFailingMultipleErrors(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

unknownInstruction()
print(unknownVariable)

unknownInstruction2()
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)

	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	[4:1]: undefined: unknownInstruction
	[5:7]: undefined: unknownVariable
	[7:1]: undefined: unknownInstruction2
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ScriptFailingSyntaxError(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

load("otherScript.start") # fails b/c load takes in at least 2 args
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)

	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	[4:5]: load statement must import at least 1 symbol
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ScriptFailingLoadBindingError(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

load("otherScript.star", "a") # fails b/c load current binding throws
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)

	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	EvaluationError: cannot load otherScript.star: Loading external Startosis scripts is not supported yet
		at [4:1]: <toplevel>
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ValidSimpleScriptWithInstruction(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
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

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 1, len(instructions))
	assert.Nil(t, interpretationError)
	assert.Nil(t, err)

	addServiceInstruction := kurtosis_instruction.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(13, 12),
		service.ServiceID("example-datastore-server"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	assert.Equal(t, instructions[0], addServiceInstruction)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", scriptOutput.Get()))
}

func TestStartosisCompiler_ValidSimpleScriptWithInstructionMissingContainerName(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
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

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)
	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	EvaluationError: Missing ` + "`container_image_name`" + ` as part of the struct object
		at [13:12]: <toplevel>
		at [0:0]: add_service
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ValidSimpleScriptWithInstructionTypoInProtocol(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = 1323, protocol = "TCPK") # typo in protocol
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)
	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	EvaluationError: port protocol should be either ` + "`TCP`, `UDP` or `SCTP`" + `
		at [13:12]: <toplevel>
		at [0:0]: add_service
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ValidSimpleScriptWithInstructionPortNumberAsString(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

service_id = "example-datastore-server"
print("Adding service " + service_id)

service_config = struct(
	container_image_name = "kurtosistech/example-datastore-server",
	used_ports = {
		"grpc": struct(number = "1234", protocol = "TCP") # port number should be an int
	}
)
add_service(service_id = service_id, service_config = service_config)
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 0, len(instructions))
	assert.Nil(t, scriptOutput)
	assert.Nil(t, err)
	expectedOutput := `/!\ Errors interpreting Startosis script /!\
	EvaluationError: ` + "`number` arg is expected to be a uint32" + `
		at [13:12]: <toplevel>
		at [0:0]: add_service
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", interpretationError.Get()))
}

func TestStartosisCompiler_ValidScriptWithMultipleInstructions(t *testing.T) {
	interpreter := NewStartosisInterpreter(nil)
	script := `
print("Starting Startosis script!")

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
                        protocol = "TCP")
                })
        add_service(service_id = unique_service_id, service_config = service_config)

deploy_datastore_services()
print("Done!")
`

	scriptOutput, interpretationError, instructions, err := interpreter.Interpret(context.Background(), script)
	assert.Equal(t, 3, len(instructions))
	assert.Nil(t, interpretationError)
	assert.Nil(t, err)

	addServiceInstruction0 := kurtosis_instruction.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(20, 26),
		service.ServiceID("example-datastore-server-0"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1323,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction1 := kurtosis_instruction.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(20, 26),
		service.ServiceID("example-datastore-server-1"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1324,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})
	addServiceInstruction2 := kurtosis_instruction.NewAddServiceInstruction(
		nil,
		*kurtosis_instruction.NewInstructionPosition(20, 26),
		service.ServiceID("example-datastore-server-2"),
		&kurtosis_core_rpc_api_bindings.ServiceConfig{
			ContainerImageName: "kurtosistech/example-datastore-server",
			PrivatePorts: map[string]*kurtosis_core_rpc_api_bindings.Port{
				"grpc": {
					Number:   1325,
					Protocol: kurtosis_core_rpc_api_bindings.Port_TCP,
				},
			},
		})

	assert.Equal(t, instructions[0], addServiceInstruction0)
	assert.Equal(t, instructions[1], addServiceInstruction1)
	assert.Equal(t, instructions[2], addServiceInstruction2)

	expectedOutput := `Starting Startosis script!
Adding service example-datastore-server-0
Adding service example-datastore-server-1
Adding service example-datastore-server-2
Done!
`
	assert.Equal(t, expectedOutput, fmt.Sprintf("%s", scriptOutput.Get()))
}
