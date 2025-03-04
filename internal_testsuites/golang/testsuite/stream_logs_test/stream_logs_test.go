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
	testName = "stream-logs"

	exampleServiceName services.ServiceName = "stream-logs"

	testTimeOut = 180 * time.Second

	shouldFollowLogs    = true
	shouldNotFollowLogs = false

	nonExistentServiceUuid = "stream-logs-1667939326-non-existent"

	firstLogLine  = "kurtosis"
	secondLogLine = "test"
	thirdLogLine  = "running"
	lastLogLine   = "successfully"

	secondsToWaitForLogs = 8 * time.Second
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

	logLinesByService = map[services.ServiceName][]string{
		exampleServiceName: exampleServiceLogLines,
	}
)

type serviceLogsRequestInfoAndExpectedResults struct {
	requestedEnclaveIdentifier   string
	requestedServiceUuids        map[services.ServiceUUID]bool
	requestedFollowLogs          bool
	expectedLogLines             []string
	expectedNotFoundServiceUuids map[services.ServiceUUID]bool
	logLineFilter                *kurtosis_context.LogLineFilter
}

func TestStreamLogs(t *testing.T) {

	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, _, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	//defer func() {
	//	err = destroyEnclaveFunc()
	//	require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	//}()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	serviceList, err := test_helpers.AddServicesWithLogLines(ctx, enclaveCtx, logLinesByService)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// It takes some time for logs to persist so we sleep to ensure logs have persisted
	// Otherwise the test is flaky
	time.Sleep(secondsToWaitForLogs)
	// ------------------------------------- TEST RUN ----------------------------------------------

	enclaveUuid := enclaveCtx.GetEnclaveUuid()

	serviceUuids := map[services.ServiceUUID]bool{}
	for _, serviceCtx := range serviceList {
		serviceUuid := serviceCtx.GetServiceUUID()
		serviceUuids[serviceUuid] = true
	}

	serviceLogsRequestInfoAndExpectedResultsList := getServiceLogsRequestInfoAndExpectedResultsList(
		string(enclaveUuid),
		serviceUuids,
	)

	for _, serviceLogsRequestInfoAndExpectedResultsObj := range serviceLogsRequestInfoAndExpectedResultsList {

		requestedEnclaveIdentifier := serviceLogsRequestInfoAndExpectedResultsObj.requestedEnclaveIdentifier
		requestedServiceUuids := serviceLogsRequestInfoAndExpectedResultsObj.requestedServiceUuids
		requestedShouldFollowLogs := serviceLogsRequestInfoAndExpectedResultsObj.requestedFollowLogs
		expectedLogLines := serviceLogsRequestInfoAndExpectedResultsObj.expectedLogLines
		expectedNonExistenceServiceUuids := serviceLogsRequestInfoAndExpectedResultsObj.expectedNotFoundServiceUuids
		filter := serviceLogsRequestInfoAndExpectedResultsObj.logLineFilter

		expectedLogLinesByService := map[services.ServiceUUID][]string{}
		for userServiceUuid := range requestedServiceUuids {
			expectedLogLinesByService[userServiceUuid] = expectedLogLines
		}

		testEvaluationErr, receivedLogLinesByService, receivedNotFoundServiceUuids := test_helpers.GetLogsResponse(
			t,
			ctx,
			testTimeOut,
			kurtosisCtx,
			requestedEnclaveIdentifier,
			requestedServiceUuids,
			expectedLogLinesByService,
			requestedShouldFollowLogs,
			filter,
		)

		require.NoError(t, testEvaluationErr)
		for userServiceUuid := range requestedServiceUuids {
			require.Equal(t, expectedLogLines, receivedLogLinesByService[userServiceUuid])
		}
		require.Equal(t, expectedNonExistenceServiceUuids, receivedNotFoundServiceUuids)
	}
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func getServiceLogsRequestInfoAndExpectedResultsList(
	enclaveIdentifier string,
	serviceUuids map[services.ServiceUUID]bool,
) []*serviceLogsRequestInfoAndExpectedResults {

	emptyServiceUuids := map[services.ServiceUUID]bool{}

	nonExistentServiceUuids := map[services.ServiceUUID]bool{
		nonExistentServiceUuid: true,
	}

	firstCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceUuids:        serviceUuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{lastLogLine},
		expectedNotFoundServiceUuids: emptyServiceUuids,
		logLineFilter:                doesContainTextFilter,
	}

	secondCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceUuids:        serviceUuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceUuids: emptyServiceUuids,
		logLineFilter:                doNotFilterLogLines,
	}

	thirdCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceUuids:        serviceUuids,
		requestedFollowLogs:          shouldNotFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceUuids: emptyServiceUuids,
		logLineFilter:                doNotFilterLogLines,
	}

	fourthCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceUuids:        nonExistentServiceUuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{},
		expectedNotFoundServiceUuids: nonExistentServiceUuids,
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
