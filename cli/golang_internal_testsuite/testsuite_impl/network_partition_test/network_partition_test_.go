/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package network_partition_test

import (
	"context"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	defaultPartitionId   networks.PartitionID = ""
	apiPartitionId       networks.PartitionID = "api"
	datastorePartitionId networks.PartitionID = "datastore"

	datastoreServiceId   services.ServiceID   = "datastore"
	api1ServiceId        services.ServiceID   = "api1"
	api2ServiceId        services.ServiceID   = "api2"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxNumPolls       = 15

	testPersonId                          = "46"
	contextTimeOut = 2 * time.Second
)

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
func (test NetworkPartitionTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	ctx := context.Background()

	datastoreContainerConfigSupplier := test_helpers.GetDatastoreContainerConfigSupplier(test.datastoreImage)

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)
	datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.NewDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = test_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
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
		if err := apiClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the API client, but doing so threw an error:\n%v", err)
		}
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

	return networkCtx, nil
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

	apiClient, _, err := test_helpers.NewExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", api1ServiceId, apiServiceContext.GetIPAddress())
	}

	ctxWithTimeOut, cancelCtxWithTimeOut := context.WithTimeout(ctx, contextTimeOut)
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
		if err := apiClient2ConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the API client 2, but doing so threw an error:\n%v", err)
		}
	}()
	logrus.Info("Second API container added successfully")

	ctxWithTimeOut2, cancelCtxWithTimeOut2 := context.WithTimeout(ctx, contextTimeOut)
	defer cancelCtxWithTimeOut2()

	logrus.Info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")
	if _, err := apiClient2.IncrementBooksRead(ctxWithTimeOut2, incrementBooksReadArgs);err == nil {
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
	return nil
}

// ========================================================================================================
//                                     Private helper functions
// ========================================================================================================
func (test NetworkPartitionTest) addApiService(
	ctx context.Context,
	networkCtx *networks.NetworkContext,
	serviceId services.ServiceID,
	partitionId networks.PartitionID,
	datastoreIp string) (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {

	apiServiceContainerConfigSupplier := test_helpers.GetApiServiceContainerConfigSupplier(test.apiImage, datastoreIp)

	apiServiceContext, hostPortBindings, err := networkCtx.AddServiceToPartition(serviceId, partitionId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, apiClientConnCloseFunc, err := test_helpers.NewExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", serviceId, apiServiceContext.GetIPAddress())
	}

	err = test_helpers.WaitForHealthy(ctx, apiClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service '%v' with host port bindings: %+v", serviceId, hostPortBindings)
	return apiClient, apiClientConnCloseFunc, nil
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
