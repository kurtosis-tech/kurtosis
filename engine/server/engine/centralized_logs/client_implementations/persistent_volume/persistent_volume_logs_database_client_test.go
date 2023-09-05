package persistent_volume

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/stream_logs_strategy"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const (
	// We use this storage path for tests because fstest.MapFS doesn't like forward slashes
	logsStorageDirpathForTests = "var/log/kurtosis/"

	testEnclaveUuid      = "test-enclave"
	enclaveUuid          = enclave.EnclaveUUID(testEnclaveUuid)
	testUserService1Uuid = "test-user-service-1"
	testUserService2Uuid = "test-user-service-2"
	testUserService3Uuid = "test-user-service-3"

	logLine1  = "{\"log\":\"Starting feature 'centralized logs'\"}"
	logLine2  = "{\"log\":\"Starting feature 'runs idempotently'\"}"
	logLine3a = "{\"log\":\"Starting feature 'apic "
	logLine3b = "idempotently'\"}"
	logLine4  = "{\"log\":\"Starting feature 'files storage'\"}"
	logLine5  = "{\"log\":\"Starting feature 'files manager'\"}"
	logLine6  = "{\"log\":\"The enclave was created\"}"
	logLine7  = "{\"log\":\"User service started\"}"
	logLine8  = "{\"log\":\"The data have being loaded\"}"

	firstFilterText          = "feature"
	secondFilterText         = "Files"
	notFoundedFilterText     = "it shouldn't be found in the log lines"
	firstMatchRegexFilterStr = "Starting.*idempotently'"

	testTimeOut     = 2000 * time.Second
	doNotFollowLogs = false

	defaultYear  = 2023
	defaultDay   = 0 // sunday
	startingWeek = 4

	// base logs file path/enclave uuid/service uuid <filetype>
	perFileFilePathFmtStr = "%s%s/%s%s"

	// base logs file path/week/enclave uuid/service uuid <filetype>
	perWeekFilePathFmtStr = "%s%s/%s/%s%s"
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

	expectedFirstLogLine := "Starting feature 'runs idempotently'"

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	underlyingFs := createFilledPerFileFilesystem()
	perFileStreamStrategy := stream_logs_strategy.NewPerFileStreamLogsStrategy()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perFileStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_WithFilters(t *testing.T) {
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

	expectedFirstLogLine := "Starting feature 'runs idempotently'"

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	underlyingFs := createFilledPerWeekFilesystem(startingWeek)
	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
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

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	underlyingFs := createEmptyPerFileFilesystem()
	perFileStreamStrategy := stream_logs_strategy.NewPerFileStreamLogsStrategy()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perFileStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_NoLogsFromPersistentVolume(t *testing.T) {
	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: 0,
		testUserService2Uuid: 0,
		testUserService3Uuid: 0,
	}

	firstTextFilter := logline.NewDoesContainTextLogLineFilter(notFoundedFilterText)

	logLinesFilters := []logline.LogLineFilter{
		*firstTextFilter,
	}

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
		testUserService2Uuid: true,
		testUserService3Uuid: true,
	}

	underlyingFs := createEmptyPerWeekFilesystem(startingWeek)
	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogs_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
	expectedAmountLogLines := 10_000

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	emptyFilters := []logline.LogLineFilter{}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	logLines := []string{}

	for i := 0; i <= expectedAmountLogLines; i++ {
		logLines = append(logLines, logLine1)
	}

	logLinesStr := strings.Join(logLines, "\n")

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	file1 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, string(enclaveUuid), testUserService1Uuid, consts.Filetype)
	underlyingFs := &fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}
	perFileStreamStrategy := stream_logs_strategy.NewPerFileStreamLogsStrategy()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perFileStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
	expectedAmountLogLines := 10_000

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	emptyFilters := []logline.LogLineFilter{}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	logLines := []string{}

	for i := 0; i <= expectedAmountLogLines; i++ {
		logLines = append(logLines, logLine1)
	}

	logLinesStr := strings.Join(logLines, "\n")

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	file1 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(startingWeek), string(enclaveUuid), testUserService1Uuid, consts.Filetype)
	underlyingFs := &fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}
	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
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

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	logLinesStr := ""

	file1 := fmt.Sprintf("%s%s/%s%s", logsStorageDirpathForTests, string(enclaveUuid), testUserService1Uuid, consts.Filetype)
	underlyingFs := &fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}
	perFileStreamStrategy := stream_logs_strategy.NewPerFileStreamLogsStrategy()

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perFileStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_EmptyLogLines(t *testing.T) {
	expectedAmountLogLines := 0

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	emptyFilters := []logline.LogLineFilter{}

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	logLinesStr := ""

	file1 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(startingWeek), string(enclaveUuid), testUserService1Uuid, consts.Filetype)
	underlyingFs := &fstest.MapFS{
		file1: {
			Data: []byte(logLinesStr),
		},
	}
	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_WithLogsAcrossWeeks(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	logLinesFilters := []logline.LogLineFilter{}

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	week4logLines := []string{
		logLine5,
		logLine6,
		logLine7,
		logLine8}
	week3logLines := []string{
		logLine1,
		logLine2,
		logLine3a,
		logLine3b,
		logLine4}

	week3logLinesStr := strings.Join(week3logLines, "\n") + "\n"
	week4logLinesStr := strings.Join(week4logLines, "\n")

	week4filepath := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(4), testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week3filepath := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(3), testEnclaveUuid, testUserService1Uuid, consts.Filetype)

	underlyingFs := &fstest.MapFS{
		week4filepath: {
			Data: []byte(week4logLinesStr),
		},
		week3filepath: {
			Data: []byte(week3logLinesStr),
		},
	}
	mockTime := logs_clock.NewMockLogsClock(defaultYear, 4, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

	require.NoError(t, testEvaluationErr)
}

func TestStreamUserServiceLogsPerWeek_WithLogLineAcrossWeeks(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	logLinesFilters := []logline.LogLineFilter{}

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	week4logLines := []string{
		logLine3b,
		logLine4,
		logLine5,
		logLine6,
		logLine7,
		logLine8}
	week3logLines := []string{
		logLine1,
		logLine2,
		logLine3a}

	week3logLinesStr := strings.Join(week3logLines, "\n") + "\n"
	week4logLinesStr := strings.Join(week4logLines, "\n")

	week4filepath := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(4), testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week3filepath := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(3), testEnclaveUuid, testUserService1Uuid, consts.Filetype)

	underlyingFs := &fstest.MapFS{
		week4filepath: {
			Data: []byte(week4logLinesStr),
		},
		week3filepath: {
			Data: []byte(week3logLinesStr),
		},
	}
	mockTime := logs_clock.NewMockLogsClock(defaultYear, 4, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
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
	userServiceUuids map[service.ServiceUUID]bool,
	expectedServiceAmountLogLinesByServiceUuid map[service.ServiceUUID]int,
	shouldFollowLogs bool,
	underlyingFs *fstest.MapFS,
	streamStrategy stream_logs_strategy.StreamLogsStrategy,
) (map[service.ServiceUUID][]logline.LogLine, error) {
	ctx := context.Background()

	receivedServiceLogsByUuid := map[service.ServiceUUID][]logline.LogLine{}

	for serviceUuid := range expectedServiceAmountLogLinesByServiceUuid {
		receivedServiceLogsByUuid[serviceUuid] = []logline.LogLine{}
	}

	kurtosisBackend := backend_interface.NewMockKurtosisBackend(t)

	mockedFs := volume_filesystem.NewMockedVolumeFilesystem(underlyingFs)
	logsDatabaseClient := NewPersistentVolumeLogsDatabaseClient(kurtosisBackend, mockedFs, streamStrategy)

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

	shouldReceiveStream := true
	for shouldReceiveStream {
		select {
		case <-time.Tick(testTimeOut):
			return nil, stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
		case streamErr, isChanOpen := <-errChan:
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}
			return nil, stacktrace.Propagate(streamErr, "Receiving streaming error.")
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

func createFilledPerFileFilesystem() *fstest.MapFS {
	logLines := []string{logLine1, logLine2, logLine3a, logLine3b, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")

	file1 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	file2 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService2Uuid, consts.Filetype)
	file3 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService3Uuid, consts.Filetype)

	mapFs := &fstest.MapFS{
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

func createFilledPerWeekFilesystem(week int) *fstest.MapFS {
	logLines := []string{logLine1, logLine2, logLine3a, logLine3b, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")

	file1 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	file2 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService2Uuid, consts.Filetype)
	file3 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService3Uuid, consts.Filetype)

	mapFs := &fstest.MapFS{
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

func createEmptyPerFileFilesystem() *fstest.MapFS {
	file1 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	file2 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService2Uuid, consts.Filetype)
	file3 := fmt.Sprintf(perFileFilePathFmtStr, logsStorageDirpathForTests, testEnclaveUuid, testUserService3Uuid, consts.Filetype)

	mapFs := &fstest.MapFS{
		file1: {
			Data: []byte{},
		},
		file2: {
			Data: []byte{},
		},
		file3: {
			Data: []byte{},
		},
	}

	return mapFs
}

func createEmptyPerWeekFilesystem(week int) *fstest.MapFS {
	file1 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	file2 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService2Uuid, consts.Filetype)
	file3 := fmt.Sprintf(perWeekFilePathFmtStr, logsStorageDirpathForTests, strconv.Itoa(week), testEnclaveUuid, testUserService3Uuid, consts.Filetype)

	mapFs := &fstest.MapFS{
		file1: {
			Data: []byte{},
		},
		file2: {
			Data: []byte{},
		},
		file3: {
			Data: []byte{},
		},
	}

	return mapFs
}
