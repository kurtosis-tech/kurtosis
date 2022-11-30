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
	defaultDryRun                    = false
	emptyArgs                        = "{}"

	serviceId  = "docker-getting-started"
	serviceId2 = "docker-getting-started-2"

	starlarkScriptWithEmptyPorts = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "` + serviceId + `"

def run(args):
	print("Adding service " + SERVICE_ID + ".")
	
	config = struct(
		image = DOCKER_GETTING_STARTED_IMAGE,
		ports = {}
	)
	
	add_service(service_id = SERVICE_ID, config = config)
	print("Service " + SERVICE_ID + " deployed successfully.")
`

	starlarkScriptWithoutPorts = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_ID = "` + serviceId2 + `"

print("Adding service " + SERVICE_ID + ".")

config = struct(
    image = DOCKER_GETTING_STARTED_IMAGE,
)

add_service(service_id = SERVICE_ID, config = config)
print("Service " + SERVICE_ID + " deployed successfully.")
`
)

var serviceIds = []string{serviceId, serviceId2}
var starlarkScriptsToRun = []string{starlarkScriptWithEmptyPorts, starlarkScriptWithoutPorts}

func TestAddServiceWithEmptyPortsAndWithoutPorts(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, addServiceWithEmptyPortsTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------

	for starlarkScripIndex, starlarkScript := range starlarkScriptsToRun {
		logrus.Infof("Executing Starlark script...")
		logrus.Debugf("Starlark script content: \n%v", starlarkScript)

		outputStream, _, err := enclaveCtx.RunStarlarkScript(ctx, starlarkScript, emptyArgs, defaultDryRun)
		require.NoError(t, err, "Unexpected error executing starlark script")
		scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

		expectedScriptOutput := `Adding service ` + serviceIds[starlarkScripIndex] + `.
Service '` + serviceIds[starlarkScripIndex] + `' added with internal ID '[a-z-0-9]+'
Service ` + serviceIds[starlarkScripIndex] + ` deployed successfully.
`
		require.Nil(t, interpretationError, "Unexpected interpretation error.")
		require.Empty(t, validationErrors, "Unexpected validation error")
		require.Nil(t, executionError, "Unexpected execution error")
		require.Regexp(t, expectedScriptOutput, scriptOutput)
		logrus.Infof("Successfully ran Starlark script")

		// Ensure that the service is listed
		expectedNumberOfServices := starlarkScripIndex + 1
		serviceInfos, err := enclaveCtx.GetServices()
		require.Nil(t, err)
		actualNumberOfServices := len(serviceInfos)
		require.Equal(t, expectedNumberOfServices, actualNumberOfServices)
	}
}
