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

	exampleServiceId          services.ServiceID = "stream-logs"

	testTimeOut = 90 * time.Second

	shouldFollowLogs    = true
	shouldNotFollowLogs = false

	nonExistentServiceGuid = "stream-logs-1667939326-non-existent"

	firstLogLine  = "kurtosis"
	secondLogLine = "test"
	thirdLogLine  = "running"
	lastLogLine   = "successfully"
)

var (
	doNotFilterLogLines   *kurtosis_context.LogLineFilter = nil
	doesContainTextFilter                                 = kurtosis_context.NewDoesContainTextLogLineFilter(lastLogLine)

	exampleServiceLogLines = []string{
		firstLogLine,
		secondLogLine,
		thirdLogLine,
		lastLogLine,
	}

	logLinesByService = map[services.ServiceID][]string{
		exampleServiceId: exampleServiceLogLines,
	}
)

type serviceLogsRequestInfoAndExpectedResults struct {
	requestedEnclaveID           enclaves.EnclaveID
	requestedServiceGuids        map[services.ServiceGUID]bool
	requestedFollowLogs          bool
	expectedLogLines             []string
	expectedNotFoundServiceGuids map[services.ServiceGUID]bool
	logLineFilter                *kurtosis_context.LogLineFilter
}

func TestStreamLogs(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	serviceList, err := test_helpers.AddServicesWithLogLines(enclaveCtx, logLinesByService)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------

	enclaveId := enclaveCtx.GetEnclaveID()

	serviceGuids := map[services.ServiceGUID]bool{}
	for _, serviceCtx := range serviceList {
		serviceGuid := serviceCtx.GetServiceGUID()
		serviceGuids[serviceGuid] = true
	}

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
		filter := serviceLogsRequestInfoAndExpectedResultsObj.logLineFilter

		expectedLogLinesByService := map[services.ServiceGUID][]string{}
		for userServiceGuid := range requestedServiceGuids {
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
			filter,
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
func getServiceLogsRequestInfoAndExpectedResultsList(
	enclaveId enclaves.EnclaveID,
	serviceGuids map[services.ServiceGUID]bool,
) []*serviceLogsRequestInfoAndExpectedResults {

	emptyServiceGuids := map[services.ServiceGUID]bool{}

	nonExistentServiceGuids := map[services.ServiceGUID]bool{
		nonExistentServiceGuid: true,
	}

	firstCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doesContainTextFilter,
	}

	secondCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doNotFilterLogLines,
	}

	thirdCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldNotFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doNotFilterLogLines,
	}

	fourthCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveID:           enclaveId,
		requestedServiceGuids:        nonExistentServiceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{},
		expectedNotFoundServiceGuids: nonExistentServiceGuids,
		logLineFilter:                doNotFilterLogLines,
	}

	serviceLogsRequestInfoAndExpectedResultsList := []*serviceLogsRequestInfoAndExpectedResults{
		firstCallRequestInfoAndExpectedResults,
		secondCallRequestInfoAndExpectedResults,
		thirdCallRequestInfoAndExpectedResults,
		fourthCallRequestInfoAndExpectedResults,
	}

	return serviceLogsRequestInfoAndExpectedResultsList
}
