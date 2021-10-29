/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package basic_datastore_and_api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	datastorePort                         = 1323

	apiServiceId    services.ServiceID = "api"
	apiServicePort                     = 2434

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15

	testPersonId     = "23"
	testNumBooksRead = 3

	configFilepathRelativeToSharedDirRoot = "config-file.txt"
)

type GRPCAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

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

func (test BasicDatastoreAndApiTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test BasicDatastoreAndApiTest) Setup(networkCtx *networks.NetworkContext) (network networks.Network, returnErr error) {
	ctx := context.Background()

	datastoreContainerConfigSupplier := test.getDatastoreContainerConfigSupplier()

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := newDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		err = datastoreClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	err = waitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)

	apiServiceContainerConfigSupplier := test.getApiServiceContainerConfigSupplier(datastoreServiceContext.GetIPAddress())

	apiServiceContext, apiSvcHostPortBindings, err := networkCtx.AddService(apiServiceId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, apiClientConnCloseFunc, err := newExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", apiServiceId, apiServiceContext.GetIPAddress())
	}
	defer func() {
		err = apiClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	err = waitForHealthy(ctx, apiClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", apiSvcHostPortBindings)
	return networkCtx, returnErr
}

func (test BasicDatastoreAndApiTest) Run(network networks.Network) (returnErr error) {
	ctx := context.Background()

	// Go doesn't have generics, so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	serviceContext, err := castedNetwork.GetServiceContext(apiServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API service context")
	}

	apiClient, apiClientConnCloseFunc, err := newExampleAPIServerClient(serviceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", apiServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		err = apiClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	logrus.Infof("Verifying that person with test ID '%v' doesn't already exist...", testPersonId)
	getPersonArgs := &example_api_server_rpc_api_bindings.GetPersonArgs{
		PersonId: testPersonId,
	}
	if _, err = apiClient.GetPerson(ctx, getPersonArgs); err == nil {
		return stacktrace.NewError("Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
	}
	logrus.Infof("Verified that test person doesn't already exist")

	logrus.Infof("Adding test person with ID '%v'...", testPersonId)
	addPersonArgs := &example_api_server_rpc_api_bindings.AddPersonArgs{
		PersonId: testPersonId,
	}
	if _, err := apiClient.AddPerson(ctx, addPersonArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred adding test person with ID '%v'", testPersonId)
	}
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read by %v...", testNumBooksRead)
	for i := 0; i < testNumBooksRead; i++ {
		incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
			PersonId: testPersonId,
		}
		if _, err := apiClient.IncrementBooksRead(ctx, incrementBooksReadArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred incrementing the number of books read")
		}
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read...")
	getPersonResponse, err := apiClient.GetPerson(ctx, getPersonArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test person to verify the number of books read")
	}
	logrus.Info("Retrieved test person")

	personBooksReadBase := 10
	personBooksReadBitSize := 32

	personBooksRead, err := strconv.ParseInt(getPersonResponse.GetBooksRead(), personBooksReadBase, personBooksReadBitSize)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing person books read string '%v' to int", personBooksRead)
	}

	if personBooksRead != testNumBooksRead {
		return stacktrace.NewError(
			"Expected number of book read '%v' != actual number of books read '%v'",
			testNumBooksRead,
			personBooksRead,
		)
	}

	return returnErr
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (test BasicDatastoreAndApiTest) getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			test.datastoreImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func (test BasicDatastoreAndApiTest) getApiServiceContainerConfigSupplier(datastoreIP string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		datastoreConfigFileFilePath, err := createDatastoreConfigFileInServiceDirectory(datastoreIP, sharedDirectory)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating data store config file in service container")
		}

		startCmd := []string{
			"./example-api-server.bin",
			"--config",
			datastoreConfigFileFilePath.GetAbsPathOnServiceContainer(),
		}

		containerConfig := services.NewContainerConfigBuilder(
			test.apiImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", apiServicePort): true},
		).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}

func createDatastoreConfigFileInServiceDirectory(datastoreIP string, sharedDirectory *services.SharedPath) (*services.SharedPath, error) {
	configFileFilePath := sharedDirectory.GetChildPath(configFilepathRelativeToSharedDirRoot)

	logrus.Infof("Config file absolute path on this container: %v , on service container: %v", configFileFilePath.GetAbsPathOnThisContainer(), configFileFilePath.GetAbsPathOnServiceContainer())

	logrus.Debugf("Datastore IP: %v , port: %v", datastoreIP, datastorePort)

	configObj := datastoreConfig{
		DatastoreIp:   datastoreIP,
		DatastorePort: datastorePort,
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

func newDatastoreClient(datastoreIp string) (datastore_rpc_api_bindings.DatastoreServiceClient, func() error, error) {
	datastoreURL := fmt.Sprintf(
		"%v:%v",
		datastoreIp,
		datastorePort,
	)

	conn, err := grpc.Dial(datastoreURL, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling the datastore container via its URL")
	}

	datastoreServiceClient := datastore_rpc_api_bindings.NewDatastoreServiceClient(conn)

	return datastoreServiceClient, conn.Close, nil
}

func newExampleAPIServerClient(exampleAPIServerIp string) (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {
	exampleAPIServerURL := fmt.Sprintf(
		"%v:%v",
		exampleAPIServerIp,
		apiServicePort,
	)

	conn, err := grpc.Dial(exampleAPIServerURL, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred dialling the example API server container via its URL")
	}

	exampleAPIServerClient := example_api_server_rpc_api_bindings.NewExampleAPIServerServiceClient(conn)

	return exampleAPIServerClient, conn.Close, nil
}

func waitForHealthy(ctx context.Context, client GRPCAvailabilityChecker, retries uint32, retriesDelayMilliseconds uint32) error {

	var (
		emptyArgs = &empty.Empty{}
		err       error
	)

	for i := uint32(0); i < retries; i++ {
		_, err = client.IsAvailable(ctx, emptyArgs)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(err,
			"The datastore service didn't return a success code, even after %v retries with %v milliseconds in between retries",
			retries, retriesDelayMilliseconds)
	}

	return nil
}
