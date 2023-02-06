package startosis_add_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServicesTestName = "add-services-test"

	addServicesScript = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_NAME_PREFIX = "service-"
NUM_SERVICES = 4

def run(plan):
    plan.print("Adding {0} services to enclave".format(NUM_SERVICES))
    
    config = ServiceConfig(
        image = DOCKER_GETTING_STARTED_IMAGE,
    )

    configs = {}
    for index in range(NUM_SERVICES):
        service_name = SERVICE_NAME_PREFIX + str(index)
        configs[service_name] = config
    
    plan.add_services(configs)
`
)

func TestAddServices(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, addServicesTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Starlark script content: \n%v", addServicesScript)

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, addServicesScript)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	expectedScriptOutput := `Adding 4 services to enclave
Successfully added the following '4' services:
  Service 'service-[0-9]' added with UUID '[a-f0-9]{32}'
  Service 'service-[0-9]' added with UUID '[a-f0-9]{32}'
  Service 'service-[0-9]' added with UUID '[a-f0-9]{32}'
  Service 'service-[0-9]' added with UUID '[a-f0-9]{32}'
`
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Starlark script")

	// Ensure that the service is listed
	expectedNumberOfServices := 4
	serviceInfos, err := enclaveCtx.GetServices()
	require.Nil(t, err)
	actualNumberOfServices := len(serviceInfos)
	require.Equal(t, expectedNumberOfServices, actualNumberOfServices)

}
