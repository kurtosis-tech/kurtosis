package module_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/modules"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "module"
	isPartitioningEnabled = false

	testModuleImage = "kurtosistech/datastore-army-module:0.2.13"

	datastoreArmyModuleId modules.ModuleID = "datastore-army"

	numModuleExecuteCalls = 2
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
		require.NoError(
			t,
			test_helpers.ValidateDatastoreServiceHealthy(ctx, enclaveCtx, serviceId, portId),
			"Error validating datastore server '%s' is healthy",
			serviceId,
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
