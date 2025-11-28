package persisted_logs_test

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
	// Tests that logs are still around after services are removed
	testName = "persisted-logs"

	exampleServiceNamePrefix = "persisted-logs-"

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

func TestPersistedLogs(t *testing.T) {
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
	for serviceName, serviceCtx := range serviceList {
		serviceUuid := serviceCtx.GetServiceUUID()
		userServiceUuids[serviceUuid] = true

		// NOW REMOVE THE SERVICE
		err = test_helpers.RemoveService(ctx, enclaveCtx, serviceName)
		require.NoError(t, err, "An error occurred trying to remove service '%v' during test", serviceName)
	}

	expectedLogLinesByService := map[services.ServiceUUID][]string{}

	for requestIndex, filter := range filtersByRequest {

		for serviceUuid := range userServiceUuids {
			expectedLogLinesByService[serviceUuid] = expectedLogLinesByRequest[requestIndex]
		}

		shouldFollowLogsOption := shouldFollowLogsValueByRequest[requestIndex]

		receivedLogLinesByService, receivedNotFoundServiceUuids, testEvaluationErr := test_helpers.GetLogsResponse(
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

		require.NoError(t, testEvaluationErr)
		for serviceUuid := range userServiceUuids {
			for logNum, expectedLogLine := range expectedLogLinesByRequest[requestIndex] {
				require.Contains(t, receivedLogLinesByService[serviceUuid][logNum], expectedLogLine)
			}
		}
		require.Equal(t, expectedNonExistenceServiceUuids, receivedNotFoundServiceUuids)
	}
}
