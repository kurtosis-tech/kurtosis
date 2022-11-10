//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because, as of 2022-10-28, the centralized logs feature is not implemented in Kubernetes yet

package stream_logs_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testName              = "stream-logs"
	isPartitioningEnabled = false

	dockerGettingStartedImage                    = "docker/getting-started"
	exampleServiceId          services.ServiceID = "stream-logs"


	waitForAllLogsBeingCollectedInSeconds = 2

	testTimeOut = 90 * time.Second

	shouldFollowLogs  = true
	shouldNotFollowLogs = false

	nonExistentUserServiceGuid = "stream-logs-1667939326"
)

type getUserServiceLogsRequestInfoAndExpectedResults struct {
	requestedEnclaveID enclaves.EnclaveID
	requestedUserServiceGuids map[services.ServiceGUID]bool
	requestedFollowLogs bool
	expectedLogLines []string
	expectedNotFoundServiceGuids map[services.ServiceGUID]bool
}

func TestStreamLogs(t *testing.T) {
	ctx := context.Background()

	expectedLogLineValues := []string{"kurtosis", "test", "running", "successfully"}
	expectedEmptyLogLineValues := []string{}

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

	serviceCtx, err := enclaveCtx.GetServiceContext(exampleServiceId)
	require.NoError(t, err)

	userServiceGuid := serviceCtx.GetServiceGUID()

	userServiceGuids := map[services.ServiceGUID]bool{
		userServiceGuid: true,
	}

	emptyUserServiceGuids := map[services.ServiceGUID]bool{}

	nonExistentUserServiceGuids := map[services.ServiceGUID]bool{
		nonExistentUserServiceGuid: true,
	}

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	firstCallRequestInfoAndExpectedResults := &getUserServiceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveID,
		requestedUserServiceGuids:    userServiceGuids,
		requestedFollowLogs:          shouldNotFollowLogs,
		expectedLogLines:             expectedLogLineValues,
		expectedNotFoundServiceGuids: emptyUserServiceGuids,

	}

	secondCallRequestInfoAndExpectedResults := &getUserServiceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveID,
		requestedUserServiceGuids:    userServiceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             expectedLogLineValues,
		expectedNotFoundServiceGuids: emptyUserServiceGuids,
	}

	thirdCallRequestInfoAndExpectedResults := &getUserServiceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID: enclaveID,
		requestedUserServiceGuids: nonExistentUserServiceGuids,
		requestedFollowLogs: shouldFollowLogs,
		expectedLogLines: expectedEmptyLogLineValues,
		expectedNotFoundServiceGuids: nonExistentUserServiceGuids,
	}

	getUserServiceLogsRequestInfoAndExpectedResultsList := []*getUserServiceLogsRequestInfoAndExpectedResults{
		firstCallRequestInfoAndExpectedResults,
		secondCallRequestInfoAndExpectedResults,
		thirdCallRequestInfoAndExpectedResults,
	}

	for _, userServiceLogsRequestInfoAndExpectedResults := range getUserServiceLogsRequestInfoAndExpectedResultsList {

		requestedEnclaveId := userServiceLogsRequestInfoAndExpectedResults.requestedEnclaveID
		requestedUserServiceGuids := userServiceLogsRequestInfoAndExpectedResults.requestedUserServiceGuids
		requestedShouldFollowLogs := userServiceLogsRequestInfoAndExpectedResults.requestedFollowLogs
		expectedLogLines := userServiceLogsRequestInfoAndExpectedResults.expectedLogLines
		expectedNonExistenceUserServiceGuids := userServiceLogsRequestInfoAndExpectedResults.expectedNotFoundServiceGuids

		serviceLogsStreamContentChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, requestedEnclaveId, requestedUserServiceGuids, requestedShouldFollowLogs)
		require.NoError(t, err)
		require.NotNil(t, cancelStreamUserServiceLogsFunc)
		require.NotNil(t, serviceLogsStreamContentChan)

		receivedLogLines := []string{}
		receivedNotFoundUserServiceGuids := map[services.ServiceGUID]bool{}

		var testEvaluationErr error

		shouldContinueInTheLoop := true

		for shouldContinueInTheLoop {
			select {
			case <-time.Tick(testTimeOut):
				testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
				shouldContinueInTheLoop = false
				break
			case serviceLogsStreamContent, isChanOpen := <-serviceLogsStreamContentChan:
				if !isChanOpen {
					shouldContinueInTheLoop = false
					break
				}

				userServiceLogsByGuid := serviceLogsStreamContent.GetServiceLogsByServiceGuids()
				notFoundGuids := serviceLogsStreamContent.GetNotFoundServiceGuids()

				userServiceLogs, found := userServiceLogsByGuid[userServiceGuid]
				if len(expectedNonExistenceUserServiceGuids) > 0 {
					require.False(t, found)
				} else {
					require.True(t, found)
				}

				receivedNotFoundUserServiceGuids = notFoundGuids

				for _, serviceLog := range userServiceLogs {
					receivedLogLines = append(receivedLogLines, serviceLog.GetContent())
				}

				if len(receivedLogLines) == len(expectedLogLines) {
					shouldContinueInTheLoop = false
					break
				}
			}
		}

		require.NoError(t, testEvaluationErr)
		require.Equal(t, expectedLogLines, receivedLogLines)
		require.Equal(t, expectedNonExistenceUserServiceGuids, receivedNotFoundUserServiceGuids)
		cancelStreamUserServiceLogsFunc()
	}
}


// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getExampleServiceConfig() *services.ContainerConfig {

	entrypointArgs := []string{"/bin/sh", "-c"}
	cmdArgs := []string{"for i in kurtosis test running successfully; do echo \"$i\"; done;"}

	containerConfig := services.NewContainerConfigBuilder(
		dockerGettingStartedImage,
	).WithEntrypointOverride(
		entrypointArgs,
	).WithCmdOverride(
		cmdArgs,
	).Build()
	return containerConfig
}
