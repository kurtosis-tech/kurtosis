package startosis_add_service_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
)

const (
	invalidServiceName = "this;.is:invalid"

	addServiceInvalidServiceNameTestScript = `
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

func (suite *StartosisAddServiceTestSuite) TestAddServiceWithInvalidServiceNameFailsValidation() {
	ctx := context.Background()
	runResult, _ := suite.RunScript(ctx, addServiceInvalidServiceNameTestScript)

	t := suite.T()

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected some validation errors")
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, fmt.Sprintf("Service name '%s' is invalid as it contains disallowed characters. Service names can only contain characters 'a-z', 'A-Z', '0-9', '-' & '_'", invalidServiceName))
}
