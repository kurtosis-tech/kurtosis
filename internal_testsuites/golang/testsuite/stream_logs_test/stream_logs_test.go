package stream_logs_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite/consts"
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

	// Retry configuration for log retrieval
	// Logs may not be immediately available due to Fluentbit/Vector flush timing;
	// in CI the flush can take well over 30 seconds so we allow up to ~100 s of retries.
	maxLogRetrievalRetries    = 10
	logRetrievalRetryInterval = 10 * time.Second
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
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	require.NoError(t, err)

	serviceList, err := test_helpers.AddServicesWithLogLines(ctx, enclaveCtx, logLinesByService)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// It takes some time for logs to persist so we sleep to ensure logs have persisted
	// Otherwise the test is flaky
	time.Sleep(consts.FluentbitRefreshInterval)
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

		// Retry log retrieval to handle Fluentbit/Vector flush timing
		var receivedLogLinesByService map[services.ServiceUUID][]string
		var receivedNotFoundServiceUuids map[services.ServiceUUID]bool
		var testEvaluationErr error
		var logsRetrieved bool

		for attempt := 0; attempt < maxLogRetrievalRetries; attempt++ {
			receivedLogLinesByService, receivedNotFoundServiceUuids, testEvaluationErr = test_helpers.GetLogsResponse(
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

			if testEvaluationErr != nil {
				t.Logf("Attempt %d: error retrieving logs: %v", attempt+1, testEvaluationErr)
				time.Sleep(logRetrievalRetryInterval)
				continue
			}

			// Check if we received enough logs for all services
			logsRetrieved = true
			for userServiceUuid := range requestedServiceUuids {
				expectedLines := expectedLogLinesByService[userServiceUuid]
				receivedLines := receivedLogLinesByService[userServiceUuid]
				if len(receivedLines) < len(expectedLines) {
					logsRetrieved = false
					t.Logf("Attempt %d: expected %d log lines for service %s, got %d. Retrying...",
						attempt+1, len(expectedLines), userServiceUuid, len(receivedLines))
					break
				}
			}

			if logsRetrieved {
				break
			}
			time.Sleep(logRetrievalRetryInterval)
		}

		require.True(t, logsRetrieved || len(expectedLogLines) == 0,
			"Failed to retrieve expected logs after %d attempts", maxLogRetrievalRetries)
		require.NoError(t, testEvaluationErr)

		for userServiceUuid := range requestedServiceUuids {
			expectedLines := expectedLogLinesByService[userServiceUuid]
			receivedLines := receivedLogLinesByService[userServiceUuid]
			require.GreaterOrEqual(t, len(receivedLines), len(expectedLines),
				"Expected at least %d log lines for service %s, but got %d",
				len(expectedLines), userServiceUuid, len(receivedLines))
			for logNum, expectedLogLine := range expectedLines {
				require.Contains(t, receivedLines[logNum], expectedLogLine)
			}
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
