package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"testing"
)

const (
	testContainerImageName = "kurtosistech/example-datastore-server"
)

func TestAddServiceShared_EntryPointArgsRuntimeValueAreReplaced(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(8765),
	})
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig := services.NewServiceConfigBuilder(
		testContainerImageName,
	).WithEntryPointArgs(
		[]string{"-- " + runtimeValue},
	).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "-- 8765", replacedServiceConfig.EntrypointArgs[0])
}

func TestAddServiceShared_CmdArgsRuntimeValueAreReplaced(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(999999),
	})
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig := services.NewServiceConfigBuilder(
		testContainerImageName,
	).WithCmdArgs(
		[]string{"bash", "-c", "sleep " + runtimeValue},
	).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "sleep 999999", replacedServiceConfig.CmdArgs[2])
}

func TestAddServiceShared_EnvVarsWithRuntimeValueAreReplaced(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(8765),
	})
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig := services.NewServiceConfigBuilder(
		testContainerImageName,
	).WithEnvVars(map[string]string{
		"PORT": runtimeValue,
	}).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	expectedEnvVars := map[string]string{
		"PORT": "8765",
	}
	require.Equal(t, expectedEnvVars, replacedServiceConfig.EnvVars)
}

func TestAddServiceShared_ServiceNameWithRuntimeValuesAreReplaced(t *testing.T) {
	runtimeValueStore := runtime_value_store.NewRuntimeValueStore()
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	valueName := "value"
	runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		"value": starlark.String("database-1"),
	})
	stringRuntimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, valueName)

	serviceName := service.ServiceName(stringRuntimeValue)
	serviceConfig := services.NewServiceConfigBuilder(
		testContainerImageName,
	).Build()

	replacedServiceName, _, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, service.ServiceName("database-1"), replacedServiceName)
}
