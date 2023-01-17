//go:build !minikube
// +build !minikube

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

	exampleServiceId services.ServiceID = "stream-logs"

	testTimeOut = 180 * time.Second

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
	requestedEnclaveIdentifier   string
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

	serviceList, err := test_helpers.AddServicesWithLogLines(ctx, enclaveCtx, logLinesByService)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------

	enclaveUuid := enclaveCtx.GetEnclaveUuid()

	serviceGuids := map[services.ServiceGUID]bool{}
	for _, serviceCtx := range serviceList {
		serviceGuid := serviceCtx.GetServiceGUID()
		serviceGuids[serviceGuid] = true
	}

	serviceLogsRequestInfoAndExpectedResultsList := getServiceLogsRequestInfoAndExpectedResultsList(
		string(enclaveUuid),
		serviceGuids,
	)

	for _, serviceLogsRequestInfoAndExpectedResultsObj := range serviceLogsRequestInfoAndExpectedResultsList {

		requestedEnclaveIdentifier := serviceLogsRequestInfoAndExpectedResultsObj.requestedEnclaveIdentifier
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
			string(requestedEnclaveIdentifier),
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
//
//	Private helper functions
//
// ====================================================================================================
func getServiceLogsRequestInfoAndExpectedResultsList(
	enclaveIdentifier string,
	serviceGuids map[services.ServiceGUID]bool,
) []*serviceLogsRequestInfoAndExpectedResults {

	emptyServiceGuids := map[services.ServiceGUID]bool{}

	nonExistentServiceGuids := map[services.ServiceGUID]bool{
		nonExistentServiceGuid: true,
	}

	firstCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doesContainTextFilter,
	}

	secondCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doNotFilterLogLines,
	}

	thirdCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
		requestedServiceGuids:        serviceGuids,
		requestedFollowLogs:          shouldNotFollowLogs,
		expectedLogLines:             []string{firstLogLine, secondLogLine, thirdLogLine, lastLogLine},
		expectedNotFoundServiceGuids: emptyServiceGuids,
		logLineFilter:                doNotFilterLogLines,
	}

	fourthCallRequestInfoAndExpectedResults := &serviceLogsRequestInfoAndExpectedResults{
		requestedEnclaveIdentifier:   enclaveIdentifier,
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
