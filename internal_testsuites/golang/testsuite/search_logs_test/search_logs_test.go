package search_logs_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite/consts"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/stretchr/testify/require"
)

const (
	testName = "search-logs"

	exampleServiceNamePrefix = "search-logs-"

	shouldFollowLogs    = true
	shouldNotFollowLogs = false

	service1ServiceName services.ServiceName = exampleServiceNamePrefix + "service-1"

	firstFilterText     = "The data have being loaded"
	secondFilterText    = "Starting feature"
	thirdFilterText     = "pool"
	matchRegexFilterStr = "Starting.*logs'"

	testTimeOut = 180 * time.Second

	logLine1 = "Starting feature 'centralized logs'"
	logLine2 = "Starting feature 'enclave pool'"
	logLine3 = "Starting feature 'enclave pool with size 2'"
	logLine4 = "The data have being loaded"

	maxLogRetrievalRetries    = 10
	logRetrievalRetryInterval = 10 * time.Second
)

var (
	expectedNonExistenceServiceUuids = map[services.ServiceUUID]bool{}

	service1LogLines = []string{
		logLine1,
		logLine2,
		logLine3,
		logLine4,
	}

	logLinesByService = map[services.ServiceName][]string{
		service1ServiceName: service1LogLines,
	}

	doesContainTextFilterForFirstRequest  = kurtosis_context.NewDoesContainTextLogLineFilter(firstFilterText)
	doesContainTextFilterForSecondRequest = kurtosis_context.NewDoesContainTextLogLineFilter(secondFilterText)
	doesNotContainTextFilter              = kurtosis_context.NewDoesNotContainTextLogLineFilter(thirdFilterText)
	doesContainMatchRegexFilter           = kurtosis_context.NewDoesContainMatchRegexLogLineFilter(matchRegexFilterStr)
	doesNotContainMatchRegexFilter        = kurtosis_context.NewDoesNotContainMatchRegexLogLineFilter(matchRegexFilterStr)

	filtersByRequest = []kurtosis_context.LogLineFilter{
		*doesContainTextFilterForFirstRequest,
		*doesContainTextFilterForSecondRequest,
		*doesNotContainTextFilter,
		*doesContainMatchRegexFilter,
		*doesNotContainMatchRegexFilter,
	}

	expectedLogLinesByRequest = [][]string{
		{
			logLine4,
		},
		{
			logLine1,
			logLine2,
			logLine3,
		},
		{
			logLine1,
			logLine4,
		},
		{
			logLine1,
		},
		{
			logLine2,
			logLine3,
			logLine4,
		},
	}

	shouldFollowLogsValueByRequest = []bool{
		shouldFollowLogs,
		shouldNotFollowLogs,
		shouldNotFollowLogs,
		shouldNotFollowLogs,
		shouldNotFollowLogs,
	}
)

func TestSearchLogs(t *testing.T) {
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
	require.NoError(t, err, "An error occurred adding services with log lines '%+v'", logLinesByService)
	require.Equal(t, len(logLinesByService), len(serviceList))

	// It takes some time for logs to persist so we sleep to ensure logs have persisted
	// Otherwise the test is flaky
	time.Sleep(consts.FluentbitRefreshInterval)
	// ------------------------------------- TEST RUN -------------------------------------------------
	enclaveUuid := enclaveCtx.GetEnclaveUuid()

	userServiceUuids := map[services.ServiceUUID]bool{}
	for _, serviceCtx := range serviceList {
		serviceUuid := serviceCtx.GetServiceUUID()
		userServiceUuids[serviceUuid] = true
	}

	expectedLogLinesByService := map[services.ServiceUUID][]string{}

	for requestIndex, filter := range filtersByRequest {

		for serviceUuid := range userServiceUuids {
			expectedLogLinesByService[serviceUuid] = expectedLogLinesByRequest[requestIndex]
		}

		shouldFollowLogsOption := shouldFollowLogsValueByRequest[requestIndex]

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
				string(enclaveUuid),
				userServiceUuids,
				expectedLogLinesByService,
				shouldFollowLogsOption,
				&filter,
			)

			if testEvaluationErr != nil {
				t.Logf("Attempt %d: error retrieving logs: %v", attempt+1, testEvaluationErr)
				time.Sleep(logRetrievalRetryInterval)
				continue
			}

			logsRetrieved = true
			for serviceUuid := range userServiceUuids {
				expectedLogLines := expectedLogLinesByRequest[requestIndex]
				receivedLogLines := receivedLogLinesByService[serviceUuid]
				if len(receivedLogLines) < len(expectedLogLines) {
					logsRetrieved = false
					t.Logf("Attempt %d: expected %d log lines for service %s, got %d. Retrying...",
						attempt+1, len(expectedLogLines), serviceUuid, len(receivedLogLines))
					break
				}
			}

			if logsRetrieved {
				break
			}
			time.Sleep(logRetrievalRetryInterval)
		}

		require.True(t, logsRetrieved, "Failed to retrieve expected logs after %d attempts", maxLogRetrievalRetries)
		require.NoError(t, testEvaluationErr)
		for serviceUuid := range userServiceUuids {
			receivedLogLines := receivedLogLinesByService[serviceUuid]
			expectedLogLines := expectedLogLinesByRequest[requestIndex]
			require.GreaterOrEqual(t, len(receivedLogLines), len(expectedLogLines),
				"Expected at least %d log lines for service %s but got %d", len(expectedLogLines), serviceUuid, len(receivedLogLines))
			for logNum, expectedLogLine := range expectedLogLines {
				require.Contains(t, receivedLogLines[logNum], expectedLogLine)
			}
		}
		require.Equal(t, expectedNonExistenceServiceUuids, receivedNotFoundServiceUuids)
	}
}
