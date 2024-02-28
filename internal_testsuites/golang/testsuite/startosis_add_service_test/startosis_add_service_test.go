package startosis_add_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/strings/slices"
)

const (
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
		max_cpu=500,
		min_cpu=100,
		memory_allocation=256,
		max_memory=1024,
		min_memory=512,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP")
		}
	)
	datastore_1 = plan.add_service(name = SERVICE_NAME, config = config)
	datastore_2 = plan.add_service(name = SERVICE_NAME_2, config = config)

	ds1_through_get = plan.get_service(SERVICE_NAME)
	plan.print(ds1_through_get)

	test_hostname_cmd = "nc -zv {0} {1}".format(datastore_1.hostname, GRPC_PORT)
	connection_result = plan.exec(
		recipe = ExecRecipe(
			command=["sh", "-c", test_hostname_cmd],
		),
		service_name = SERVICE_NAME_2,
	)
	plan.verify(connection_result["code"], "==", SUCCESS_CODE)
	
	test_ip_address_cmd = "nc -zv {0} {1}".format(datastore_1.ip_address, GRPC_PORT) 
	connection_result = plan.exec(
		recipe = ExecRecipe(
			command=["sh", "-c", test_ip_address_cmd],
		),
		service_name = SERVICE_NAME_2,
	)
	plan.verify(connection_result["code"], "==", SUCCESS_CODE)
`
)

func (suite *StartosisAddServiceTestSuite) TestAddTwoServicesAndTestConnection() {
	ctx := context.Background()
	runResult, err := suite.RunScript(ctx, addServiceAndTestConnectionScript)

	t := suite.T()

	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Adding services ` + serviceName + ` and ` + serviceName2 + `
Service '` + serviceName + `' added with service UUID '[a-z-0-9]+'
Service '` + serviceName2 + `' added with service UUID '[a-z-0-9]+'
Fetched service '` + "" + `'
Service\(name="datastore-1", hostname="datastore-1", ip_address="[0-9\.]+", ports=\{"grpc": PortSpec\(number=1323, transport_protocol="TCP", wait="2m0s"\)\}\)
Command returned with exit code '0' and the following output:
--------------------
[a-z-0-9]+ \([0-9\.]+:1323\) open

--------------------
Verification succeeded. Value is '0'.
Command returned with exit code '0' and the following output:
--------------------
[0-9\.]+ \([0-9\.]+:1323\) open

--------------------
Verification succeeded. Value is '0'.
`
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script")

	// Ensure that the service is listed
	expectedNumberOfServices := 2
	serviceInfos, err := suite.enclaveCtx.GetServices()
	require.Nil(t, err)

	serviceNames := []string{serviceName, serviceName2}
	startedServices := []services.ServiceName{}
	for userServiceName := range serviceInfos {
		if slices.Contains(serviceNames, string(userServiceName)) {
			startedServices = append(startedServices, userServiceName)
		}
	}
	actualNumberOfServices := len(startedServices)
	require.Equal(t, expectedNumberOfServices, actualNumberOfServices)
}
