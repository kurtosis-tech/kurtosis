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

	serviceName  = "docker-getting-started"
	serviceName2 = "docker-getting-started-2"

	starlarkScriptWithEmptyPorts = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_NAME = "` + serviceName + `"

def run(plan):
	plan.print("Adding service " + SERVICE_NAME + ".")
	
	config = ServiceConfig(
		image = DOCKER_GETTING_STARTED_IMAGE,
		cpu_allocation = 500,
		memory_allocation = 512,
	)
	
	plan.add_service(service_name = SERVICE_NAME, config = config)
	plan.print("Service " + SERVICE_NAME + " deployed successfully.")
`

	starlarkScriptWithoutPorts = `
DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"
SERVICE_NAME = "` + serviceName2 + `"

def run(plan, args):
	plan.print("Adding service " + SERVICE_NAME + ".")
	
	config = ServiceConfig(
		image = DOCKER_GETTING_STARTED_IMAGE,
	)
	
	plan.add_service(service_name = SERVICE_NAME, config = config)
	plan.print("Service " + SERVICE_NAME + " deployed successfully.")
`
)

var serviceNames = []string{serviceName, serviceName2}
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

		runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, emptyArgs, defaultDryRun)
		require.NoError(t, err, "Unexpected error executing starlark script")

		expectedScriptOutput := `Adding service ` + serviceNames[starlarkScripIndex] + `.
Service '` + serviceNames[starlarkScripIndex] + `' added with service UUID '[a-z-0-9]+'
Service ` + serviceNames[starlarkScripIndex] + ` deployed successfully.
`
		require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
		require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
		require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
		require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
		logrus.Infof("Successfully ran Starlark script")

		// Ensure that the service is listed
		expectedNumberOfServices := starlarkScripIndex + 1
		serviceInfos, err := enclaveCtx.GetServices()
		require.Nil(t, err)
		actualNumberOfServices := len(serviceInfos)
		require.Equal(t, expectedNumberOfServices, actualNumberOfServices)
	}
}
