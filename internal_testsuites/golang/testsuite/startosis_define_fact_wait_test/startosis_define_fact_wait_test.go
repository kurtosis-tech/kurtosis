package startosis_define_fact_wait_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "startosis_define_fact_wait_test"
	isPartitioningEnabled = false
	defaultDryRun         = false

	serviceId          = "http-echo"
	portId             = "http"
	expectedGetOutput  = "get-result"
	expectedPostOutput = "post-result"

	startosisScript = `
IMAGE = "mendhak/http-https-echo:26"
SERVICE_ID = "` + serviceId + `"
PORT_ID = "` + portId + `"
PORT_NUMBER = 8080
PORT_PROTOCOL = "TCP"
GET_ENDPOINT = "?service=` + expectedGetOutput + `"
GET_FACT_NAME = "get-fact"
POST_ENDPOINT = "/"
POST_BODY = "` + expectedPostOutput + `"
POST_FACT_NAME = "post-fact"

config = struct(
    image = IMAGE,
    ports = {
        PORT_ID: struct(number = PORT_NUMBER, protocol = PORT_PROTOCOL)
    }
)

add_service(service_id = SERVICE_ID, config = config)
print("Service deployed successfully.")

define_fact(service_id = SERVICE_ID, fact_name = GET_FACT_NAME, fact_recipe=struct(method="GET", endpoint=GET_ENDPOINT, port_id=PORT_ID, field_extractor=".query.service"))
get_fact = wait(service_id=SERVICE_ID, fact_name=GET_FACT_NAME)

add_service(service_id = get_fact, config = config)
print("Service dependency 1 deployed successfully.")

define_fact(service_id = SERVICE_ID, fact_name = POST_FACT_NAME, fact_recipe=struct(method="POST", endpoint=POST_ENDPOINT, port_id=PORT_ID, field_extractor=".body", content_type="text/plain", body=POST_BODY))
post_fact = wait(service_id=SERVICE_ID, fact_name=POST_FACT_NAME)

add_service(service_id = post_fact, config = config)
print("Service dependency 2 deployed successfully.")
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
	logrus.Debugf("Startosis script content: \n%v", startosisScript)

	outputStream, _, err := enclaveCtx.ExecuteKurtosisScript(ctx, startosisScript, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis script")
	interpretationError, validationErrors, executionError, instructions := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Service deployed successfully.
Service dependency 1 deployed successfully.
Service dependency 2 deployed successfully.
`
	require.Nil(t, interpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, test_helpers.GenerateScriptOutput(instructions))
	logrus.Infof("Successfully ran Startosis script")

	serviceInfos, err := enclaveCtx.GetServices()
	require.Equal(t, 3, len(serviceInfos))
	require.Contains(t, serviceInfos, services.ServiceID(expectedGetOutput))
	require.Contains(t, serviceInfos, services.ServiceID(expectedPostOutput))
	require.Nil(t, err)
}
