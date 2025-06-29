package interpretation_time_value_store

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"os"
	"testing"
)

const (
	testServiceName        = service.ServiceName("datastore-service")
	testContainerImageName = "datastore-image"
	enclaveDbFilePerm      = 0666
)

func TestGetServiceConfigReturnsError(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()
	itvs, err := CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(t, err)

	// no service config exists in store
	_, err = itvs.GetServiceConfig(testServiceName)
	require.Error(t, err)
}

func TestPutServiceConfig(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()
	itvs, err := CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(t, err)

	expectedServiceConfig, err := getTestServiceConfigForService(testServiceName, "latest")
	require.NoError(t, err)

	itvs.PutServiceConfig(testServiceName, expectedServiceConfig)

	actualServiceConfig, err := itvs.GetServiceConfig(testServiceName)
	require.NoError(t, err)
	require.Equal(t, expectedServiceConfig.GetContainerImageName(), actualServiceConfig.GetContainerImageName())
}

func TestPutNewServiceConfig(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	dummySerde := shared_helpers.NewDummyStarlarkValueSerDeForTest()
	itvs, err := CreateInterpretationTimeValueStore(enclaveDb, dummySerde)
	require.NoError(t, err)

	oldServiceConfig, err := getTestServiceConfigForService(testServiceName, "older")
	require.NoError(t, err)
	itvs.PutServiceConfig(testServiceName, oldServiceConfig)

	newerServiceConfig, err := getTestServiceConfigForService(testServiceName, "latest")
	require.NoError(t, err)
	itvs.SetServiceConfig(testServiceName, newerServiceConfig)

	actualNewerServiceConfig, err := itvs.GetNewServiceConfig(testServiceName)
	require.NoError(t, err)
	require.Equal(t, newerServiceConfig.GetContainerImageName(), actualNewerServiceConfig.GetContainerImageName())
}

func getTestServiceConfigForService(name service.ServiceName, imageTag string) (*service.ServiceConfig, error) {
	return service.CreateServiceConfig(fmt.Sprintf("%v-%v:%v", name, testContainerImageName, imageTag), nil, nil, nil, nil, nil, []string{}, []string{}, map[string]string{}, nil, nil, 0, 0, "IP-ADDRESS", 0, 0, map[string]string{}, nil, nil, nil, image_download_mode.ImageDownloadMode_Always, true, false)
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), enclaveDbFilePerm, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}
