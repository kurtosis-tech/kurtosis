package basic_datastore_and_api_test

import (
	"context"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "basic-datastore-and-api"
	isPartitioningEnabled = false

	datastoreServiceId services.ServiceID = "datastore"
	apiServiceId       services.ServiceID = "api"

	testPersonId     = "23"
	testNumBooksRead = uint32(3)
)

func TestBasicDatastoreAndAPITest(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()


	// ------------------------------------- TEST SETUP ----------------------------------------------
	// TODO replace with datastore launcher inside the lib
	logrus.Infof("Adding datastore service...")
	datastoreServiceCtx, _, datastoreClientCloseFunc, err := test_helpers.AddDatastoreService(ctx, datastoreServiceId, enclaveCtx)
	require.NoError(t, err, "An error occurred adding the datastore service to the enclave")
	defer datastoreClientCloseFunc()
	logrus.Infof("Added datastore service")

	logrus.Infof("Adding API service...")
	_, apiClient, apiClientCloseFunc, err := test_helpers.AddAPIService(ctx, apiServiceId, enclaveCtx, datastoreServiceCtx.GetPrivateIPAddress())
	require.NoError(t, err, "An error occurred adding the API service to the enclave")
	defer apiClientCloseFunc()
	logrus.Infof("Added API service")

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Verifying that person with test ID '%v' doesn't already exist...", testPersonId)
	getPersonArgs := &example_api_server_rpc_api_bindings.GetPersonArgs{
		PersonId: testPersonId,
	}
	 _, err = apiClient.GetPerson(ctx, getPersonArgs)
	require.Error(t, err, "Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
	logrus.Infof("Verified that test person doesn't already exist")

	logrus.Infof("Adding test person with ID '%v'...", testPersonId)
	addPersonArgs := &example_api_server_rpc_api_bindings.AddPersonArgs{
		PersonId: testPersonId,
	}
	 _, err = apiClient.AddPerson(ctx, addPersonArgs)
	require.NoError(t, err, "An error occurred adding test person with ID '%v'", testPersonId)
	logrus.Info("Test person added")

	logrus.Infof("Incrementing test person's number of books read by %v...", testNumBooksRead)
	for i := uint32(0); i < testNumBooksRead; i++ {
		incrementBooksReadArgs := &example_api_server_rpc_api_bindings.IncrementBooksReadArgs{
			PersonId: testPersonId,
		}
		_, err = apiClient.IncrementBooksRead(ctx, incrementBooksReadArgs)
		require.NoError(t, err, "An error occurred incrementing the number of books read")
	}
	logrus.Info("Incremented number of books read")

	logrus.Info("Retrieving test person to verify number of books read...")
	getPersonResponse, err := apiClient.GetPerson(ctx, getPersonArgs)
	require.NoError(t, err, "An error occurred getting the test person to verify the number of books read")
	logrus.Info("Retrieved test person")

	require.Equal(
		t,
		testNumBooksRead,
		getPersonResponse.GetBooksRead(),
		"Expected number of book read '%v' != actual number of books read '%v'",
		testNumBooksRead,
		getPersonResponse.GetBooksRead(),
	)
}
