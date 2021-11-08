package basic_datastore_test

import (
	"context"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	testName = "basic-datastore-test"

	datastoreServiceId services.ServiceID = "datastore"
	testKey                               = "test-key"
	testValue                             = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15
)

func TestBasicDatastoreTest(t *testing.T) {
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, context.Background(), testName)
	assert.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	// TODO replace with datastore launcher inside the lib
	datastoreContainerConfigSupplier := test_helpers.GetDatastoreContainerConfigSupplier()

	serviceContext, hostPortBindings, err := enclaveCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	assert.NoError(t, err, "An error occurred adding the datastore service")

	datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.NewDatastoreClient(serviceContext.GetIPAddress())
	assert.NoError(t, err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, serviceContext.GetIPAddress())
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = test_helpers.WaitForHealthy(context.Background(), datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	assert.NoError(t, err, "An error occurred waiting for the datastore service to become available")

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	existsArgs := &datastore_rpc_api_bindings.ExistsArgs{
		Key: testKey,
	}
	existsResponse, err := datastoreClient.Exists(context.Background(), existsArgs)
	assert.NoError(t, err, "An error occurred checking if the test key exists")
	assert.False(t, existsResponse.GetExists(), "Test key should not exist yet")
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testKey,
		Value: testValue,
	}
	_, err = datastoreClient.Upsert(context.Background(), upsertArgs)
	assert.NoError(t, err, "An error occurred upserting the test key")
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testKey,
	}
	getResponse, err := datastoreClient.Get(context.Background(), getArgs)
	assert.NoError(t, err, "An error occurred getting the test key after upload")
	assert.Equal(t, testValue, getResponse.GetValue(), "Returned value '%v' != test value '%v'", getResponse.GetValue(), testValue)
	logrus.Info("Value verified")
}
