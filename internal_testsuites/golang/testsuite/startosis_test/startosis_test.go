package startosis_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

const (
	testName              = "module"
	isPartitioningEnabled = false

	scriptFilePath = "../../static_files/startosis_valid_script.star"

	serviceIdPlaceholder = "[!SERVICE_ID_PLACEHOLDER]"
	serviceId            = "example-datastore-server-1"
	portIdPlaceholder    = "[!PORT_ID_PLACEHOLDER]"
	portId               = "grpc"
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	rawStartosisScript, err := os.ReadFile(scriptFilePath)
	require.NoError(t, err, "Error reading Startosis script file")
	renderedStartosisScript := injectAllScriptPlaceholders(string(rawStartosisScript))

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script '%s'...", scriptFilePath)
	logrus.Debugf("Startosis script content: \n%v", renderedStartosisScript)

	executionResult, err := enclaveCtx.ExecuteStartosisScript(renderedStartosisScript)
	require.NoError(t, err, "Unexpected error executing startosis script '%s'", scriptFilePath)

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service example-datastore-server-1 deployed successfully.
`
	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, executionResult.ValidationError, "Unexpected validation error")
	require.Empty(t, executionResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Infof("Successfully ran Startosis script '%s'", scriptFilePath)

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is healthy",
		serviceId,
	)
	logrus.Infof("All services added via the module work as expected")
}

func injectAllScriptPlaceholders(scriptStr string) string {
	// TODO: this is a hack to be able to pass dynamic param to the Startosis script. Once we have a first class way
	//  to do this, remove this bit
	renderedScript := injectPlaceholderString(scriptStr, serviceIdPlaceholder, serviceId)
	renderedScript = injectPlaceholderString(renderedScript, portIdPlaceholder, portId)
	return renderedScript
}

func injectPlaceholderString(src string, placeholderKey string, value string) string {
	if !strings.Contains(src, placeholderKey) {
		logrus.Warnf("Placeholder '%s' not found in source script. Test will continue but this might be a signal for a bug in the test", placeholderKey)
	}
	return strings.ReplaceAll(src, placeholderKey, fmt.Sprintf("\"%s\"", value))
}
