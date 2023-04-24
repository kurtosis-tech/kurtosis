package check_that_apic_network_is_private

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "apic-is-private"
	isPartitioningEnabled = false
	defaultDryRun         = false
	defaultParallelism    = 4
	emptyArgs             = "{}"

	// This Script fails if the APIC has a non private IP address range
	starlarkScript = `
WEAVIATE_IMAGE="semitechnologies/weaviate:1.18.3"

WEAVIATE_PORT = 8080
WEAVIATE_PORT_ID = "http"

def run(plan, args):

    plan.add_service(
        name = "weaviate",
        config = ServiceConfig(
            image = WEAVIATE_IMAGE,
            ports = {
                WEAVIATE_PORT_ID: PortSpec(number = WEAVIATE_PORT, transport_protocol = "TCP")
            },
            env_vars = {
                "QUERY_DEFAULTS_LIMIT": str(25),
                "AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED": 'true',
                "PERSISTENCE_DATA_PATH": '/var/lib/weaviate',
                "DEFAULT_VECTORIZER_MODULE": 'none',
                "CLUSTER_HOSTNAME": 'node1'
            },
        )
    )
    
    # sleep for a bit so that the container can start an die in case of non private ip
    plan.exec(
        service_name = "weaviate",
        recipe = ExecRecipe(
            command = ["sleep", "5"]
        )
    )

    # check the port
    plan.request(
        service_name = "weaviate",
        recipe = GetHttpRequestRecipe(
            port_id = WEAVIATE_PORT_ID,
            endpoint = "/",	
        ),
        acceptable_codes = [200],
    )
`
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", starlarkScript)

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, emptyArgs, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing startosis script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")
}
