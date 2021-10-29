/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package network_partition_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"time"
)

const (
	defaultPartitionId   networks.PartitionID = ""
	apiPartitionId       networks.PartitionID = "api"
	datastorePartitionId networks.PartitionID = "datastore"
	datastorePort                             = 1323
	apiServicePort                            = 2434
	datastoreServiceId   services.ServiceID   = "datastore"
	api1ServiceId        services.ServiceID   = "api1"
	api2ServiceId        services.ServiceID   = "api2"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxNumPolls       = 15

	testPersonId                          = "46"
	configFilepathRelativeToSharedDirRoot = "config-file"
)

type GRPCAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort int    `json:"datastorePort"`
}

type NetworkPartitionTest struct {
	datastoreImage string
	apiImage       string
}

func NewNetworkPartitionTest(datastoreImage string, apiImage string) *NetworkPartitionTest {
	return &NetworkPartitionTest{datastoreImage: datastoreImage, apiImage: apiImage}
}

func (test NetworkPartitionTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(
		60,
	).WithRunTimeoutSeconds(
		60,
	).WithPartitioningEnabled(true)
}

// Instantiates the network with no partition and one person in the datatstore
func (test NetworkPartitionTest) Setup(networkCtx *networks.NetworkContext) (network networks.Network, returnErr error) {
	ctx := context.Background()

	datastoreContainerConfigSupplier := test.getDatastoreContainerConfigSupplier()

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)
	datastoreClient, datastoreClientConnCloseFunc, err := newDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		err = datastoreClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	err = waitForHealthy(ctx, datastoreClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	apiClient, apiClientConnCloseFunc, err := test.addApiService(
		ctx,
		networkCtx,
		api1ServiceId,
		defaultPartitionId,
		datastoreServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding service '%v'", api1ServiceId)
	}
	defer func() {
		err = apiClientConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()

	addPersonArgs := &example_api_server_rpc_api_bindings.AddPersonArgs{
		PersonId: testPersonId,
	}
	if _, err := apiClient.AddPerson(ctx, addPersonArgs); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding test person with ID '%v'", testPersonId)
	}

	incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
		PersonId: testPersonId,
	}
	if _, err := apiClient.IncrementBooksRead(ctx, incrementBooksReadArgs); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred test person's books read in preparation for the test")
	}

	return networkCtx, returnErr
}

func (test NetworkPartitionTest) Run(network networks.Network) (returnErr error) {
	ctx := context.Background()

	// Go doesn't have generics, so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	logrus.Info("Partitioning API and datastore services off from each other...")
	if err := repartitionNetwork(castedNetwork, true, false); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the network to block access between API <-> datastore")
	}
	logrus.Info("Repartition complete")

	datastoreServiceContext, err := castedNetwork.GetServiceContext(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service context")
	}

	logrus.Info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")
	apiServiceContext, err := castedNetwork.GetServiceContext(api1ServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API 1 service context")
	}

	apiClient, _, err := newExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", api1ServiceId, apiServiceContext.GetIPAddress())
	}

	ctxWithTimeOut, cancelCtxWithTimeOut := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancelCtxWithTimeOut()

	incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
		PersonId: testPersonId,
	}
	if _, err := apiClient.IncrementBooksRead(ctxWithTimeOut, incrementBooksReadArgs); err == nil {
		return stacktrace.NewError("Expected the book increment call via API 1 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown")
	} else {
		logrus.Infof("Incrementing books read via API 1 threw the following error as expected due to network partition: %v", err)
	}

	// Adding another API service while the partition is in place ensures that partitioning works even when you add a node
	logrus.Info("Adding second API container, to ensure adding a network under partition works...")

	apiClient2, apiClient2ConnCloseFunc, err := test.addApiService(
		ctx,
		castedNetwork,
		api2ServiceId,
		apiPartitionId,
		datastoreServiceContext.GetIPAddress(),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the second API service to the network")
	}
	defer func() {
		err = apiClient2ConnCloseFunc()
		returnErr = stacktrace.Propagate(err, "An error occurred closing GRPC client")
	}()
	logrus.Info("Second API container added successfully")

	logrus.Info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")
	if _, err := apiClient2.IncrementBooksRead(ctx, incrementBooksReadArgs);err == nil {
		return stacktrace.NewError("Expected the book increment call via API 2 to fail due to the network " +
			"partition between API and datastore services, but no error was thrown")
	} else {
		logrus.Infof("Incrementing books read via API 2 threw the following error as expected due to network partition: %v", err)
	}

	// Now, open the network back up
	logrus.Info("Repartitioning to heal partition between API and datastore...")
	if err := repartitionNetwork(castedNetwork, false, true); err != nil {
		return stacktrace.Propagate(err, "An error occurred healing the partition")
	}
	logrus.Info("Partition healed successfully")

	logrus.Info("Making another call via API 1 to increment books read, to ensure the partition is open...")

	// Use infinite timeout because we expect the partition healing to fix the issue
	if _, err := apiClient.IncrementBooksRead(ctx, incrementBooksReadArgs); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 1, even though the partition should have been "+
				"healed by the goroutine",
		)
	}
	logrus.Info("Successfully incremented books read via API 1, indicating that the partition has healed successfully!")

	logrus.Info("Making another call via API 2 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	if _, err := apiClient2.IncrementBooksRead(ctx, incrementBooksReadArgs); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred incrementing the number of books read via API 2, even though the partition should have been "+
				"healed by the goroutine",
		)
	}
	logrus.Info("Successfully incremented books read via API 2, indicating that the partition has healed successfully!")
	return returnErr
}

// ========================================================================================================
//                                     Private helper functions
// ========================================================================================================
func (test NetworkPartitionTest) getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			test.datastoreImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func (test NetworkPartitionTest) getApiServiceContainerConfigSupplier(datastoreIP string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

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

func (test NetworkPartitionTest) addApiService(
	ctx context.Context,
	networkCtx *networks.NetworkContext,
	serviceId services.ServiceID,
	partitionId networks.PartitionID,
	datastoreIp string) (client example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, closeConnFunc func() error, returnErr error) {

	apiServiceContainerConfigSupplier := test.getApiServiceContainerConfigSupplier(datastoreIp)

	apiServiceContext, hostPortBindings, err := networkCtx.AddServiceToPartition(serviceId, partitionId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, apiClientConnCloseFunc, err := newExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", serviceId, apiServiceContext.GetIPAddress())
	}

	err = waitForHealthy(ctx, apiClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service '%v' with host port bindings: %+v", serviceId, hostPortBindings)
	return apiClient, apiClientConnCloseFunc, nil
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
/*
Creates a repartitioner that will partition the network between the API & datastore services, with the connection between them configurable
*/
func repartitionNetwork(
	networkCtx *networks.NetworkContext,
	isConnectionBlocked bool,
	isApi2ServiceAddedYet bool) error {
	apiPartitionServiceIds := map[services.ServiceID]bool{
		api1ServiceId: true,
	}
	if isApi2ServiceAddedYet {
		apiPartitionServiceIds[api2ServiceId] = true
	}

	partitionServices := map[networks.PartitionID]map[services.ServiceID]bool{
		apiPartitionId: apiPartitionServiceIds,
		datastorePartitionId: {
			datastoreServiceId: true,
		},
	}
	partitionConnections := map[networks.PartitionID]map[networks.PartitionID]*kurtosis_core_rpc_api_bindings.PartitionConnectionInfo{
		apiPartitionId: {
			datastorePartitionId: &kurtosis_core_rpc_api_bindings.PartitionConnectionInfo{
				IsBlocked: isConnectionBlocked,
			},
		},
	}
	defaultPartitionConnection := &kurtosis_core_rpc_api_bindings.PartitionConnectionInfo{
		IsBlocked: false,
	}
	if err := networkCtx.RepartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred repartitioning the network with isConnectionBlocked = %v",
			isConnectionBlocked)
	}
	return nil
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

