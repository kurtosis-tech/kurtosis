package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"net"
	"testing"
)

func TestAddServiceShared_EntryPointArgsWithIpAddressAndRuntimeValueAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceName]net.IP{
		"foo_service": net.ParseIP("172.17.3.13"),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)

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
		[]string{"-- {{kurtosis:foo_service.ip_address}} " + runtimeValue},
	).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(serviceNetwork, runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "-- 172.17.3.13 8765", replacedServiceConfig.EntrypointArgs[0])
}

func TestAddServiceShared_CmdArgsWithIpAddressAndRuntimeValueAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceName]net.IP{
		"foo_service": net.ParseIP("172.17.3.13"),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)

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
		[]string{"bash", "-c", "ping {{kurtosis:foo_service.ip_address}} && sleep " + runtimeValue},
	).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(serviceNetwork, runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "ping 172.17.3.13 && sleep 999999", replacedServiceConfig.CmdArgs[2])
}

func TestAddServiceShared_EnvVarsWithIpAddressAndRuntimeValueAreReplaced(t *testing.T) {
	ipAddresses := map[service.ServiceName]net.IP{
		"foo_service": net.ParseIP("172.17.3.13"),
	}
	serviceNetwork := service_network.NewMockServiceNetwork(ipAddresses)

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
		"IP_ADDRESS": "{{kurtosis:foo_service.ip_address}}",
		"PORT":       runtimeValue,
	}).Build()

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(serviceNetwork, runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	expectedEnvVars := map[string]string{
		"IP_ADDRESS": "172.17.3.13",
		"PORT":       "8765",
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

	replacedServiceName, _, err := replaceMagicStrings(nil, runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, service.ServiceName("database-1"), replacedServiceName)
}
