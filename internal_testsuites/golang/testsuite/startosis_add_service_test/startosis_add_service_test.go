package startosis_add_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServiceWithEmptyPortsTestName = "two-service-connection-test"
	isPartitioningEnabled            = false
	defaultDryRun                    = false
	emptyArgs                        = "{}"

	serviceName  = "datastore-1"
	serviceName2 = "datastore-2"

	addServiceAndTestConnectionScript = `
CONTAINER_IMAGE = "kurtosistech/example-datastore-server"
SERVICE_NAME = "` + serviceName + `"
SERVICE_NAME_2 = "` + serviceName2 + `"
GRPC_PORT = 1323
SUCCESS_CODE = 0

def run(plan):
	plan.print("Adding services " + SERVICE_NAME + " and " + SERVICE_NAME_2)
	
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		cpu_allocation=500,
		memory_allocation=256,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP")
		}
	)
	datastore_1 = plan.add_service(name = SERVICE_NAME, config = config)
	datastore_2 = plan.add_service(name = SERVICE_NAME_2, config = config)

	test_hostname_cmd = "nc -zv {0} {1}".format(datastore_1.hostname, GRPC_PORT)
	connection_result = plan.exec(
		recipe = ExecRecipe(
			command=["sh", "-c", test_hostname_cmd],
		),
		service_name = SERVICE_NAME_2,
	)
	plan.assert(connection_result["code"], "==", SUCCESS_CODE)
	
	test_ip_address_cmd = "nc -zv {0} {1}".format(datastore_1.ip_address, GRPC_PORT) 
	connection_result = plan.exec(
		recipe = ExecRecipe(
			command=["sh", "-c", test_ip_address_cmd],
		),
		service_name = SERVICE_NAME_2,
	)
	plan.assert(connection_result["code"], "==", SUCCESS_CODE)
`
)

func TestAddTwoServicesAndTestConnection(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, addServiceWithEmptyPortsTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		destroyErr := destroyEnclaveFunc()
		if destroyErr != nil {
			logrus.Errorf("Error destroying enclave at the end of integration test '%s'",
				addServiceWithEmptyPortsTestName)
		}
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------

	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Starlark script contents: \n%v", addServiceAndTestConnectionScript)

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, addServiceAndTestConnectionScript)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Adding services ` + serviceName + ` and ` + serviceName2 + `
Service '` + serviceName + `' added with service UUID '[a-z-0-9]+'
Service '` + serviceName2 + `' added with service UUID '[a-z-0-9]+'
Command returned with exit code '0' and the following output:
--------------------
[a-z-0-9]+ \([0-9\.]+:1323\) open

--------------------
Assertion succeeded. Value is '0'.
Command returned with exit code '0' and the following output:
--------------------
[0-9\.]+ \([0-9\.]+:1323\) open

--------------------
Assertion succeeded. Value is '0'.
`
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script")

	// Ensure that the service is listed
	expectedNumberOfServices := 2
	serviceInfos, err := enclaveCtx.GetServices()
	require.Nil(t, err)
	actualNumberOfServices := len(serviceInfos)
	require.Equal(t, expectedNumberOfServices, actualNumberOfServices)
}
