/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package basic_datastore_and_api_test

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/example-microservice/api/api_service_client"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	datastoreImage                        = "kurtosistech/example-microservices_datastore"
	datastoreServiceId services.ServiceID = "datastore"
	datastorePort                         = 1323

	apiServiceImage                    = "kurtosistech/example-microservices_api"
	apiServiceId    services.ServiceID = "api"
	apiServicePort                     = 2434

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15

	testPersonId     = 23
	testNumBooksRead = 3

	configFileKey = "config-file"
)

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort int    `json:"datastorePort"`
}

type BasicDatastoreAndApiTest struct {
	datastoreImage string
	apiImage       string
}

func NewBasicDatastoreAndApiTest(datastoreImage string, apiImage string) *BasicDatastoreAndApiTest {
	return &BasicDatastoreAndApiTest{datastoreImage: datastoreImage, apiImage: apiImage}
}

func (b BasicDatastoreAndApiTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (b BasicDatastoreAndApiTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {

	datastoreContainerCreationConfig, datastoreRunConfigFunc := getDatastoreServiceConfigurations()

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerCreationConfig, datastoreRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient := datastore_service_client.NewDatastoreClient(datastoreServiceContext.GetIPAddress(), datastorePort)

	err = datastoreClient.WaitForHealthy(waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)

	apiServiceContainerCreationConfig, apiServiceRunConfigFunc := getApiServiceConfigurations(datastoreClient)

	apiServiceContext, apiSvcHostPortBindings, err := networkCtx.AddService(apiServiceId, apiServiceContainerCreationConfig, apiServiceRunConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient := api_service_client.NewAPIClient(apiServiceContext.GetIPAddress(), apiServicePort)

	err = apiClient.WaitForHealthy(waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the api service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", apiSvcHostPortBindings)
	return networkCtx, nil
}

func (b BasicDatastoreAndApiTest) Run(network networks.Network) error {
	// Go doesn't have generics so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	serviceContext, err := castedNetwork.GetServiceContext(apiServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API service context")
	}

	apiClient := api_service_client.NewAPIClient(serviceContext.GetIPAddress(), apiServicePort)

	logrus.Infof("Verifying that person with test ID '%v' doesn't already exist...", testPersonId)
	if _, err = apiClient.GetPerson(testPersonId); err == nil {
		return stacktrace.NewError("Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
	}
	logrus.Infof("Verified that test person doesn't already exist")

	logrus.Infof("Adding test person with ID '%v'...", testPersonId)
	if err := apiClient.AddPerson(testPersonId); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding person with test ID '%v'", testPersonId)
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read by %v...", testNumBooksRead)
	for i := 0; i < testNumBooksRead; i++ {
		if err := apiClient.IncrementBooksRead(testPersonId); err != nil {
			return stacktrace.Propagate(err, "An error occurred incrementing the number of books read")
		}
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read...")
	person, err := apiClient.GetPerson(testPersonId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test person to verify the number of books read")
	}
	logrus.Info("Retrieved test person")

	if person.BooksRead != testNumBooksRead {
		return stacktrace.NewError(
			"Expected number of book read '%v' != actual number of books read '%v'",
			testNumBooksRead,
			person.BooksRead,
		)
	}

	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getDatastoreServiceConfigurations() (*services.ContainerCreationConfig, func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error)) {
	datastoreContainerCreationConfig := getDataStoreContainerCreationConfig()

	datastoreRunConfigFunc := getDataStoreRunConfigFunc()
	return datastoreContainerCreationConfig, datastoreRunConfigFunc
}

func getDataStoreContainerCreationConfig() *services.ContainerCreationConfig {
	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		datastoreImage,
	).WithUsedPorts(
		map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
	).Build()
	return containerCreationConfig
}

func getDataStoreRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	runConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		return services.NewContainerRunConfigBuilder().Build(), nil
	}
	return runConfigFunc
}

func getApiServiceConfigurations(datastoreClient *datastore_service_client.DatastoreClient) (*services.ContainerCreationConfig, func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error)) {
	configInitializingFunc := getApiServiceConfigInitializingFunc(datastoreClient)

	apiServiceContainerCreationConfig := getApiServiceContainerCreationConfig(configInitializingFunc)

	apiServiceRunConfigFunc := getApiServiceRunConfigFunc()
	return apiServiceContainerCreationConfig, apiServiceRunConfigFunc
}

func getApiServiceConfigInitializingFunc(datastoreClient *datastore_service_client.DatastoreClient) func(fp *os.File) error {
	configInitializingFunc := func(fp *os.File) error {
		logrus.Debugf("Datastore IP: %v , port: %v", datastoreClient.IpAddr(), datastoreClient.Port())
		configObj := datastoreConfig{
			DatastoreIp:   datastoreClient.IpAddr(),
			DatastorePort: datastoreClient.Port(),
		}
		configBytes, err := json.Marshal(configObj)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
		}

		logrus.Debugf("API config JSON: %v", string(configBytes))

		if _, err := fp.Write(configBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
		}

		return nil
	}
	return configInitializingFunc
}

func getApiServiceContainerCreationConfig(configInitializingFunc func(fp *os.File) error) *services.ContainerCreationConfig {
	apiServiceContainerCreationConfig := services.NewContainerCreationConfigBuilder(
		apiServiceImage,
	).WithUsedPorts(
		map[string]bool{fmt.Sprintf("%v/tcp", apiServicePort): true},
	).WithGeneratedFiles(map[string]func(*os.File) error{
		configFileKey: configInitializingFunc,
	}).Build()
	return apiServiceContainerCreationConfig
}

func getApiServiceRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	apiServiceRunConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		configFilepath, found := generatedFileFilepaths[configFileKey]
		if !found {
			return nil, stacktrace.NewError("No filepath found for config file key '%v'", configFileKey)
		}
		startCmd := []string{
			"./api.bin",
			"--config",
			configFilepath,
		}
		result := services.NewContainerRunConfigBuilder().WithCmdOverride(startCmd).Build()
		return result, nil
	}
	return apiServiceRunConfigFunc
}
