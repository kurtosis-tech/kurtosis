//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because, as of 2022-07-07, Kubernetes doesn't support network partitioning

package network_partition_starlark

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	starlarkSubnetworkTestName = "starlark-subnetwork"
	isPartitioningEnabled      = true
	executeNoDryRun            = false
	emptyArgs                  = "{}"

	serviceName1 = "service_1"
	serviceName2 = "service_2"
	subnetwork1  = "subnetwork_1"
	subnetwork2  = "subnetwork_2"
	subnetwork3  = "subnetwork_3"

	subnetworkInStarlarkScript = `DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"

SERVICE_NAME_1 = "` + serviceName1 + `"
SERVICE_NAME_2 = "` + serviceName2 + `"

SUBNETWORK_1 = "` + subnetwork1 + `"
SUBNETWORK_2 = "` + subnetwork2 + `"
SUBNETWORK_3 = "` + subnetwork3 + `"

CONNECTION_SUCCESS = 0
CONNECTION_FAILURE = 1

def run(plan, args):
	# block all connections by default
	plan.set_connection(config=kurtosis.connection.BLOCKED)

	# adding 2 services to play with, each in their own subnetwork
	service_1 = plan.add_service(
		name=SERVICE_NAME_1, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_1,
		)
	)
	service_2 = plan.add_service(
		name=SERVICE_NAME_2, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_2
		)
	)

	# Validate connection is indeed blocked
	connection_result = plan.exec(recipe=ExecRecipe(
		command=["ping", "-W", "1", "-c", "1", service_1.hostname],
	), service_name=SERVICE_NAME_2, acceptable_codes = [CONNECTION_FAILURE])

	# Allow connection between 1 and 2
	plan.set_connection(subnetworks=(SUBNETWORK_1, SUBNETWORK_2), config=kurtosis.connection.ALLOWED)

	# Connection now works
	connection_result = plan.exec(recipe=ExecRecipe(
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	), service_name=SERVICE_NAME_2, acceptable_codes = [CONNECTION_SUCCESS])

	# Reset connection to default (which is BLOCKED)
	plan.remove_connection((SUBNETWORK_1, SUBNETWORK_2))

	# Connection is back to BLOCKED
	connection_result = plan.exec(recipe=ExecRecipe(
		command=["ping", "-W", "1", "-c", "1", service_1.hostname],
	), service_name=SERVICE_NAME_2, acceptable_codes = [CONNECTION_FAILURE])

	# Create a third subnetwork connected to SUBNETWORK_1 and add service2 to it
	plan.set_connection((SUBNETWORK_3, SUBNETWORK_1), config=ConnectionConfig(packet_loss_percentage=0.0))
	plan.update_service(SERVICE_NAME_2, config=UpdateServiceConfig(subnetwork=SUBNETWORK_3))
	
	# Service 2 can now talk to Service 1 again!
	connection_result = plan.exec(recipe=ExecRecipe(
		command=["ping", "-W", "1", "-c", "1", service_1.ip_address],
	), service_name=SERVICE_NAME_2)
	plan.assert(connection_result["code"], "==", CONNECTION_SUCCESS)

	plan.print("Test successfully executed")
`
)

func TestAddServiceWithEmptyPortsAndWithoutPorts(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, starlarkSubnetworkTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	result, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, subnetworkInStarlarkScript)
	require.Nil(t, err, "Unexpected error happened executing Starlark script")

	require.Nil(t, result.InterpretationError)
	require.Empty(t, result.ValidationErrors)
	require.Nil(t, result.ExecutionError)
	require.Len(t, result.Instructions, 13)

	require.Contains(t, result.RunOutput, "Test successfully executed")
}
