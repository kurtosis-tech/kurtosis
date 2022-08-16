package bulk_api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	testName              = "bulk-api"
	isPartitioningEnabled = false

	configFilename                = "config.json"
	configMountpathOnApiContainer = "/config"

	datastoreImage  = "kurtosistech/example-datastore-server"
	apiServiceImage = "kurtosistech/example-api-server"

	datastorePortId string = "rpc"
	apiPortId       string = "rpc"

	datastoreServiceId services.ServiceID = "datastore"
	apiServiceID       services.ServiceID = "api-service"
)

var datastorePortSpec = services.NewPortSpec(
	datastore_rpc_api_consts.ListenPort,
	services.PortProtocol_TCP,
)
var apiPortSpec = services.NewPortSpec(
	example_api_server_rpc_api_consts.ListenPort,
	services.PortProtocol_TCP,
)

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort uint16 `json:"datastorePort"`
}

func TestAddingDatastoreServicesInBulk(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	datastoreServiceIDs := map[int]services.ServiceID{}
	datastoreServiceConfigSuppliers := map[services.ServiceID]func(string) (*services.ContainerConfig, error){}
	for i := 0; i < 3; i++ {
		serviceID := services.ServiceID(fmt.Sprintf("%v-%v", datastoreServiceId, i))
		datastoreServiceIDs[i] = serviceID
		datastoreServiceConfigSuppliers[serviceID] = getDatastoreContainerConfigSupplier()
	}

	logrus.Infof("Adding three datastore services simultaneously...")
	successfulDatastoreServiceContexts, failedDatatstoreErrs, err := enclaveCtx.AddServices(datastoreServiceConfigSuppliers)
	require.NoError(t, err, "An error occurred adding the datastore services to the enclave")
	logrus.Infof("Added datastore service")
	require.Equal(t, 3, len(successfulDatastoreServiceContexts))
	require.Equal(t, 0, len(failedDatatstoreErrs))

	apiServiceConfigSuppliers := map[services.ServiceID]func(string) (*services.ContainerConfig, error){}
	for i := 0; i < 3; i++ {
		datastoreServiceCtx := successfulDatastoreServiceContexts[datastoreServiceIDs[i]]
		configFilepath, err := createApiConfigFile(datastoreServiceCtx.GetPrivateIPAddress())
		require.NoError(t, err, "An error occurred creating api config file for service.")
		datastoreConfigArtifactUuid, err := enclaveCtx.UploadFiles(configFilepath)
		require.NoError(t, err, "An error occurred uploading files to enclave for service.\"")
		serviceID := fmt.Sprintf("%v-%v", apiServiceID, i)
		apiServiceConfigSuppliers[services.ServiceID(serviceID)] = getApiServiceContainerConfigSupplier(datastoreConfigArtifactUuid)
	}
	logrus.Infof("Adding three api services simultaneously...")
	successfulAPIServiceCtx, failedAPIServiceErrs, err := enclaveCtx.AddServices(apiServiceConfigSuppliers)
	require.NoError(t, err, "An error occurred adding the api services to the enclave")
	logrus.Infof("Added datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------
	require.Equal(t, 3, len(successfulAPIServiceCtx))
	require.Equal(t, 0, len(failedAPIServiceErrs))
}

func getDatastoreContainerConfigSupplier() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			datastorePortId: datastorePortSpec,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getApiServiceContainerConfigSupplier(apiConfigArtifactUuid services.FilesArtifactUUID) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		startCmd := []string{
			"./example-api-server.bin",
			"--config",
			path.Join(configMountpathOnApiContainer, configFilename),
		}

		containerConfig := services.NewContainerConfigBuilder(
			apiServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			apiPortId: apiPortSpec,
		}).WithFiles(map[services.FilesArtifactUUID]string{
			apiConfigArtifactUuid: configMountpathOnApiContainer,
		}).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}

func createApiConfigFile(datastoreIP string) (string, error) {
	tempDirpath, err := ioutil.TempDir("", "")
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating a temporary directory to house the datastore config file")
	}
	tempFilepath := path.Join(tempDirpath, configFilename)

	configObj := datastoreConfig{
		DatastoreIp:   datastoreIP,
		DatastorePort: datastore_rpc_api_consts.ListenPort,
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	if err := ioutil.WriteFile(tempFilepath, configBytes, os.ModePerm); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return tempFilepath, nil
}
