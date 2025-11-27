package startosis_add_service_test

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/strings/slices"
)

const (
	addServicePortTest = `
CONTAINER_IMAGE = "kurtosistech/example-datastore-server"
SERVICE_NAME = "` + serviceName + `"
SERVICE_NAME_2 = "` + serviceName2 + `"
GRPC_PORT = 1323
SUCCESS_CODE = 0

def run(plan):
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		max_cpu=500,
		min_cpu=100,
		memory_allocation=256,
		max_memory=1024,
		min_memory=512,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP", application_protocol="grpc")
		}
	)
	datastore_1 = plan.add_service(name = SERVICE_NAME, config = config)
	plan.verify(datastore_1.ports["grpc"].url, "==", "grpc://` + serviceName + `:" + str(GRPC_PORT))
	
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		max_cpu=500,
		min_cpu=100,
		memory_allocation=256,
		max_memory=1024,
		min_memory=512,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP", url="http://foobar:9932", application_protocol="grpc")
		}
	)
	datastore_2 = plan.add_service(name = SERVICE_NAME_2, config = config)
	plan.verify(datastore_2.ports["grpc"].url, "==", "http://foobar:9932")
`
)

func (suite *StartosisAddServiceTestSuite) TestAddServicePortUrl() {
	ctx := context.Background()
	runResult, err := suite.RunScript(ctx, addServicePortTest)

	t := suite.T()

	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Service '` + serviceName + `' added with service UUID '[a-z-0-9]+'
Verification succeeded. Value is '"grpc://datastore-1:1323"'.
Service '` + serviceName2 + `' added with service UUID '[a-z-0-9]+'
Verification succeeded. Value is '"http://foobar:9932"'.
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
