/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networks_impl

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/example-microservice/api/api_service_client"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
)

const (
	datastoreImage                        = "kurtosistech/example-microservices_datastore"
	datastoreServiceId services.ServiceID = "datastore"
	datastorePort                         = 1323

	apiServiceImage    = "kurtosistech/example-microservices_api"
	apiServiceIdPrefix = "api-"
	apiServicePort     = 2434

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxNumPolls       = 15
	configFileKey                   = "config-file.txt"
)

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort int    `json:"datastorePort"`
}

//  A custom Network implementation is intended to make test-writing easier by wrapping low-level
//    NetworkContext calls with custom higher-level business logic
type TestNetwork struct {
	networkCtx                *networks.NetworkContext
	datastoreServiceImage     string
	apiServiceImage           string
	datastoreClient           *datastore_service_client.DatastoreClient
	personModifyingApiClient  *api_service_client.APIClient
	personRetrievingApiClient *api_service_client.APIClient
	nextApiServiceId          int
}

func NewTestNetwork(networkCtx *networks.NetworkContext, datastoreServiceImage string, apiServiceImage string) *TestNetwork {
	return &TestNetwork{
		networkCtx:                networkCtx,
		datastoreServiceImage:     datastoreServiceImage,
		apiServiceImage:           apiServiceImage,
		datastoreClient:           nil,
		personModifyingApiClient:  nil,
		personRetrievingApiClient: nil,
		nextApiServiceId:          0,
	}
}

//  Custom network implementations usually have a "setup" method (possibly parameterized) that is used
//   in the Test.Setup function of each test
func (network *TestNetwork) SetupDatastoreAndTwoApis() error {

	if network.datastoreClient != nil {
		return stacktrace.NewError("Cannot add datastore client to network; datastore client already exists!")
	}

	if network.personModifyingApiClient != nil || network.personRetrievingApiClient != nil {
		return stacktrace.NewError("Cannot add API services to network; one or more API services already exists")
	}

	datastoreContainerConfigSupplier := getDatastoreContainerConfigSupplier()

	datastoreServiceContext, hostPortBindings, err := network.networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceContext.GetIPAddress(), datastorePort)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)

	network.datastoreClient = datastoreClient

	personModifyingApiClient, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-modifying API client")
	}
	network.personModifyingApiClient = personModifyingApiClient

	personRetrievingApiClient, err := network.addApiService()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-retrieving API client")
	}
	network.personRetrievingApiClient = personRetrievingApiClient

	return nil
}

//  Custom network implementations will also usually have getters, to retrieve information about the
//   services created during setup
func (network *TestNetwork) GetPersonModifyingApiClient() (*api_service_client.APIClient, error) {
	if network.personModifyingApiClient == nil {
		return nil, stacktrace.NewError("No person-modifying API client exists")
	}
	return network.personModifyingApiClient, nil
}
func (network *TestNetwork) GetPersonRetrievingApiClient() (*api_service_client.APIClient, error) {
	if network.personRetrievingApiClient == nil {
		return nil, stacktrace.NewError("No person-retrieving API client exists")
	}
	return network.personRetrievingApiClient, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (network *TestNetwork) addApiService() (*api_service_client.APIClient, error) {

	if network.datastoreClient == nil {
		return nil, stacktrace.NewError("Cannot add API service to network; no datastore client exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	apiServiceContainerConfigSupplier := getApiServiceContainerConfigSupplier(network)

	apiServiceContext, hostPortBindings, err := network.networkCtx.AddService(serviceId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), apiServicePort)

	err = apiClient.WaitForHealthy(waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	return apiClient, nil
}

func getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
				datastoreImage,
			).WithUsedPorts(
				map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		    ).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getApiServiceContainerConfigSupplier(network *TestNetwork) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		datastoreConfigFileFilePath, err := createDatastoreConfigFileInServiceDirectory(network, sharedDirectory)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating data store config file in service container")
		}

		startCmd := []string{
			"./api.bin",
			"--config",
			datastoreConfigFileFilePath.GetAbsPathOnServiceContainer(),
		}

		containerConfig := services.NewContainerConfigBuilder(
				apiServiceImage,
			).WithUsedPorts(
				map[string]bool{fmt.Sprintf("%v/tcp", apiServicePort): true},
			).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}

func createDatastoreConfigFileInServiceDirectory(network *TestNetwork, sharedDirectory *services.SharedPath) (*services.SharedPath, error) {
	configFileFilePath, err := sharedDirectory.GetChildPath(configFileKey)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting file object '%v' from shared directory", configFileKey)
	}

	logrus.Infof("Config file absolute path on this container: %v , on service container: %v", configFileFilePath.GetAbsPathOnThisContainer(), configFileFilePath.GetAbsPathOnServiceContainer())

	datastoreClient := network.datastoreClient

	logrus.Debugf("Datastore IP: %v , port: %v", datastoreClient.IpAddr(), datastoreClient.Port())

	configObj := datastoreConfig{
		DatastoreIp:   datastoreClient.IpAddr(),
		DatastorePort: datastoreClient.Port(),
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	logrus.Debugf("API config JSON: %v", string(configBytes))

	if err := ioutil.WriteFile(configFileFilePath.GetAbsPathOnThisContainer(), configBytes, os.ModePerm); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return configFileFilePath, nil
}
