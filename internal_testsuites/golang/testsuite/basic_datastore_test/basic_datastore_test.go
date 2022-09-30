package basic_datastore_test

import (
	"context"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "basic-datastore"
	isPartitioningEnabled = false

	datastoreServiceId services.ServiceID = "datastore"
	testKey                               = "test-key"
	testValue                             = "test-value"
)

func TestBasicDatastoreTest(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	// TODO replace with datastore launcher inside the lib
	logrus.Infof("Adding datastore service...")
	_, datastoreClient, clientCloseFunc, err := test_helpers.AddDatastoreService(ctx, datastoreServiceId, enclaveCtx)
	require.NoError(t, err, "An error occurred adding the datastore service to the enclave")
	defer clientCloseFunc()
	logrus.Infof("Added datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	existsArgs := &datastore_rpc_api_bindings.ExistsArgs{
		Key: testKey,
	}
	existsResponse, err := datastoreClient.Exists(context.Background(), existsArgs)
	require.NoError(t, err, "An error occurred checking if the test key exists")
	require.False(t, existsResponse.GetExists(), "Test key should not exist yet")
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testKey,
		Value: testValue,
	}
	_, err = datastoreClient.Upsert(context.Background(), upsertArgs)
	require.NoError(t, err, "An error occurred upserting the test key")
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testKey,
	}
	getResponse, err := datastoreClient.Get(context.Background(), getArgs)
	require.NoError(t, err, "An error occurred getting the test key after upload")
	require.Equal(t, testValue, getResponse.GetValue(), "Returned value '%v' != test value '%v'", getResponse.GetValue(), testValue)
	logrus.Info("Value verified")
}
