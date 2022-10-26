package stream_logs_test

import (
	"bufio"
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

const (
	testName              = "stream-logs"
	isPartitioningEnabled = false

	dockerGettingStartedImage                       = "docker/getting-started"
	exampleServiceId             services.ServiceID = "stream-logs"

	exampleServicePortId                            = "http"
	exampleServicePrivatePortNum                    = 80

	lineBreakRune = '\n'
	lineBreakStr = "\n"

	waitForAllLogsBeingCollectedInSeconds = 2
)

var exampleServicePrivatePortSpec = services.NewPortSpec(
	exampleServicePrivatePortNum,
	services.PortProtocol_TCP,
)
func TestStreamLogs(t *testing.T) {
	ctx := context.Background()

	expectedLogLines := []string{"kurtosis", "test", "running", "successfully"}
	expectedAmountOfLogLines := 4

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfig := getExampleServiceConfig()
	_, err = enclaveCtx.AddService(exampleServiceId, containerConfig)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	enclaveID := enclaveCtx.GetEnclaveID()

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	serviceCtx, err := enclaveCtx.GetServiceContext(exampleServiceId)
	require.NoError(t, err)

	userServiceGuid := serviceCtx.GetServiceGUID()

	userServiceGUIDs := map[services.ServiceGUID]bool{
		userServiceGuid: true,
	}

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	userServiceReadCloserLogs, err := kurtosisCtx.StreamUserServiceLogs(ctx, enclaveID, userServiceGUIDs)
	require.NoError(t, err)

	userServiceReadCloserLog, found := userServiceReadCloserLogs[userServiceGuid]
	require.True(t, found)

	buffReader := bufio.NewReader(userServiceReadCloserLog)

	receivedLogLines := []string{}

	for  {
		logLine, err := buffReader.ReadString(lineBreakRune)
		longLineWithoutLineBreak := strings.ReplaceAll(logLine, lineBreakStr, "")
		require.NoError(t, err)
		receivedLogLines = append(receivedLogLines, longLineWithoutLineBreak)
		if len(receivedLogLines) == expectedAmountOfLogLines {
			break
		}
	}
	require.Equal(t, expectedLogLines, receivedLogLines)

}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getExampleServiceConfig() *services.ContainerConfig {

	entrypointArgs := []string{"/bin/sh", "-c"}
	cmdArgs := []string{"for i in kurtosis test running successfully; do echo \"$i\"; if [ \"$i\" == \"successfully\" ]; then sleep 300; fi; done;"}

	containerConfig := services.NewContainerConfigBuilder(
		dockerGettingStartedImage,
	).WithEntrypointOverride(
		entrypointArgs,
	).WithCmdOverride(
		cmdArgs,
	).WithUsedPorts(map[string]*services.PortSpec{
		exampleServicePortId: exampleServicePrivatePortSpec,
	}).Build()
	return containerConfig
}
