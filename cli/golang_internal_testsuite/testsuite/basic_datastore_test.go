package testsuite

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/client_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const (
	datastoreServiceId services.ServiceID = "datastore"
	testKey                               = "test-key"
	testValue                             = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15

	datastoreImage = "kurtosistech/example-datastore-server"
)

func TestBasicDatastoreTest(t *testing.T) {
	// ------------------------------------- TEST SETUP ----------------------------------------------
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	assert.NoError(t, err, "An error occurred connecting to the Kurtosis engine for running the test")
	enclaveId := enclaves.EnclaveID(fmt.Sprintf(
		"%v_engine-server-test_basic-datastore-test",
		time.Now().Second(),
	))
	enclaveCtx, err := kurtosisCtx.CreateEnclave(context.Background(), enclaveId, false)
	assert.NoError(t, err, "An error occurred creating enclave '%v'", enclaveId)
	defer func() {
		if err := kurtosisCtx.DestroyEnclave(context.Background(), enclaveId); err != nil {
			logrus.Error("An error occurred destroying enclave '%v' that we created for this test:\n%v", err)
			logrus.Error("ACTION REQUIRED: You'll need to delete enclave '%v' manually!!!!")
		}
	}()

	// TODO replace with datastore launcher inside the lib
	datastoreContainerConfigSupplier := client_helpers.GetDatastoreContainerConfigSupplier(datastoreImage)

	serviceContext, hostPortBindings, err := enclaveCtx.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := client_helpers.NewDatastoreClient(serviceContext.GetIPAddress())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	err = client_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}

	logrus.Infof("Added datastore service with host port bindings: %+v", hostPortBindings)
	return networkCtx, nil
}

func (test BasicDatastoreTest) Run(network networks.Network) error {
	ctx := context.Background()

	// Necessary because Go doesn't have generics
	castedNetwork := network.(*networks.NetworkContext)

	serviceContext, err := castedNetwork.GetServiceContext(datastoreServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the datastore service info")
	}

	datastoreClient, datastoreClientConnCloseFunc, err := client_helpers.NewDatastoreClient(serviceContext.GetIPAddress())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", datastoreServiceId, serviceContext.GetIPAddress())
	}
	defer func() {
		if err := datastoreClientConnCloseFunc(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}()

	logrus.Infof("Verifying that key '%v' doesn't already exist...", testKey)
	existsArgs := &datastore_rpc_api_bindings.ExistsArgs{
		Key: testKey,
	}
	existsResponse, err := datastoreClient.Exists(ctx, existsArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if the test key exists")
	}
	if existsResponse.GetExists() {
		return stacktrace.NewError("Test key should not exist yet")
	}
	logrus.Infof("Confirmed that key '%v' doesn't already exist", testKey)

	logrus.Infof("Inserting value '%v' at key '%v'...", testKey, testValue)
	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testKey,
		Value: testValue,
	}
	if _, err = datastoreClient.Upsert(ctx, upsertArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred upserting the test key")
	}
	logrus.Infof("Inserted value successfully")

	logrus.Infof("Getting the key we just inserted to verify the value...")
	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testKey,
	}
	getResponse, err := datastoreClient.Get(ctx, getArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key after upload")
	}
	if getResponse.GetValue() != testValue {
		return stacktrace.NewError("Returned value '%v' != test value '%v'", getResponse.GetValue(), testValue)
	}
	logrus.Info("Value verified")
	return nil
}
