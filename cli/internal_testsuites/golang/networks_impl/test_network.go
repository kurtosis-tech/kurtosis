/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package networks_impl

import (
	"context"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"strconv"
)

const (
	datastoreServiceId services.ServiceID = "datastore"

	apiServiceIdPrefix = "api-"

	waitForStartupDelayMilliseconds       = 1000
	waitForStartupMaxNumPolls             = 15
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
	enclaveCtx                         *networks.NetworkContext
	datastoreServiceImage              string
	apiServiceImage                    string
	datastoreClient                    datastore_rpc_api_bindings.DatastoreServiceClient
	personModifyingApiClient           example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient
	personRetrievingApiClient          example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient
	personModifyingApiClientCloseFunc  func() error
	personRetrievingApiClientCloseFunc func() error
	nextApiServiceId                   int
}

func NewTestNetwork(enclaveCtx *networks.NetworkContext, datastoreServiceImage string, apiServiceImage string) *TestNetwork {
	return &TestNetwork{
		enclaveCtx:                         enclaveCtx,
		datastoreServiceImage:              datastoreServiceImage,
		apiServiceImage:                    apiServiceImage,
		datastoreClient:                    nil,
		personModifyingApiClient:           nil,
		personRetrievingApiClient:          nil,
		personModifyingApiClientCloseFunc:  nil,
		personRetrievingApiClientCloseFunc: nil,
		nextApiServiceId:                   0,
	}
}

//  Custom network implementations usually have a "setup" method (possibly parameterized) that is used
//   in the Test.Setup function of each test
func (network *TestNetwork) SetupDatastoreAndTwoApis() error {
	ctx := context.Background()

	if network.datastoreClient != nil {
		return stacktrace.NewError("Cannot add datastore client to network; datastore client already exists!")
	}

	if network.personModifyingApiClient != nil || network.personRetrievingApiClient != nil {
		return stacktrace.NewError("Cannot add API services to network; one or more API services already exists")
	}

	datastoreContainerConfigSupplier := test_helpers.GetDatastoreContainerConfigSupplier(network.datastoreServiceImage)

	datastoreServiceContext, hostPortBindings, err := network.enclaveCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.NewDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			// I made this a "warn" in this case b/c there's really nothing the user can do to fix it, and it's really not a big deal if this fails
			// In other areas where it is a big deal, and I need the user to do an action, I make it an Error with a big "ACTION REQUIRED" tag
			//  (e.g. if our defer function to destroy a Docker network fails, so the user needs to delete the network else it'll hang around forever)
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = test_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)

	network.datastoreClient = datastoreClient

	personModifyingApiClient,  personModifyingApiClientCloseFunc, err := network.addApiService(ctx, datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-modifying API client")
	}
	network.personModifyingApiClient = personModifyingApiClient
	network.personModifyingApiClientCloseFunc = personModifyingApiClientCloseFunc

	personRetrievingApiClient, personRetrievingApiClientCloseFunc, err := network.addApiService(ctx, datastoreServiceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the person-retrieving API client")
	}
	network.personRetrievingApiClient = personRetrievingApiClient
	network.personRetrievingApiClientCloseFunc = personRetrievingApiClientCloseFunc

	return nil
}

//  Custom network implementations will also usually have getters, to retrieve information about the
//   services created during setup
func (network *TestNetwork) GetPersonModifyingApiClient() (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {
	if network.personModifyingApiClient == nil {
		return nil, nil, stacktrace.NewError("No person-modifying API client exists")
	}
	return network.personModifyingApiClient, network.personModifyingApiClientCloseFunc, nil
}
func (network *TestNetwork) GetPersonRetrievingApiClient() (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {
	if network.personRetrievingApiClient == nil {
		return nil, nil, stacktrace.NewError("No person-retrieving API client exists")
	}
	return network.personRetrievingApiClient, network.personRetrievingApiClientCloseFunc, nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (network *TestNetwork) addApiService(ctx context.Context, datastoreIp string) (example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func() error, error) {

	if network.datastoreClient == nil {
		return nil, nil, stacktrace.NewError("Cannot add API service to network; no datastore client exists")
	}

	serviceIdStr := apiServiceIdPrefix + strconv.Itoa(network.nextApiServiceId)
	network.nextApiServiceId = network.nextApiServiceId + 1
	serviceId := services.ServiceID(serviceIdStr)

	apiServiceContainerConfigSupplier := test_helpers.GetApiServiceContainerConfigSupplier(network.apiServiceImage, datastoreIp)

	apiServiceContext, hostPortBindings, err := network.enclaveCtx.AddService(serviceId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, apiClientCloseFunc, err := test_helpers.NewExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", serviceId, apiServiceContext.GetIPAddress())
	}

	err = test_helpers.WaitForHealthy(ctx, apiClient, waitForStartupMaxNumPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", hostPortBindings)
	return apiClient, apiClientCloseFunc, nil
}

