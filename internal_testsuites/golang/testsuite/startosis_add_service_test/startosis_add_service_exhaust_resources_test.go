//go:build !kubernetes
// +build !kubernetes

// We don't run this on Kubernetes as we don't have resource calculation available there

package startosis_add_service_test

import (
	"context"
	"github.com/stretchr/testify/require"
)

const (
	addServiceMemoryValidationTest = `
CONTAINER_IMAGE = "kurtosistech/example-datastore-server"
SERVICE_NAME = "` + serviceName + `"
GRPC_PORT = 1323
SUCCESS_CODE = 0

def run(plan):
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		max_cpu=500,
		min_cpu=100,
		memory_allocation=256,
		min_memory=51200000,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP")
		}
	)
	datastore_1 = plan.add_service(name = SERVICE_NAME, config = config)`

	addServiceCPUValidationTest = `
CONTAINER_IMAGE = "kurtosistech/example-datastore-server"
SERVICE_NAME = "` + serviceName + `"
GRPC_PORT = 1323
SUCCESS_CODE = 0

def run(plan):
	config = ServiceConfig(
		image = CONTAINER_IMAGE,
		min_cpu=1000000,
		memory_allocation=256,
		min_memory=512,
		ports = {
			"grpc": PortSpec(number = GRPC_PORT, transport_protocol = "TCP")
		}
	)
	datastore_1 = plan.add_service(name = SERVICE_NAME, config = config)`
)

func (suite *StartosisAddServiceTestSuite) TestAddServices_FailsIfWeConsumeMoreThanAvailableMemory() {
	ctx := context.Background()
	runResult, err := suite.RunScript(ctx, addServiceMemoryValidationTest)

	t := suite.T()

	require.NotNil(t, err)
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation errors to be non empty")
	require.Contains(t, runResult.ValidationErrors[0].GetErrorMessage(), "service 'datastore-1' requires '51200000' megabytes of memory")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
}

func (suite *StartosisAddServiceTestSuite) TestAddServices_FailsIfWeConsumeMoreThanAvailableCPU() {
	ctx := context.Background()
	runResult, err := suite.RunScript(ctx, addServiceCPUValidationTest)

	t := suite.T()

	require.NotNil(t, err)
	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error.")
	require.NotEmpty(t, runResult.ValidationErrors, "Expected validation errors to be non empty")
	require.Contains(t, runResult.ValidationErrors[0].GetErrorMessage(), "service 'datastore-1' requires '1000000' millicores of cpu")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
}
