//+build !minikube

// We don't run this test in Kubernetes because, as of 2022-07-07, Kubernetes doesn't support network partitioning

package network_partition_test

import (
	"context"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testName              = "network-partition"
	isPartitioningEnabled = true

	apiPartitionId       enclaves.PartitionID = "api"
	datastorePartitionId enclaves.PartitionID = "datastore"

	datastoreServiceId services.ServiceID = "datastore"
	api1ServiceId      services.ServiceID = "api1"
	api2ServiceId      services.ServiceID = "api2"

	testPersonId   = "46"
	contextTimeOut = 2 * time.Second
)

func TestNetworkPartition(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------

	datastoreServiceCtx, _, datastoreClientCloseFunc, err := test_helpers.AddDatastoreService(ctx, datastoreServiceId, enclaveCtx)
	require.NoError(t, err, "An error occurred adding the datastore service")
	defer datastoreClientCloseFunc()

	_, api1Client, api1ClientCloseFunc, err := test_helpers.AddAPIService(ctx, api1ServiceId, enclaveCtx, datastoreServiceCtx.GetPrivateIPAddress())
	require.NoError(t, err, "An error occurred adding the first API service")
	defer api1ClientCloseFunc()

	addPersonArgs := &example_api_server_rpc_api_bindings.AddPersonArgs{
		PersonId: testPersonId,
	}
	_, err = api1Client.AddPerson(ctx, addPersonArgs)
	require.NoError(t, err, "An error occurred adding test person with ID '%v'", testPersonId)

	incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
		PersonId: testPersonId,
	}
	_, err = api1Client.IncrementBooksRead(ctx, incrementBooksReadArgs)
	require.NoError(t, err, "An error occurred test person's books read in preparation for the test")

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Partitioning API and datastore services off from each other...")
	require.NoError(t, repartitionNetwork(enclaveCtx, true, false), "An error occurred repartitioning the network to block access between API <-> datastore")
	logrus.Info("Repartition complete")

	logrus.Info("Incrementing books read via API 1 while partition is in place, to verify no comms are possible...")
	ctxWithTimeOut, cancelCtxWithTimeOut := context.WithTimeout(ctx, contextTimeOut)
	defer cancelCtxWithTimeOut()
	_, err = api1Client.IncrementBooksRead(ctxWithTimeOut, incrementBooksReadArgs)
	require.Error(t, err, "Expected the book increment call via API 1 to fail due to the network "+
		"partition between API and datastore services, but no error was thrown")
	logrus.Infof("Incrementing books read via API 1 threw the following error as expected due to network partition: %v", err)

	// Adding another API service while the partition is in place ensures that partitioning works even when you add a node
	logrus.Info("Adding second API container, to ensure adding a service under partition works...")
	_, apiClient2, apiClient2ConnCloseFunc, err := test_helpers.AddAPIServiceToPartition(
		ctx,
		api2ServiceId,
		enclaveCtx,
		datastoreServiceCtx.GetPrivateIPAddress(),
		apiPartitionId,
	)
	require.NoError(t, err, "An error occurred adding the second API service to the network")
	defer apiClient2ConnCloseFunc()
	logrus.Info("Second API container added successfully")

	logrus.Info("Incrementing books read via API 2 while partition is in place, to verify no comms are possible...")
	ctxWithTimeOut2, cancelCtxWithTimeOut2 := context.WithTimeout(ctx, contextTimeOut)
	defer cancelCtxWithTimeOut2()
	_, err = apiClient2.IncrementBooksRead(ctxWithTimeOut2, incrementBooksReadArgs)
	require.Error(t, err, "Expected the book increment call via API 2 to fail due to the network "+
		"partition between API and datastore services, but no error was thrown")
	logrus.Infof("Incrementing books read via API 2 threw the following error as expected due to network partition: %v", err)

	// Now, open the network back up
	logrus.Info("Repartitioning to heal partition between API and datastore...")
	require.NoError(t, repartitionNetwork(enclaveCtx, false, true), "An error occurred healing the partition")
	logrus.Info("Partition healed successfully")

	logrus.Info("Making another call via API 1 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	_, err = api1Client.IncrementBooksRead(ctx, incrementBooksReadArgs)
	require.NoError(t, err, "An error occurred incrementing the number of books read via API 1, even though the partition should have been "+
		"healed by the goroutine")
	logrus.Info("Successfully incremented books read via API 1, indicating that the partition has healed successfully!")

	logrus.Info("Making another call via API 2 to increment books read, to ensure the partition is open...")
	// Use infinite timeout because we expect the partition healing to fix the issue
	_, err = apiClient2.IncrementBooksRead(ctx, incrementBooksReadArgs)
	require.NoError(t, err, "An error occurred incrementing the number of books read via API 2, even though the partition should have been "+
		"healed by the goroutine")
	logrus.Info("Successfully incremented books read via API 2, indicating that the partition has healed successfully!")
}

/*
Creates a repartitioner that will partition the network between the API & datastore services, with the connection between them configurable
*/
func repartitionNetwork(
	enclaveCtx *enclaves.EnclaveContext,
	isConnectionBlocked bool,
	isApi2ServiceAddedYet bool,
) error {
	apiPartitionServiceIds := map[services.ServiceID]bool{
		api1ServiceId: true,
	}
	if isApi2ServiceAddedYet {
		apiPartitionServiceIds[api2ServiceId] = true
	}

	var connectionBetweenPartitions enclaves.PartitionConnection
	if isConnectionBlocked {
		connectionBetweenPartitions = enclaves.NewBlockedPartitionConnection()
	} else {
		connectionBetweenPartitions = enclaves.NewUnblockedPartitionConnection()
	}
	partitionServices := map[enclaves.PartitionID]map[services.ServiceID]bool{
		apiPartitionId: apiPartitionServiceIds,
		datastorePartitionId: {
			datastoreServiceId: true,
		},
	}
	partitionConnections := map[enclaves.PartitionID]map[enclaves.PartitionID]enclaves.PartitionConnection{
		apiPartitionId: {
			datastorePartitionId: connectionBetweenPartitions,
		},
	}
	defaultPartitionConnection := enclaves.NewUnblockedPartitionConnection()
	if err := enclaveCtx.RepartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred repartitioning the network with isConnectionBlocked = %v",
			isConnectionBlocked,
		)
	}
	return nil
}
