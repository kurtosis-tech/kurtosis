package startosis_add_service_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	invalidServiceName                       = "this;.is:invalid"
	addServiceWithInvalidServiceNameTestName = "add-service-invalid-service-name"
	addServiceInvalidServiceNameTestScript   = `
CONTAINER_IMAGE = "kurtosistech/example-datastore-server"
GRPC_PORT = 1323

def run(plan):
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		cpu_allocation=500,
		memory_allocation=256,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP")
		}
	)
	plan.add_service(name = "%s", config = config)
`
)

func TestAddServiceWithInvalidServiceNameFailsValidation(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, addServiceWithInvalidServiceNameTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		destroyErr := destroyEnclaveFunc()
		if destroyErr != nil {
			logrus.Errorf("Error destroying enclave at the end of integration test '%s'",
				addServiceWithInvalidServiceNameTestName)
		}
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------

	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Starlark script contents: \n%v", fmt.Sprintf(addServiceInvalidServiceNameTestScript, invalidServiceName))

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, fmt.Sprintf(addServiceInvalidServiceNameTestScript, invalidServiceName))
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected some validation errors")
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, fmt.Sprintf("Service name '%s' is invalid as it contains disallowed characters. Service names can only contain characters 'a-z', 'A-Z', '0-9', '-' & '_'", invalidServiceName))
}
