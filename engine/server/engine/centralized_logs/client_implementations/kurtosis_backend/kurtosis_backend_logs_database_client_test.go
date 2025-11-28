package kurtosis_backend

import (
	"bytes"
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	testEnclaveUuid      = "test-enclave"
	enclaveUuid          = enclave.EnclaveUUID(testEnclaveUuid)
	testUserService1Uuid = "test-user-service-1"
	testUserService2Uuid = "test-user-service-2"
	testUserService3Uuid = "test-user-service-3"

	logLine1 = "Starting feature 'centralized logs'"
	logLine2 = "Starting feature 'network partitioning'"
	logLine3 = "Starting feature 'network soft partitioning'"
	logLine4 = "Starting feature 'files storage'"
	logLine5 = "Starting feature 'files manager'"
	logLine6 = "The enclave was created"
	logLine7 = "User service started"
	logLine8 = "The data have being loaded"

	firstFilterText           = "feature"
	secondFilterText          = "Files"
	notFoundedFilterText      = "it shouldn't be found in the log lines"
	firstMatchRegexFilterStr  = "Starting.*partitioning'"
	secondMatchRegexFilterStr = "[S].*manager"

	testTimeOut     = 2 * time.Second
	followLogs      = true
	doNotFollowLogs = false
)

// We created this buffer type just to implement io.ReaderCloser
type closingBuffer struct {
	*bytes.Buffer
}

func (cb *closingBuffer) Close() error {
	//we don't actually have to do anything here, since the buffer is just some data in memory
	return nil
}

func TestStreamUserServiceLogs_WithFilters(t *testing.T) {

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 2,
		testUserService2Uuid: 2,
		testUserService3Uuid: 2,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(firstFilterText)
	secondTextFilter := logline.NewDoesNotContainTextLogLineFilter(secondFilterText)
	regexFilter := logline.NewDoesContainMatchRegexLogLineFilter(firstMatchRegexFilterStr)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
		*secondTextFilter,
		*regexFilter,
	}

	expectedFirstLogLine := logLine2

	successfulServiceLogs := getCommonSuccessfulServiceLogs()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)

}

func TestStreamUserServiceLogs_WithFiltersAndWithFollowLogs(t *testing.T) {

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 1,
		testUserService2Uuid: 1,
		testUserService3Uuid: 1,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(secondFilterText)
	regexFilter := logline.NewDoesContainMatchRegexLogLineFilter(secondMatchRegexFilterStr)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
		*regexFilter,
	}

	expectedFirstLogLine := logLine5

	successfulServiceLogs := getCommonSuccessfulServiceLogs()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogs_FilteredAllLogLines(t *testing.T) {

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	successfulServiceLogs := getCommonSuccessfulServiceLogs()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)

}

func TestStreamUserServiceLogs_FilteredAllLogLinesWithFollowLogs(t *testing.T) {
	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	successfulServiceLogs := getCommonSuccessfulServiceLogs()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)

}

func TestStreamUserServiceLogs_ErrorFromKurtosisBackend(t *testing.T) {

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	kurtosisBackendError := errors.New("and error coming from Kurtosis backend")

	successfulServiceLogs := getCommonSuccessfulServiceLogs()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		successfulServiceLogs,
		kurtosisBackendError,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.Error(t, testEvaluationErr)
	require.Contains(t, testEvaluationErr.Error(), kurtosisBackendError.Error())

}

func TestStreamUserServiceLogs_NoLogsFromKurtosisBackend(t *testing.T) {
	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	successfulServiceLogs := map[service.ServiceUUID]io.ReadCloser{}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogs_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
	expectedAmountLogLines := 100000

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	emptyFilters := []logline.LogLineFilter{}

	expectedFirstLogLine := logLine1

	logLines := []string{}

	for i := 0; i <= expectedAmountLogLines; i++ {
		logLines = append(logLines, logLine1)
	}

	logLinesStr := strings.Join(logLines, "\n")

	logLinesReadCloser := &closingBuffer{bytes.NewBufferString(logLinesStr)}

	successfulServiceLogs := map[service.ServiceUUID]io.ReadCloser{
		testUserService1Uuid: logLinesReadCloser,
	}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogs_EmptyLogLines(t *testing.T) {
	expectedAmountLogLines := 0

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	emptyFilters := []logline.LogLineFilter{}

	logLinesStr := ""

	logLinesReadCloser := &closingBuffer{bytes.NewBufferString(logLinesStr)}

	successfulServiceLogs := map[service.ServiceUUID]io.ReadCloser{
		testUserService1Uuid: logLinesReadCloser,
	}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		successfulServiceLogs,
		nil,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func executeStreamCallAndGetReceivedServiceLogLines(
	t *testing.T,
	logLinesFilters []logline.LogLineFilter,
	expectedServiceAmountLogLinesByServiceUuid map[service.ServiceUUID]int,
	shouldFollowLogs bool,
	successfulServiceLogs map[service.ServiceUUID]io.ReadCloser,
	kurtosisBackendError error,
) (map[service.ServiceUUID][]logline.LogLine, error) {

	ctx := context.Background()
	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	defer cancelCtxFunc()

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	userServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceUuids,
		Statuses: nil,
	}

	erroredUserServiceUuids := map[service.ServiceUUID]error{}

	receivedServiceLogsByUuid := map[service.ServiceUUID][]logline.LogLine{}

	for serviceUuid := range expectedServiceAmountLogLinesByServiceUuid {
		receivedServiceLogsByUuid[serviceUuid] = []logline.LogLine{}
	}

	kurtosisBackend := backend_interface.NewMockKurtosisBackend(t)

	kurtosisBackend.EXPECT().
		GetUserServiceLogs(ctxWithCancel, enclaveUuid, userServiceFilters, shouldFollowLogs).
		Return(
			successfulServiceLogs,
			erroredUserServiceUuids,
			kurtosisBackendError,
		)

	logsDatabaseClient := NewKurtosisBackendLogsDatabaseClient(kurtosisBackend)

	userServiceLogsByUuidChan, errChan, receivedCancelCtxFunc, err := logsDatabaseClient.StreamUserServiceLogs(
		ctx,
		enclaveUuid,
		userServiceUuids,
		logLinesFilters,
		shouldFollowLogs,
		true,
		0)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs for UUIDs '%+v' using log line filters '%v' in enclave '%v'", userServiceUuids, logLinesFilters, enclaveUuid)
	}
	defer func() {
		if receivedCancelCtxFunc != nil {
			receivedCancelCtxFunc()
		}
	}()

	require.NotNil(t, userServiceLogsByUuidChan, "Received a nil user service logs channel, but a non-nil value was expected")
	require.NotNil(t, errChan, "Received a nil error logs channel, but a non-nil value was expected")

	shouldReceiveStream := true
	for shouldReceiveStream {
		select {
		case <-time.Tick(testTimeOut):
			return nil, stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
		case userServiceLogsByUuid, isChanOpen := <-userServiceLogsByUuidChan:
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}

			for serviceUuid, serviceLogLines := range userServiceLogsByUuid {
				_, found := userServiceUuids[serviceUuid]
				require.True(t, found)

				currentServiceLogLines := receivedServiceLogsByUuid[serviceUuid]
				allServiceLogLines := append(currentServiceLogLines, serviceLogLines...)
				receivedServiceLogsByUuid[serviceUuid] = allServiceLogLines
			}

			for serviceUuid, expectedAmountLogLines := range expectedServiceAmountLogLinesByServiceUuid {
				if len(receivedServiceLogsByUuid[serviceUuid]) == expectedAmountLogLines {
					shouldReceiveStream = false
				} else {
					shouldReceiveStream = true
					break
				}
			}
		}
	}

	return receivedServiceLogsByUuid, nil
}

func getCommonSuccessfulServiceLogs() map[service.ServiceUUID]io.ReadCloser {
	logLines := []string{logLine1, logLine2, logLine3, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")

	logLinesReadCloser1 := &closingBuffer{bytes.NewBufferString(logLinesStr)}
	logLinesReadCloser2 := &closingBuffer{bytes.NewBufferString(logLinesStr)}
	logLinesReadCloser3 := &closingBuffer{bytes.NewBufferString(logLinesStr)}

	successfulServiceLogs := map[service.ServiceUUID]io.ReadCloser{
		testUserService1Uuid: logLinesReadCloser1,
		testUserService2Uuid: logLinesReadCloser2,
		testUserService3Uuid: logLinesReadCloser3,
	}

	return successfulServiceLogs
}
