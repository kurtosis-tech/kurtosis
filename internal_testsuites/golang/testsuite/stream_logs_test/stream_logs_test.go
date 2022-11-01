//+build !minikube

// We don't run this test in Kubernetes because, as of 2022-10-28, the centralized logs feature is not implemented in Kubernetes yet

package stream_logs_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/stretchr/testify/require"
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

	userServiceLogsByGuidChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.StreamUserServiceLogs(ctx, enclaveID, userServiceGUIDs)
	require.NoError(t, err)
	require.NotNil(t, cancelStreamUserServiceLogsFunc)
	require.NotNil(t, userServiceLogsByGuidChan)
	defer cancelStreamUserServiceLogsFunc()

	receivedLogLines := []string{}

	for  {
		userServiceLogsByGuid, isChanOpen := <-userServiceLogsByGuidChan
		require.True(t, isChanOpen)

		userServiceLogs, found := userServiceLogsByGuid[userServiceGuid]
		require.True(t, found)

		for _, serviceLog := range userServiceLogs {
			receivedLogLines = append(receivedLogLines, serviceLog.GetContent())
		}

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
