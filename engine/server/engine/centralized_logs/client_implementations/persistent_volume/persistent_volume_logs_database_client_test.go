package persistent_volume

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const (
	testEnclaveUuid      = "test-enclave"
	enclaveUuid          = enclave.EnclaveUUID(testEnclaveUuid)
	testUserService1Uuid = "test-user-service-1"
	testUserService2Uuid = "test-user-service-2"
	testUserService3Uuid = "test-user-service-3"

	logLine1 = "{\"log\":\"Starting feature 'centralized logs'\"}"
	logLine2 = "{\"log\":\"Starting feature 'network partitioning'\"}"
	logLine3 = "{\"log\":\"Starting feature 'network soft partitioning'\"}"
	logLine4 = "{\"log\":\"Starting feature 'files storage'\"}"
	logLine5 = "{\"log\":\"Starting feature 'files manager'\"}"
	logLine6 = "{\"log\":\"The enclave was created\"}"
	logLine7 = "{\"log\":\"User service started\"}"
	logLine8 = "{\"log\":\"The data have being loaded\"}"

	firstFilterText           = "feature"
	secondFilterText          = "Files"
	notFoundedFilterText      = "it shouldn't be found in the log lines"
	firstMatchRegexFilterStr  = "Starting.*partitioning'"
	secondMatchRegexFilterStr = "[S].*manager"

	testTimeOut     = 2 * time.Second
	followLogs      = true
	doNotFollowLogs = false
)

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

	underlyingFs := createUnderlyingMapFilesystem()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
	)

	t.Logf("RECEIVER USER SERVICE LOGS BY UUID: %v", receivedUserServiceLogsByUuid)
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

	underlyingFs := createUnderlyingMapFilesystem()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		underlyingFs,
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

	underlyingFs := createUnderlyingMapFilesystem()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
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

	underlyingFs := createUnderlyingMapFilesystem()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		underlyingFs,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)

}

func TestStreamUserServiceLogs_NoLogsFromPersistentVolume(t *testing.T) {
	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	underlyingFs := fstest.MapFS{}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		followLogs,
		underlyingFs,
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

	file1 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), testUserService1Uuid, filetype)

	underlyingFs := fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
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

	file1 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), testUserService1Uuid, filetype)

	underlyingFs := fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
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
	underlyingFs fstest.MapFS,
) (map[service.ServiceUUID][]logline.LogLine, error) {
	t.Log("EXECUTING STREAM CALL AND GETTING RECEIVED SERVICE LOG LINE")
	ctx := context.Background()

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	receivedServiceLogsByUuid := map[service.ServiceUUID][]logline.LogLine{}

	for serviceUuid := range expectedServiceAmountLogLinesByServiceUuid {
		receivedServiceLogsByUuid[serviceUuid] = []logline.LogLine{}
	}

	kurtosisBackend := backend_interface.NewMockKurtosisBackend(t)
	mockedFs := NewMockedVolumeFilesystem(underlyingFs)

	logsDatabaseClient := NewPersistentVolumeLogsDatabaseClient(kurtosisBackend, mockedFs)
	t.Log("CALLING STREAM USER SERVICE LOGS")
	userServiceLogsByUuidChan, errChan, receivedCancelCtxFunc, err := logsDatabaseClient.StreamUserServiceLogs(ctx, enclaveUuid, userServiceUuids, logLinesFilters, shouldFollowLogs)
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
	t.Log("DID NOT RECEIVE NIL SERVICE LOGS CHANNEL")

	shouldReceiveStream := true
	for shouldReceiveStream {
		t.Log("RECEIVING STREAM")
		select {
		case <-time.Tick(testTimeOut):
			t.Log("TIMEOUT ERROR")
			return nil, stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
		case streamErr, isChanOpen := <-errChan:
			t.Log("RECEIVING STREAM ERROR")
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}

			t.Logf("RECEIVING STREAM ERR: %v", streamErr)
		case userServiceLogsByUuid, isChanOpen := <-userServiceLogsByUuidChan:
			t.Log("RECEIVING LOGS")
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

func createUnderlyingMapFilesystem() fstest.MapFS {
	logLines := []string{logLine1, logLine2, logLine3, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")

	file1 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), testUserService1Uuid, filetype)
	file2 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), testUserService2Uuid, filetype)
	file3 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpath, string(enclaveUuid), testUserService3Uuid, filetype)

	mapFs := fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
		file2: {
			Data: []byte(logLinesStr),
		},
		file3: {
			Data: []byte(logLinesStr),
		},
	}

	return mapFs
}
