/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networks_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
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

	apiServiceIdPrefix = "api-"
	apiServicePort     = 2434

	waitForStartupDelayMilliseconds       = 1000
	waitForStartupMaxNumPolls             = 15
	configFilepathRelativeToSharedDirRoot = "config-file.txt"
)

type GRPCAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

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
	datastoreClient           datastore_rpc_api_bindings.DatastoreServiceClient
	personModifyingApiClient  example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient
	personRetrievingApiClient example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient
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
func (network *TestNetwork) SetupDatastoreAndTwoApis() (returnErr error) {
	ctx := context.Background()

	if network.datastoreClient != nil {
		return stacktrace.NewError("Cannot add datastore client to network; datastore client already exists!")
	}

	if network.personModifyingApiClient != nil || network.personRetrievingApiClient != nil {
		return stacktrace.NewError("Cannot add API services to network; one or more API services already exists")
	}

	datastoreContainerConfigSupplier := network.getDatastoreContainerConfigSupplier()

	datastoreServiceContext, hostPortBindings, err := network.networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := newDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		err = datastoreClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	err = waitForHealthy(ctx, datastoreClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)

	network.datastoreClient = datastoreClient

	personModifyingApiClient, err := network.addApiService(ctx, datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-modifying API client")
	}
	network.personModifyingApiClient = personModifyingApiClient

	personRetrievingApiClient, err := network.addApiService(ctx, datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-retrieving API client")
	}
	network.personRetrievingApiClient = personRetrievingApiClient

	return returnErr
}

//  Custom network implementations will also usually have getters, to retrieve information about the
//   services created during setup
func (network *TestNetwork) GetPersonModifyingApiClient() (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, error) {
	if network.personModifyingApiClient == nil {
		return nil, stacktrace.NewError("No person-modifying API client exists")
	}
	return network.personModifyingApiClient, nil
}
func (network *TestNetwork) GetPersonRetrievingApiClient() (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, error) {
	if network.personRetrievingApiClient == nil {
		return nil, stacktrace.NewError("No person-retrieving API client exists")
	}
	return network.personRetrievingApiClient, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (network *TestNetwork) addApiService(ctx context.Context, datastoreIp string) (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, error) {

	if network.datastoreClient == nil {
		return nil, stacktrace.NewError("Cannot add API service to network; no datastore client exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	apiServiceContainerConfigSupplier := network.getApiServiceContainerConfigSupplier(datastoreIp)

	apiServiceContext, hostPortBindings, err := network.networkCtx.AddService(serviceId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, _, err := newExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", serviceId, apiServiceContext.GetIPAddress())
	}

	err = waitForHealthy(ctx, apiClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	return apiClient, nil
}

func (network *TestNetwork) getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			network.datastoreServiceImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func (network *TestNetwork) getApiServiceContainerConfigSupplier(datastoreIP string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

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
			network.apiServiceImage,
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
