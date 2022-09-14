package bulk_api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/services"
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
	numServicesToAdd                      = 3
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
	datastoreServiceConfigSuppliers := map[services.ServiceID]*services.ContainerConfig{}
	for i := 0; i < numServicesToAdd; i++ {
		serviceID := services.ServiceID(fmt.Sprintf("%v-%v", datastoreServiceId, i))
		datastoreServiceIDs[i] = serviceID
		datastoreServiceConfigSuppliers[serviceID] = getDatastoreContainerConfigSupplier()
	}

	logrus.Infof("Adding three datastore services simultaneously...")
	successfulDatastoreServiceContexts, failedDatastoreServiceErrs, err := enclaveCtx.AddServices(datastoreServiceConfigSuppliers)
	require.NoError(t, err, "An error occurred adding the datastore services to the enclave")
	logrus.Infof("Added datastore service")
	require.Equal(t, numServicesToAdd, len(successfulDatastoreServiceContexts))
	require.Equal(t, 0, len(failedDatastoreServiceErrs))

	apiServiceConfigSuppliers := map[services.ServiceID]*services.ContainerConfig{}
	for i := 0; i < numServicesToAdd; i++ {
		datastoreServiceID, found := datastoreServiceIDs[i]
		require.True(t, found)
		datastoreServiceCtx, found := successfulDatastoreServiceContexts[datastoreServiceID]
		require.True(t, found)
		configFilepath, err := createApiConfigFile(datastoreServiceCtx.GetPrivateIPAddress())
		require.NoErrorf(t, err, "An error occurred creating api config file for service '%v'", datastoreServiceID)
		datastoreConfigArtifactUUID, err := enclaveCtx.UploadFiles(configFilepath)
		require.NoErrorf(t, err, "An error occurred uploading files to enclave for service '%v'", datastoreServiceID)
		serviceID := fmt.Sprintf("%v-%v", apiServiceID, i)
		apiServiceConfigSuppliers[services.ServiceID(serviceID)] = getContainerConfig(datastoreConfigArtifactUUID)
	}
	logrus.Infof("Adding three api services simultaneously...")
	successfulAPIServiceCtx, failedAPIServiceErrs, err := enclaveCtx.AddServices(apiServiceConfigSuppliers)
	require.NoError(t, err, "An error occurred adding the api services to the enclave")
	require.Equal(t, numServicesToAdd, len(successfulAPIServiceCtx))
	require.Equal(t, 0, len(failedAPIServiceErrs))
}

func getDatastoreContainerConfigSupplier() *services.ContainerConfig {
	containerConfigSupplier := func(ipAddr string) *services.ContainerConfig {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			datastorePortId: datastorePortSpec,
		}).Build()
		return containerConfig
	}
	return containerConfigSupplier("foo")
}

func getContainerConfig(apiConfigArtifactUuid services.FilesArtifactUUID) *services.ContainerConfig {
	containerConfig := func(ipAddr string) (*services.ContainerConfig) {
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

		return containerConfig
	}

	return containerConfig("foo")
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
