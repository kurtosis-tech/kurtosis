/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package module_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/client_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/modules"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	testModuleImage = "kurtosistech/datastore-army-module:0.1.5"

	datastoreArmyModuleId modules.ModuleID = "datastore-army"

	numModuleExecuteCalls = 2

	testDatastoreKey = "my-key"
	testDatastoreValue = "test-value"

	waitForStartupDelayMilliseconds = 1000
	waitForStartupMaxPolls          = 15
)

type ModuleTest struct {}

type DatastoreArmyModuleResult struct {
	CreatedServiceIdPorts map[string]uint32 `json:"createdServiceIdPorts"`
}

func (test ModuleTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(90)
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
	ctx := context.Background()

	// Because Go doesn't have generics
	networkCtx, ok := rawNetwork.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred casting the generic network to a NetworkContext type")
	}

	moduleCtx, err := networkCtx.GetModuleContext(datastoreArmyModuleId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the context for module '%v'", datastoreArmyModuleId)
	}
	serviceIdList := []services.ServiceID{}
	for i := 0; i < numModuleExecuteCalls; i++ {
		logrus.Info("Adding two datastore services via the module...")
		serviceIdList, err = addTwoDatastoreServices(moduleCtx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred adding two datastore services via the module")
		}
		logrus.Info("Successfully added two datastore services via the module")
	}

	// Sanity-check that the datastore services that the module created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the module work as expected...", len(serviceIdList))
	for _, serviceId := range serviceIdList {
		serviceCtx, err := networkCtx.GetServiceContext(serviceId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the service context for service '%v'; this indicates that the module says it created a service that it actually didn't", serviceId)
		}
		ipAddr := serviceCtx.GetIPAddress()

		datastoreClient, datastoreClientConnCloseFunc, err := client_helpers.NewDatastoreClient(ipAddr)
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
	if err := networkCtx.UnloadModule(datastoreArmyModuleId); err != nil {
		return stacktrace.Propagate(err, "An error occurred unloading module '%v'", datastoreArmyModuleId)
	}

	if _, err := networkCtx.GetModuleContext(datastoreArmyModuleId); err == nil {
		return stacktrace.Propagate(err, "Getting module context for module '%v' should throw an error because it should had been unloaded", datastoreArmyModuleId)
	}
	logrus.Infof("Module '%v' successfully unloaded", datastoreArmyModuleId)

	return nil
}

func addTwoDatastoreServices(moduleCtx *modules.ModuleContext) ([]services.ServiceID, error) {
	paramsJsonStr := `{"numDatastores": 2}`
	respJsonStr, err := moduleCtx.Execute(paramsJsonStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the datastore army module")
	}

	parsedResult := new(DatastoreArmyModuleResult)
	if err := json.Unmarshal([]byte(respJsonStr), parsedResult); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the module response")
	}

	result := []services.ServiceID{}
	for createdServiceIdStr, _ := range parsedResult.CreatedServiceIdPorts {
		result = append(result, services.ServiceID(createdServiceIdStr))
	}
	return result, nil
}
