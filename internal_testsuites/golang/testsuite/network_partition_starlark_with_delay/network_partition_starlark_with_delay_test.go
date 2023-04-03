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

	serviceId1  = "service_1"
	serviceId2  = "service_2"
	serviceId3  = "service_3"
	subnetwork1 = "subnetwork_1"

	subnetworkInStarlarkScript = `DOCKER_GETTING_STARTED_IMAGE = "docker/getting-started:latest"

SERVICE_ID_1 = "` + serviceId1 + `"
SERVICE_ID_2 = "` + serviceId2 + `"
SERVICE_ID_3 = "` + serviceId3 + `"

SUBNETWORK_1 = "` + subnetwork1 + `"

CONNECTION_SUCCESS = 0
CONNECTION_FAILURE = 1

def run(plan, args):
	plan.set_connection(config=kurtosis.connection.ALLOWED)

	# adding 2 services to play with, each in their own subnetwork
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
			subnetwork=SUBNETWORK_1
		)
	)

	service_3 = plan.add_service(
		name=SERVICE_ID_3, 
		config=ServiceConfig(
			image=DOCKER_GETTING_STARTED_IMAGE,
		)
	)
	
	# the ping command below with returns the avg latency (rtt) for the 5 packets sent
	service_one_cmd =  "ping -c 5 -W 5 " + service_1.ip_address +  " | tail -1| awk '{print $4}' | cut -d '/' -f 2"
	service_three_cmd =  "ping -c 5 -W 5 " + service_3.ip_address +  " | tail -1| awk '{print $4}' | cut -d '/' -f 2"

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_one_cmd],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], "<", "2")

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_three_cmd],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], "<", "2")

	delay = UniformPacketDelayDistribution(750)
	plan.set_connection(config=ConnectionConfig(packet_delay_distribution=delay))
	
	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_one_cmd],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], "<", "2")

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_three_cmd],
	)
	
	# this is doing string comparison
	# have not found a way to convert output to int
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], ">", "1449")
    
	uniform_delay_distribution = UniformPacketDelayDistribution(ms=350)	
	plan.set_connection(config=ConnectionConfig(packet_delay_distribution=uniform_delay_distribution))

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_one_cmd],
	)
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], "<", "2")

	recipe = ExecRecipe(
		command=["/bin/sh", "-c", service_three_cmd],
	)
	
	# this is doing string comparison
	# have not found a way to convert output to int
	# the overall latency should be greater than 350*2, but
	# added some buffer to handle 50ms outliers
	res = plan.exec(recipe=recipe, service_name=SERVICE_ID_2)
	plan.assert(res["output"], ">", "649")
	plan.print("Test successfully executed")
`
)

func TestNetworkParitionWithSomeDelay(t *testing.T) {
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
	require.Len(t, result.Instructions, 19)

	require.Contains(t, result.RunOutput, "Test successfully executed")
}
