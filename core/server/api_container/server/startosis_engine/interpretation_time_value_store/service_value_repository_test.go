package interpretation_time_value_store

import (
	"fmt"
	port_spec_core "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/port_spec"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"os"
	"testing"
)

const (
	starlarkThreadName     = "thread-for-db-test"
	serviceName            = service.ServiceName("datastore")
	serviceNameStarlarkStr = starlark.String(serviceName)
	hostName               = serviceNameStarlarkStr
	ipAddress              = starlark.String("172.23.34.44")
)

func TestPutGet_Succeed(t *testing.T) {
	repository := getServiceInterpretationTimeValueRepository(t)
	require.NotNil(t, repository)

	applicationProtocol := ""
	maybeUrl := ""

	port, interpretationErr := port_spec.CreatePortSpecUsingGoValues(
		string(serviceName),
		uint16(443),
		port_spec_core.TransportProtocol_TCP,
		&applicationProtocol,
		"10s",
		&maybeUrl,
	)
	require.Nil(t, interpretationErr)
	ports := starlark.NewDict(1)
	require.NoError(t, ports.SetKey(starlark.String("http"), port))

	expectedService, interpretationErr := kurtosis_types.CreateService(serviceNameStarlarkStr, hostName, ipAddress, ports)
	require.Nil(t, interpretationErr)

	err := repository.PutService(serviceName, expectedService)
	require.Nil(t, err)

	actualService, err := repository.GetService(serviceName)
	require.Nil(t, err)
	require.Equal(t, expectedService.AttrNames(), actualService.AttrNames())
	require.Equal(t, expectedService.String(), actualService.String())
}

func TestPutGetFail_ForMissingServiceName(t *testing.T) {
	repository := getServiceInterpretationTimeValueRepository(t)
	require.NotNil(t, repository)

	actualService, err := repository.GetService(serviceName)
	require.Error(t, err)
	require.Nil(t, actualService)
}

func TestGetServices(t *testing.T) {
	repository := getServiceInterpretationTimeValueRepository(t)
	require.NotNil(t, repository)

	applicationProtocol := ""
	maybeUrl := ""

	port, interpretationErr := port_spec.CreatePortSpecUsingGoValues(
		string(serviceName),
		uint16(443),
		port_spec_core.TransportProtocol_TCP,
		&applicationProtocol,
		"10s",
		&maybeUrl,
	)
	require.Nil(t, interpretationErr)
	ports := starlark.NewDict(1)
	require.NoError(t, ports.SetKey(starlark.String("http"), port))

	serviceOneName := fmt.Sprintf("%v-%v", serviceNameStarlarkStr, "1")
	serviceOne, interpretationErr := kurtosis_types.CreateService(starlark.String(serviceOneName), hostName, ipAddress, ports)
	require.Nil(t, interpretationErr)

	serviceTwoName := fmt.Sprintf("%v-%v", serviceNameStarlarkStr, "2")
	serviceTwo, interpretationErr := kurtosis_types.CreateService(starlark.String(serviceTwoName), hostName, ipAddress, ports)
	require.Nil(t, interpretationErr)

	err := repository.PutService(service.ServiceName(serviceOneName), serviceOne)
	require.Nil(t, err)
	err = repository.PutService(service.ServiceName(serviceTwoName), serviceTwo)
	require.Nil(t, err)

	actualServices, err := repository.GetServices()
	require.NoError(t, err)
	require.Len(t, actualServices, 2)
	// TODO: check the contents of each objects
}

func getServiceInterpretationTimeValueRepository(t *testing.T) *serviceInterpretationValueRepository {
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

	dummySerde := newDummyStarlarkValueSerDeForTest()

	repository, err := getOrCreateNewServiceInterpretationTimeValueRepository(enclaveDb, dummySerde)
	require.NoError(t, err)

	return repository
}

func newDummyStarlarkValueSerDeForTest() *kurtosis_types.StarlarkValueSerde {
	thread := &starlark.Thread{
		Name:       starlarkThreadName,
		Print:      nil,
		Load:       nil,
		OnMaxSteps: nil,
		Steps:      0,
	}
	starlarkEnv := starlark.StringDict{
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make),

		kurtosis_types.ServiceTypeName: starlark.NewBuiltin(kurtosis_types.ServiceTypeName, kurtosis_types.NewServiceType().CreateBuiltin()),
		port_spec.PortSpecTypeName:     starlark.NewBuiltin(port_spec.PortSpecTypeName, port_spec.NewPortSpecType().CreateBuiltin()),
	}
	return kurtosis_types.NewStarlarkValueSerde(thread, starlarkEnv)
}
