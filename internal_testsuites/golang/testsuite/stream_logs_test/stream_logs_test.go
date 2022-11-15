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

	nonExistentServiceGuid = "stream-logs-1667939326-non-existent"
)

type serviceLogsRequestInfoAndExpectedResults struct {
	requestedEnclaveID    enclaves.EnclaveID
	requestedServiceGuids map[services.ServiceGUID]bool
	requestedFollowLogs   bool
	expectedLogLines []string
	expectedNotFoundServiceGuids map[services.ServiceGUID]bool
}

func TestStreamLogs(t *testing.T) {
	ctx := context.Background()

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

	serviceGuid := serviceCtx.GetServiceGUID()

	serviceGuids := map[services.ServiceGUID]bool{
		serviceGuid: true,
	}

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	serviceLogsRequestInfoAndExpectedResultsList := getServiceLogsRequestInfoAndExpectedResultsList(
		enclaveID,
		serviceGuids,
	)

	for _, serviceLogsRequestInfoAndExpectedResultsObj := range serviceLogsRequestInfoAndExpectedResultsList {

		requestedEnclaveId := serviceLogsRequestInfoAndExpectedResultsObj.requestedEnclaveID
		requestedServiceGuids := serviceLogsRequestInfoAndExpectedResultsObj.requestedServiceGuids
		requestedShouldFollowLogs := serviceLogsRequestInfoAndExpectedResultsObj.requestedFollowLogs
		expectedLogLines := serviceLogsRequestInfoAndExpectedResultsObj.expectedLogLines
		expectedNonExistenceServiceGuids := serviceLogsRequestInfoAndExpectedResultsObj.expectedNotFoundServiceGuids

		serviceLogsStreamContentChan, cancelStreamServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, requestedEnclaveId, requestedServiceGuids, requestedShouldFollowLogs)
		require.NoError(t, err)
		require.NotNil(t, cancelStreamServiceLogsFunc)
		require.NotNil(t, serviceLogsStreamContentChan)

		receivedLogLines := []string{}
		receivedNotFoundServiceGuids := map[services.ServiceGUID]bool{}

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

				serviceLogsByGuid := serviceLogsStreamContent.GetServiceLogsByServiceGuids()
				notFoundGuids := serviceLogsStreamContent.GetNotFoundServiceGuids()

				serviceLogLines, found := serviceLogsByGuid[serviceGuid]
				if len(expectedNonExistenceServiceGuids) > 0 {
					require.False(t, found)
				} else {
					require.True(t, found)
				}

				receivedNotFoundServiceGuids = notFoundGuids

				for _, serviceLog := range serviceLogLines {
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
		require.Equal(t, expectedNonExistenceServiceGuids, receivedNotFoundServiceGuids)
		cancelStreamServiceLogsFunc()
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

func getServiceLogsRequestInfoAndExpectedResultsList(
	enclaveId enclaves.EnclaveID,
	serviceGuids map[services.ServiceGUID]bool,
) []*serviceLogsRequestInfoAndExpectedResults {

	expectedLogLineValues := []string{"kurtosis", "test", "running", "successfully"}
	expectedEmptyLogLineValues := []string{}

	emptyServiceGuids := map[services.ServiceGUID]bool{}

	nonExistentServiceGuids := map[services.ServiceGUID]bool{
		nonExistentServiceGuid: true,
	}

	firstCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldNotFollowLogs,
		expectedLogLines:             expectedLogLineValues,
		expectedNotFoundServiceGuids: emptyServiceGuids,

	}

	secondCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             expectedLogLineValues,
		expectedNotFoundServiceGuids: emptyServiceGuids,
	}

	thirdCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        nonExistentServiceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             expectedEmptyLogLineValues,
		expectedNotFoundServiceGuids: nonExistentServiceGuids,
	}

	serviceLogsRequestInfoAndExpectedResultsList := []*serviceLogsRequestInfoAndExpectedResults{
		firstCallRequestInfoAndExpectedResults,
		secondCallRequestInfoAndExpectedResults,
		thirdCallRequestInfoAndExpectedResults,
	}

	return serviceLogsRequestInfoAndExpectedResultsList
}
