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
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	testName              = "stream-logs"
	isPartitioningEnabled = false

	dockerGettingStartedImage                    = "docker/getting-started"
	exampleServiceId          services.ServiceID = "stream-logs"

	waitForAllLogsBeingCollectedInSeconds = 3

	testTimeOut = 90 * time.Second

	shouldFollowLogs  = true
	shouldNotFollowLogs = false

	nonExistentServiceGuid = "stream-logs-1667939326-non-existent"
)

var doNotFilterLogLines *kurtosis_context.LogLineFilter = nil

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

	enclaveId := enclaveCtx.GetEnclaveID()

	serviceCtx, err := enclaveCtx.GetServiceContext(exampleServiceId)
	require.NoError(t, err)

	serviceGuid := serviceCtx.GetServiceGUID()

	serviceGuids := map[services.ServiceGUID]bool{
		serviceGuid: true,
	}

	time.Sleep(waitForAllLogsBeingCollectedInSeconds * time.Second)

	serviceLogsRequestInfoAndExpectedResultsList := getServiceLogsRequestInfoAndExpectedResultsList(
		enclaveId,
		serviceGuids,
	)

	for _, serviceLogsRequestInfoAndExpectedResultsObj := range serviceLogsRequestInfoAndExpectedResultsList {

		requestedEnclaveId := serviceLogsRequestInfoAndExpectedResultsObj.requestedEnclaveID
		requestedServiceGuids := serviceLogsRequestInfoAndExpectedResultsObj.requestedServiceGuids
		requestedShouldFollowLogs := serviceLogsRequestInfoAndExpectedResultsObj.requestedFollowLogs
		expectedLogLines := serviceLogsRequestInfoAndExpectedResultsObj.expectedLogLines
		expectedNonExistenceServiceGuids := serviceLogsRequestInfoAndExpectedResultsObj.expectedNotFoundServiceGuids

		expectedLogLinesByService := map[services.ServiceGUID][]string{}
		for userServiceGuid := range requestedServiceGuids{
			expectedLogLinesByService[userServiceGuid] = expectedLogLines
		}

		testEvaluationErr, receivedLogLinesByService, receivedNotFoundServiceGuids := test_helpers.GetLogsResponse(
			t,
			ctx,
			testTimeOut,
			kurtosisCtx,
			requestedEnclaveId,
			requestedServiceGuids,
			expectedLogLinesByService,
			requestedShouldFollowLogs,
			doNotFilterLogLines,
		)

		require.NoError(t, testEvaluationErr)
		for userServiceGuid := range requestedServiceGuids {
			require.Equal(t, expectedLogLines, receivedLogLinesByService[userServiceGuid])
		}
		require.Equal(t, expectedNonExistenceServiceGuids, receivedNotFoundServiceGuids)
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
