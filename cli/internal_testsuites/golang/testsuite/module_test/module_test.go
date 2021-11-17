package module_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "module-test"
	isPartitioningEnabled = false

	testModuleImage = "kurtosistech/datastore-army-module:0.1.5"

	datastoreArmyModuleId modules.ModuleID = "datastore-army"

	numModuleExecuteCalls = 2

	testDatastoreKey = "my-key"
	testDatastoreValue = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15
)

type DatastoreArmyModuleResult struct {
	CreatedServiceIdPorts map[string]uint32 `json:"createdServiceIdPorts"`
}

func TestModule(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()


	// ------------------------------------- TEST SETUP ----------------------------------------------
	logrus.Info("Loading module...")
	// TODO TODO TODO Reeanble this!!!!! We currently don't have a way to get the host machine port bindings for
	//  a service started by the module, so this is dependent on https://github.com/kurtosis-tech/kurtosis-core/issues/460
	// moduleCtx, err := enclaveCtx.LoadModule(datastoreArmyModuleId, testModuleImage, "{}")
	_, err = enclaveCtx.LoadModule(datastoreArmyModuleId, testModuleImage, "{}")
	require.NoError(t, err, "An error occurred adding the datastore army module")
	logrus.Info("Module loaded successfully")

	// ------------------------------------- TEST RUN ----------------------------------------------

	// TODO TODO TODO Reeanble this!!!!! We currently don't have a way to get the host machine port bindings for
	//  a service started by the module, so this is dependent on https://github.com/kurtosis-tech/kurtosis-core/issues/460
	/*
	serviceIdsToPortUint := map[services.ServiceID]uint32{}
	for i := 0; i < numModuleExecuteCalls; i++ {
		logrus.Info("Adding two datastore services via the module...")
		createdServiceIdsAndPorts, err := addTwoDatastoreServices(moduleCtx)
		require.NoError(t, err, "An error occurred adding two datastore services via the module")
		for serviceId, port := range createdServiceIdsAndPorts {
			serviceIdsToPortUint[serviceId] = port
		}
		logrus.Info("Successfully added two datastore services via the module")
	}

	// Sanity-check that the datastore services that the module created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the module work as expected...", len(serviceIdsToPortUint))
	for serviceId, portUint := range serviceIdsToPortUint {
		serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
		require.NoError(t, err, "An error occurred getting the service context for service '%v'; this indicates that the module says it created a service that it actually didn't", serviceId)
		ipAddr := serviceCtx.GetIPAddress()

		datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.CreateDatastoreClient(
			ipAddr,
			fmt.Sprintf("%v", portUint),
		)
		require.NoError(t, err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", serviceId, ipAddr)
		defer datastoreClientConnCloseFunc()

		require.NoError(
			t,
			test_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds),
			"An error occurred waiting for the datastore service to become available",
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
	 */

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

func addTwoDatastoreServices(moduleCtx *modules.ModuleContext) (map[services.ServiceID]uint32, error) {
	paramsJsonStr := `{"numDatastores": 2}`
	respJsonStr, err := moduleCtx.Execute(paramsJsonStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the datastore army module")
	}

	parsedResult := new(DatastoreArmyModuleResult)
	if err := json.Unmarshal([]byte(respJsonStr), parsedResult); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the module response")
	}

	result := map[services.ServiceID]uint32{}
	for createdServiceIdStr, createdServiceIdPortNum := range parsedResult.CreatedServiceIdPorts {
		result[services.ServiceID(createdServiceIdStr)] = createdServiceIdPortNum
	}
	return result, nil
}

