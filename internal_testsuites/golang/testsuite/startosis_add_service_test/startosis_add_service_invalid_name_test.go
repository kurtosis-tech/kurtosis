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
	runResult, _ := suite.RunScript(ctx, fmt.Sprintf(addServiceInvalidServiceNameTestScript, invalidServiceName))

	t := suite.T()
	require.NotNil(t, runResult)
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected some validation errors")
	require.Contains(t, runResult.ValidationErrors[0].ErrorMessage, fmt.Sprintf("Service name '%v' is invalid as it contains disallowed characters. Service names must adhere to the RFC 1035 standard, specifically implementing this regex and be 1-63 characters long: ^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$. This means the service name must only contain lowercase alphanumeric characters or '-', and must start with a lowercase alphabet and end with a lowercase alphanumeric character.", invalidServiceName))
}
