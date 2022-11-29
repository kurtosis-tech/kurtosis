package startosis_add_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServiceWithEmptyPortsTestName = "add-service-empty-ports"
	isPartitioningEnabled            = false
	defaultDryRun         = false

	serviceId = "docker-getting-started"

	startosisScript = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "` + serviceId + `"

print("Adding service " + SERVICE_ID + ".")

config = struct(
    image = DOCKER_GETTING_STARTED_IMAGE,
	ports = {}
)

add_service(service_id = SERVICE_ID, config = config)
print("Service " + SERVICE_ID + " deployed successfully.")
`
)

func TestAddServiceWithEmptyPorts(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, addServiceWithEmptyPortsTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Starlark script content: \n%v", startosisScript)

	outputStream, _, err := enclaveCtx.ExecuteKurtosisScript(ctx, startosisScript, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing starlark script")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Adding service docker-getting-started.
Service docker-getting-started deployed successfully.
`
	require.Nil(t, interpretationError, "Unexpected interpretation error.")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, test_helpers.GenerateScriptOutput(instructions))
	logrus.Infof("Successfully ran Starlark script")

	// Ensure that the service is listed
	expectedAmountOfServices := 1
	serviceIds, err := enclaveCtx.GetServices()
	require.Nil(t, err)
	actualAmountOfServices := len(serviceIds)
	require.Equal(t, expectedAmountOfServices, actualAmountOfServices)
}
