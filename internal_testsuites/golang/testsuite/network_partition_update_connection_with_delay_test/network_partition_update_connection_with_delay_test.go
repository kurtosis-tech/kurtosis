//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because, as of 2022-07-07, Kubernetes doesn't support network partitioning

package network_partition_update_connection_with_delay_test

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

	serviceId1                 = "service_1"
	serviceId2                 = "service_2"
	serviceId3                 = "service_3"
	subnetwork1                = "subnetwork_1"
	subnetwork2                = "subnetwork2"
	subnetwork3                = "subnetwork3"
	subnetworkInStarlarkScript = `DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"

SERVICE_ID_1 = "` + serviceId1 + `"
SERVICE_ID_2 = "` + serviceId2 + `"
SERVICE_ID_3 = "` + serviceId3 + `"

SUBNETWORK_1 = "` + subnetwork1 + `"
SUBNETWORK_2 = "` + subnetwork2 + `"
SUBNETWORK_3 = "` + subnetwork3 + `"

CONNECTION_SUCCESS = 0
CONNECTION_FAILURE = 1

def run(plan, args):
	plan.set_connection(config=kurtosis.connection.BLOCKED)
	
	service_1 = plan.add_service(
		name=SERVICE_ID_1, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_1,
		)
	)

	service_2 = plan.add_service(
		name=SERVICE_ID_2, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
			subnetwork=SUBNETWORK_2
		)
	)

	service_one_cmd =  "ping -c 5 -W 5 " + service_1.ip_address +  " | tail -1| awk '{print $4}' | cut -d '/' -f 2"

	# blocked connection
	recipe = ExecRecipe(
		command=["ping", "-c", "1", "-W", "1", service_1.ip_address],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2, acceptable_codes=[1])

	
	# unblock connection with some delay 
	delay = UniformPacketDelayDistribution(ms=750)
	
	# minimum delay should ideally be 140*2 = 280
	# maximum delay should ideally be 160*2 = 320
	normal_delay_distribution = NormalPacketDelayDistribution(mean_ms=150, std_dev_ms=10, correlation=2.0)

    # this should pick normal delay distribution instead of delay
	plan.set_connection(
		(SUBNETWORK_1, SUBNETWORK_2), 
		config=ConnectionConfig(
			packet_delay_distribution=normal_delay_distribution
		)
	)

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_one_cmd],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)

	# given a buffer of 55ms to handle outliers
	# minimum latency observed (280-55 = 225)
	# maximum latency observed (280+55 = 375)
	plan.assert(res["output"], "<", "375")
	plan.assert(res["output"], ">", "225")

	# remove connection, should block the connection again
	plan.remove_connection((SUBNETWORK_1, SUBNETWORK_2))
	recipe = ExecRecipe(
		command=["ping", "-c", "1", "-W", "1", service_1.ip_address],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2, acceptable_codes=[1])

	plan.set_connection((SUBNETWORK_1, SUBNETWORK_3), config=ConnectionConfig(packet_delay_distribution=delay))
	plan.update_service(SERVICE_ID_2, config=UpdateServiceConfig(subnetwork=SUBNETWORK_3))
	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_one_cmd],
	)

	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], ">", "1449")
	plan.print("Test successfully executed")
`
)

func TestNetworkPartitionWithSomeDelay(t *testing.T) {
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
	require.Len(t, result.Instructions, 15)

	require.Contains(t, result.RunOutput, "Test successfully executed")
}
