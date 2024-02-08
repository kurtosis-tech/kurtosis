package add_service

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"testing"
)

const (
	testContainerImageName = "kurtosistech/example-datastore-server"
)

func TestAddServiceShared_EntryPointArgsRuntimeValueAreReplaced(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)

	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(8765),
	})
	require.NoError(t, err)
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		nil,
		nil,
		nil,
		[]string{"-- " + runtimeValue},
		nil,
		nil,
		nil,
		nil,
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		nil,
		map[string]string{},
	)
	require.NoError(t, err)

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "-- 8765", replacedServiceConfig.GetEntrypointArgs()[0])
}

func TestAddServiceShared_CmdArgsRuntimeValueAreReplaced(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(999999),
	})
	require.NoError(t, err)
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		nil,
		nil,
		nil,
		nil,
		[]string{"bash", "-c", "sleep " + runtimeValue},
		nil,
		nil,
		nil,
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		nil,
		map[string]string{},
	)
	require.NoError(t, err)

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	require.Equal(t, "sleep 999999", replacedServiceConfig.GetCmdArgs()[2])
}

func TestAddServiceShared_EnvVarsWithRuntimeValueAreReplaced(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	runtimeValueName := "value"
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		runtimeValueName: starlark.MakeInt(8765),
	})
	require.NoError(t, err)
	runtimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, runtimeValueName)

	serviceName := service.ServiceName("example-datastore-server-2")
	serviceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		map[string]string{
			"PORT": runtimeValue,
		},
		nil,
		nil,
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		nil,
		map[string]string{},
	)
	require.NoError(t, err)

	replacedServiceName, replacedServiceConfig, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, serviceName, replacedServiceName)
	expectedEnvVars := map[string]string{
		"PORT": "8765",
	}
	require.Equal(t, expectedEnvVars, replacedServiceConfig.GetEnvVars())
}

func TestAddServiceShared_ServiceNameWithRuntimeValuesAreReplaced(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)

	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()

	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(dummySerde, enclaveDb)
	require.NoError(t, err)
	stringValueUuid, err := runtimeValueStore.CreateValue()
	require.Nil(t, err, "error creating a runtime value UUID")
	valueName := "value"
	err = runtimeValueStore.SetValue(stringValueUuid, map[string]starlark.Comparable{
		"value": starlark.String("database-1"),
	})
	require.NoError(t, err)
	stringRuntimeValue := fmt.Sprintf(magic_string_helper.RuntimeValueReplacementPlaceholderFormat, stringValueUuid, valueName)

	serviceName := service.ServiceName(stringRuntimeValue)
	serviceConfig, err := service.CreateServiceConfig(
		testContainerImageName,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		0,
		0,
		"",
		0,
		0,
		map[string]string{},
		nil,
		nil,
		map[string]string{},
	)
	require.NoError(t, err)

	replacedServiceName, _, err := replaceMagicStrings(runtimeValueStore, serviceName, serviceConfig)
	require.Nil(t, err)
	require.Equal(t, service.ServiceName("database-1"), replacedServiceName)
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}
