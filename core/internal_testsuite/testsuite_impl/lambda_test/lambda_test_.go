package lambda_test

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
	testLambdaImage = "kurtosistech/datastore-army-lambda"

	datastoreArmyLambdaId modules.LambdaID = "datastore-army"

	numLambdaCalls = 2

	// TODO THIS SHOULD COME FROM THE LAMBDA ITSELF!!!
	datastoreContainerPort = 1323

	testDatastoreKey = "my-key"
	testDatastoreValue = "test-value"
)

type LambdaTest struct {}

type DatastoreArmyLambdaResult struct {
	CreatedServiceIdPorts map[string]uint32 `json:"createdServiceIdPorts"`
}

func (l LambdaTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (l LambdaTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	logrus.Info("Loading lambda...")
	_, err := networkCtx.LoadLambda(datastoreArmyLambdaId, testLambdaImage, "{}")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the datastore army Lambda")
	}
	logrus.Info("Lambda loaded successfully")
	return networkCtx, nil
}

func (l LambdaTest) Run(rawNetwork networks.Network) error {
	// Because Go doesn't have generics
	networkCtx, ok := rawNetwork.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred casting the generic network to a NetworkContext type")
	}

	lambdaCtx, err := networkCtx.GetLambdaContext(datastoreArmyLambdaId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the context for Lambda '%v'", datastoreArmyLambdaId)
	}

	createdServiceIds := map[services.ServiceID]bool{}
	for i := 0; i < numLambdaCalls; i++ {
		logrus.Info("Adding two datastore services via the Lambda...")
		newServiceIds, err := addTwoDatastoreServices(lambdaCtx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred adding two datastore services via the Lambda")
		}
		for serviceId := range newServiceIds {
			if _, found := createdServiceIds[serviceId]; found {
				return stacktrace.NewError("The Lambda created services with IDs that already exist!")
			}
			createdServiceIds[serviceId] = true
		}
		logrus.Info("Successfully added two datastore services via the Lambda")
	}

	// Sanity-check that the datastore services that the Lambda created are functional
	logrus.Infof("Sanity-checking that all %v datastore services added via the Lambda work as expected...", len(createdServiceIds))
	for serviceId := range createdServiceIds {
		serviceCtx, err := networkCtx.GetServiceContext(serviceId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting the service context for service '%v'; this indicates that the Lambda says it created a service that it actually didn't", serviceId)
		}
		ipAddr := serviceCtx.GetIPAddress()
		datastoreClient := datastore_service_client.NewDatastoreClient(ipAddr, datastoreContainerPort)
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
	logrus.Info("All services added via the Lambda work as expected")
	return nil
}

func addTwoDatastoreServices(lambdaCtx *modules.LambdaContext) (map[services.ServiceID]bool, error) {
	paramsJsonStr := `{"numDatastores": 2}`
	respJsonStr, err := lambdaCtx.Execute(paramsJsonStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the datastore army Lambda")
	}

	parsedResult := new(DatastoreArmyLambdaResult)
	if err := json.Unmarshal([]byte(respJsonStr), parsedResult); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing the Lambda response")
	}

	result := map[services.ServiceID]bool{}
	for createdServiceIdStr := range parsedResult.CreatedServiceIdPorts {
		result[services.ServiceID(createdServiceIdStr)] = true
	}
	return result, nil
}

