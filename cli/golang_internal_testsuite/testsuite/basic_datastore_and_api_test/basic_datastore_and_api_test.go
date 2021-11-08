package basic_datastore_and_api_test

import (
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	apiServiceId       services.ServiceID = "api"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15

	testPersonId     = "23"
	testNumBooksRead = 3
)

func TestBasicDatastoreAndAPITest(t *testing.T) {
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, context.Background(), testName)
	assert.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()


	// ------------------------------------- TEST SETUP ----------------------------------------------

	datastoreContainerConfigSupplier := test_helpers.GetDatastoreContainerConfigSupplier()

	datastoreServiceContext, datastoreSvcHostPortBindings, err := networkCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.NewDatastoreClient(datastoreServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, datastoreServiceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = test_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", datastoreSvcHostPortBindings)

	apiServiceContainerConfigSupplier := test_helpers.GetApiServiceContainerConfigSupplier(test.apiImage, datastoreServiceContext.GetIPAddress())

	apiServiceContext, apiSvcHostPortBindings, err := networkCtx.AddService(apiServiceId, apiServiceContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	apiClient, apiClientConnCloseFunc, err := test_helpers.NewExampleAPIServerClient(apiServiceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", apiServiceId, apiServiceContext.GetIPAddress())
	}
	defer func() {
		if err := apiClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the API client, but doing so threw an error:\n%v", err)
		}
	}()

	err = test_helpers.WaitForHealthy(ctx, apiClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the example API server service to become available")
	}

	logrus.Infof("Added API service with host port bindings: %+v", apiSvcHostPortBindings)
	return networkCtx, nil
}

func (test BasicDatastoreAndApiTest) Run(network networks.Network) error {
	ctx := context.Background()

	// Go doesn't have generics, so we have to do this cast first
	castedNetwork := network.(*networks.NetworkContext)

	serviceContext, err := castedNetwork.GetServiceContext(apiServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the API service context")
	}

	apiClient, apiClientConnCloseFunc, err := test_helpers.NewExampleAPIServerClient(serviceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new example API server client for service with ID '%v' and IP address '%v'", apiServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		if err := apiClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the API client, but doing so threw an error:\n%v", err)
		}
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

	if getPersonResponse.GetBooksRead() != testNumBooksRead {
		return stacktrace.NewError(
			"Expected number of book read '%v' != actual number of books read '%v'",
			testNumBooksRead,
			getPersonResponse.GetBooksRead(),
		)
	}

	return nil
}
