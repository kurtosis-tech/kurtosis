package update_service

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

var (
	thread = &starlark.Thread{
		Name:       "test-update-service",
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
	}
)

func TestUpdateService_Interpreter(t *testing.T) {
	var instructions []kurtosis_instruction.KurtosisInstruction
	starlarkInstruction := `update_service("datastore-service", UpdateServiceConfig("subnetwork_1"))`
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, starlarkInstruction, starlark.StringDict{
		kurtosis_types.UpdateServiceConfigTypeName: starlark.NewBuiltin(kurtosis_types.UpdateServiceConfigTypeName, kurtosis_types.MakeUpdateServiceConfig),
		UpdateServiceBuiltinName:                   starlark.NewBuiltin(UpdateServiceBuiltinName, GenerateUpdateServiceBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	expectedInstruction := NewUpdateServiceInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		"datastore-service",
		binding_constructors.NewUpdateServiceConfig("subnetwork_1"),
		starlark.StringDict{
			"service_name": starlark.String("datastore-service"),
			"config":       kurtosis_types.NewUpdateServiceConfig("subnetwork_1"),
		})
	require.Equal(t, expectedInstruction, instructions[0])
}

func TestUpdateService_GetCanonicalizedInstruction(t *testing.T) {
	serviceName := starlark.String("datastore-service")
	updateServiceConfig := kurtosis_types.NewUpdateServiceConfig("subnetwork_1")
	updateServiceInstruction := newEmptyUpdateServiceInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(22, 26, "dummyFile"),
	)
	updateServiceInstruction.starlarkKwargs[serviceNameArgName] = serviceName
	updateServiceInstruction.starlarkKwargs[updateServiceConfigArgName] = updateServiceConfig

	expectedOutput := `update_service(config=UpdateServiceConfig(subnetwork="subnetwork_1"), service_name="datastore-service")`
	require.Equal(t, expectedOutput, updateServiceInstruction.String())
}

func TestUpdateService_SerializeAndParseAgain(t *testing.T) {
	initialInstruction := NewUpdateServiceInstruction(
		nil,
		kurtosis_instruction.NewInstructionPosition(1, 15, startosis_constants.PackageIdPlaceholderForStandaloneScript),
		"datastore-service",
		binding_constructors.NewUpdateServiceConfig("subnetwork_1"),
		starlark.StringDict{
			"service_name": starlark.String("datastore-service"),
			"config":       kurtosis_types.NewUpdateServiceConfig("subnetwork_1"),
		})

	canonicalizedInstruction := initialInstruction.String()

	var instructions []kurtosis_instruction.KurtosisInstruction
	_, err := starlark.ExecFile(thread, startosis_constants.PackageIdPlaceholderForStandaloneScript, canonicalizedInstruction, starlark.StringDict{
		kurtosis_types.UpdateServiceConfigTypeName: starlark.NewBuiltin(kurtosis_types.UpdateServiceConfigTypeName, kurtosis_types.MakeUpdateServiceConfig),
		UpdateServiceBuiltinName:                   starlark.NewBuiltin(UpdateServiceBuiltinName, GenerateUpdateServiceBuiltin(&instructions, nil)),
	})
	require.Nil(t, err)

	require.Len(t, instructions, 1)
	require.Equal(t, initialInstruction, instructions[0])
}
