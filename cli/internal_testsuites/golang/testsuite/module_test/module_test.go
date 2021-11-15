package module_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
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
	moduleCtx, err := enclaveCtx.LoadModule(datastoreArmyModuleId, testModuleImage, "{}")
	require.NoError(t, err, "An error occurred adding the datastore army module")
	logrus.Info("Module loaded successfully")

	serviceIdsToPorts := map[services.ServiceID]uint32{}
	for i := 0; i < numModuleExecuteCalls; i++ {
		logrus.Info("Adding two datastore services via the module...")
		createdServiceIdsAndPorts, err := addTwoDatastoreServices(moduleCtx)
		require.NoError(t, err, "An error occurred adding two datastore services via the module")
		for serviceId, port := range createdServiceIdsAndPorts {
			serviceIdsToPorts[serviceId] = port
		}
		logrus.Info("Successfully added two datastore services via the module")
	}

	// Sanity-check that the datastore services that the module created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the module work as expected...", len(serviceIdsToPorts))
	for serviceId, port := range serviceIdsToPorts {
		serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
		require.NoError(t, err, "An error occurred getting the service context for service '%v'; this indicates that the module says it created a service that it actually didn't", serviceId)
		ipAddr := serviceCtx.GetIPAddress()

		datastoreClient, datastoreClientConnCloseFunc, err := test_helpers.CreateDatastoreClient(ipAddr, )
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", serviceId, ipAddr)
		}
		defer func() {
			if err := datastoreClientConnCloseFunc(); err != nil {
				logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
			}
		}()

		err = client_helpers.WaitForHealthy(ctx, datastoreClient, waitForStartupMaxPolls, waitForStartupDelayMilliseconds)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
		}

		upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
			Key:   testDatastoreKey,
			Value: testDatastoreValue,
		}
		if _, err := datastoreClient.Upsert(ctx, upsertArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding the test key to datastore service '%v'", serviceId)
		}

		getArgs := &datastore_rpc_api_bindings.GetArgs{
			Key: testDatastoreKey,
		}
		getResponse, err := datastoreClient.Get(ctx, getArgs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the test key from datastore service '%v'", serviceId)
		}
		actualValue := getResponse.GetValue()
		if actualValue != testDatastoreValue {
			return stacktrace.NewError(
				"Datastore service '%v' is storing value '%v' for the test key, which doesn't match the expected value '%v'",
				serviceId,
				actualValue,
				testDatastoreValue,
			)
		}
	}
	logrus.Info("All services added via the module work as expected")

	logrus.Infof("Unloading module '%v'...", datastoreArmyModuleId)
	if err := enclaveCtx.UnloadModule(datastoreArmyModuleId); err != nil {
		return stacktrace.Propagate(err, "An error occurred unloading module '%v'", datastoreArmyModuleId)
	}

	if _, err := enclaveCtx.GetModuleContext(datastoreArmyModuleId); err == nil {
		return stacktrace.Propagate(err, "Getting module context for module '%v' should throw an error because it should had been unloaded", datastoreArmyModuleId)
	}
	logrus.Infof("Module '%v' successfully unloaded", datastoreArmyModuleId)

	return nil
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

	result := map[services.ServiceID]uint32
	for createdServiceIdStr, createdServiceIdPortNum := range parsedResult.CreatedServiceIdPorts {
		result[services.ServiceID(createdServiceIdStr)] = createdServiceIdPortNum
	}
	return result, nil
}

