package startosis_add_service_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	addServiceWithInvalidServiceName = "invalid-service-name"
	addServiceInvalidServiceNameTest = `
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
	plan.add_service(service_name = "this;.is:invalid", config = config)
`
)

func TestAddServiceWithInvalidServiceNameFailsValidation(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, addServiceWithInvalidServiceName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		destroyErr := destroyEnclaveFunc()
		if destroyErr != nil {
			logrus.Errorf("Error destroying enclave at the end of integration test '%s'",
				addServiceWithInvalidServiceName)
		}
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------

	logrus.Infof("Executing Starlark script...")
	logrus.Debugf("Starlark script contents: \n%v", addServiceInvalidServiceNameTest)

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, addServiceInvalidServiceNameTest)
	require.NoError(t, err, "Unexpected error executing Starlark script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, "Service name 'this;.is:invalid' is invalid as it contains disallowed characters. Service names can only contain characters 'a-z', 'A-Z', '0-9', '-' & '_'")
}
