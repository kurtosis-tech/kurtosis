package module_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "module"
	isPartitioningEnabled = false

	testModuleImage = "kurtosistech/datastore-army-module:0.2.2"

	datastoreArmyModuleId modules.ModuleID = "datastore-army"

	numModuleExecuteCalls = 2

	testDatastoreKey = "my-key"
	testDatastoreValue = "test-value"

	millisBetweenAvailabilityRetries = 1000

	/*
	NOTE: on 2022-05-16 this failed with the following error so we bumped the num polls to 20. If this fails again, look
	into if there's some sort of nondeterminism happening.

	time="2022-05-16T23:58:21Z" level=info msg="Sanity-checking that all 4 datastore services added via the module work as expected..."
	--- FAIL: TestModule (21.46s)
	    module_test.go:81:
	        	Error Trace:	module_test.go:81
	        	Error:      	Received unexpected error:
	        	            	The service didn't return a success code, even after 15 retries with 1000 milliseconds in between retries
	        	            	 --- at /home/circleci/project/internal_testsuites/golang/test_helpers/test_helpers.go:179 (WaitForHealthy) ---
	        	            	Caused by: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:49188: connect: connection refused"
	        	Test:       	TestModule
	        	Messages:   	An error occurred waiting for the datastore service to become available
	 */
	waitForStartupMaxPolls           = 20
)

type DatastoreArmyModuleResult struct {
	CreatedServiceIdsToPortIds map[string]string `json:"createdServiceIdsToPortIds"`
}

func TestModule(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()


	// ------------------------------------- TEST SETUP ----------------------------------------------
	logrus.Info("Loading module...")
	moduleCtx, err := enclaveCtx.LoadModule(datastoreArmyModuleId, testModuleImage, "{}")
	require.NoError(t, err, "An error occurred adding the datastore army module")
	logrus.Info("Module loaded successfully")

	// ------------------------------------- TEST RUN ----------------------------------------------
	serviceIdsToPortIds := map[services.ServiceID]string{}
	for i := 0; i < numModuleExecuteCalls; i++ {
		logrus.Info("Adding two datastore services via the module...")
		createdServiceIdsToPortIds, err := addTwoDatastoreServices(moduleCtx)
		require.NoError(t, err, "An error occurred adding two datastore services via the module")
		for serviceId, portId := range createdServiceIdsToPortIds {
			serviceIdsToPortIds[serviceId] = portId
		}
		logrus.Info("Successfully added two datastore services via the module")
	}

	// Sanity-check that the datastore services that the module created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the module work as expected...", len(serviceIdsToPortIds))
	for serviceId, portId := range serviceIdsToPortIds {
		serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
		require.NoError(t, err, "An error occurred getting the service context for service '%v'; this indicates that the module says it created a service that it actually didn't", serviceId)
		ipAddr := serviceCtx.GetMaybePublicIPAddress()

		publicPort, found := serviceCtx.GetPublicPorts()[portId]
		require.True(t, found, "Expected to find public port '%v' on datastore service '%v', but none was found", portId, serviceId)

		datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.CreateDatastoreClient(
			ipAddr,
			publicPort.GetNumber(),
		)
		require.NoError(t, err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", serviceId, ipAddr)
		defer datastoreClientConnCloseFunc()

		require.NoError(
			t,
			test_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, millisBetweenAvailabilityRetries),
			"An error occurred waiting for datastore service '%v' to become available",
			serviceId,
		)

		upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
			Key:   testDatastoreKey,
			Value: testDatastoreValue,
		}
		_, err = datastoreClient.Upsert(ctx, upsertArgs)
		require.NoError(t, err, "An error occurred adding the test key to datastore service '%v'", serviceId)

		getArgs := &datastore_rpc_api_bindings.GetArgs{
			Key: testDatastoreKey,
		}
		getResponse, err := datastoreClient.Get(ctx, getArgs)
		require.NoError(t, err, "An error occurred getting the test key from datastore service '%v'", serviceId)

		actualValue := getResponse.GetValue()
		require.Equal(
			t,
			testDatastoreValue,
			actualValue,
			"Datastore service '%v' is storing value '%v' for the test key, which doesn't match the expected value '%v'",
			serviceId,
			actualValue,
			testDatastoreValue,
		)
	}
	logrus.Info("All services added via the module work as expected")

	logrus.Infof("Unloading module '%v'...", datastoreArmyModuleId)
	require.NoError(
		t,
		enclaveCtx.UnloadModule(datastoreArmyModuleId),
		"An error occurred unloading module '%v'",
		datastoreArmyModuleId,
	)

	_, err = enclaveCtx.GetModuleContext(datastoreArmyModuleId)
	require.Error(
		t,
		err,
		"Getting module context for module '%v' should throw an error because it should had been unloaded",
		datastoreArmyModuleId,
	)
	logrus.Infof("Module '%v' successfully unloaded", datastoreArmyModuleId)
}

func addTwoDatastoreServices(moduleCtx *modules.ModuleContext) (map[services.ServiceID]string, error) {
	paramsJsonStr := `{"numDatastores": 2}`
	respJsonStr, err := moduleCtx.Execute(paramsJsonStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the datastore army module")
	}

	parsedResult := new(DatastoreArmyModuleResult)
	if err := json.Unmarshal([]byte(respJsonStr), parsedResult); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the module response")
	}

	result := map[services.ServiceID]string{}
	for createdServiceIdStr, createdServicePortId := range parsedResult.CreatedServiceIdsToPortIds {
		result[services.ServiceID(createdServiceIdStr)] = createdServicePortId
	}
	return result, nil
}

