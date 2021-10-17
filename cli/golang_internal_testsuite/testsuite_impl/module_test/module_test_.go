/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_test

import (
	"encoding/json"
	"github.com/kurtosis-tech/example-microservice/datastore/datastore_service_client"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	testModuleImage = "kurtosistech/datastore-army-module:0.1.2"

	datastoreArmyModuleId modules.ModuleID = "datastore-army"

	numModuleExecuteCalls = 2

	testDatastoreKey = "my-key"
	testDatastoreValue = "test-value"
)

type ModuleTest struct {}

type DatastoreArmyModuleResult struct {
	CreatedServiceIdPorts map[string]uint32 `json:"createdServiceIdPorts"`
}

func (test ModuleTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test ModuleTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	logrus.Info("Loading module...")
	_, err := networkCtx.LoadModule(datastoreArmyModuleId, testModuleImage, "{}")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore army module")
	}
	logrus.Info("Module loaded successfully")
	return networkCtx, nil
}

func (test ModuleTest) Run(rawNetwork networks.Network) error {
	// Because Go doesn't have generics
	networkCtx, ok := rawNetwork.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred casting the generic network to a NetworkContext type")
	}

	moduleCtx, err := networkCtx.GetModuleContext(datastoreArmyModuleId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the context for module '%v'", datastoreArmyModuleId)
	}

	allNewServiceIdsAndPorts := map[services.ServiceID]uint32{}
	for i := 0; i < numModuleExecuteCalls; i++ {
		logrus.Info("Adding two datastore services via the module...")
		newServiceIdsAndPorts, err := addTwoDatastoreServices(moduleCtx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred adding two datastore services via the module")
		}
		for serviceId, portNum := range newServiceIdsAndPorts {
			if _, found := allNewServiceIdsAndPorts[serviceId]; found {
				return stacktrace.NewError("The module created services with IDs that already exist!")
			}
			allNewServiceIdsAndPorts[serviceId] = portNum
		}
		logrus.Info("Successfully added two datastore services via the module")
	}

	// Sanity-check that the datastore services that the module created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the module work as expected...", len(allNewServiceIdsAndPorts))
	for serviceId, portNum := range allNewServiceIdsAndPorts {
		serviceCtx, err := networkCtx.GetServiceContext(serviceId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the service context for service '%v'; this indicates that the module says it created a service that it actually didn't", serviceId)
		}
		ipAddr := serviceCtx.GetIPAddress()
		datastoreClient := datastore_service_client.NewDatastoreClient(ipAddr, int(portNum))
		if err := datastoreClient.Upsert(testDatastoreKey, testDatastoreValue); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding the test key to datastore service '%v'", serviceId)
		}
		actualValue, err := datastoreClient.Get(testDatastoreKey)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the test key from datastore service '%v'", serviceId)
		}
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
	if err := networkCtx.UnloadModule(datastoreArmyModuleId); err != nil {
		return stacktrace.Propagate(err, "An error occurred unloading module '%v'", datastoreArmyModuleId)
	}

	if _, err := networkCtx.GetModuleContext(datastoreArmyModuleId); err == nil {
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

	result := map[services.ServiceID]uint32{}
	for createdServiceIdStr, createdServicePortNum := range parsedResult.CreatedServiceIdPorts {
		result[services.ServiceID(createdServiceIdStr)] = createdServicePortNum
	}
	return result, nil
}

