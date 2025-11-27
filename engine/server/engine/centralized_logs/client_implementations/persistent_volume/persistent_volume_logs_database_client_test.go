package persistent_volume

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/file_layout"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/stream_logs_strategy"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/require"
	"strconv"
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

	retentionPeriodInWeeksForTesting = 5

	utcFormat              = time.RFC3339
	defaultUTCTimestampStr = "2023-09-06T00:35:15-04:00"
	logLine1               = "{\"log\":\"Starting feature 'centralized logs'\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine2               = "{\"log\":\"Starting feature 'runs idempotently'\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine3a              = "{\"log\":\"Starting feature 'apic "
	logLine3b              = "idempotently'\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine4               = "{\"log\":\"Starting feature 'files storage'\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine5               = "{\"log\":\"Starting feature 'files manager'\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine6               = "{\"log\":\"The enclave was created\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine7               = "{\"log\":\"User service started\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"
	logLine8               = "{\"log\":\"The data have being loaded\", \"timestamp\":\"2023-09-06T00:35:15-04:00\"}"

	firstFilterText          = "feature"
	secondFilterText         = "Files"
	notFoundedFilterText     = "it shouldn't be found in the log lines"
	firstMatchRegexFilterStr = "Starting.*idempotently'"

	testTimeOut     = 2 * time.Minute
	doNotFollowLogs = false

	defaultYear  = 2023
	defaultDay   = 0 // sunday
	startingWeek = 4

	defaultShouldReturnAllLogs = true
	defaultNumLogLines         = 0
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
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}
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
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

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
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}
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
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}
}

func TestStreamUserServiceLogs_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
	expectedAmountLogLines := 10_000

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var emptyFilters []logline.LogLineFilter

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	var logLines []string

	for i := 0; i <= expectedAmountLogLines; i++ {
		logLines = append(logLines, logLine1)
	}

	logLinesStr := strings.Join(logLines, "\n")

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()

	file1PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, string(enclaveUuid), testUserService1Uuid, volume_consts.Filetype)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

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
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}
}

func TestStreamUserServiceLogsPerWeek_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
	expectedAmountLogLines := 10_000

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var emptyFilters []logline.LogLineFilter

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	var logLines []string

	for i := 0; i <= expectedAmountLogLines; i++ {
		logLines = append(logLines, logLine1)
	}

	logLinesStr := strings.Join(logLines, "\n")

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()
	// %02d to format week num with leading zeros so 1-9 are converted to 01-09 for %V format
	formattedWeekNum := fmt.Sprintf("%02d", startingWeek)
	file1PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, string(enclaveUuid), testUserService1Uuid, volume_consts.Filetype)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}
}

func TestStreamUserServiceLogs_EmptyLogLines(t *testing.T) {
	expectedAmountLogLines := 0

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var emptyFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	logLinesStr := ""

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()
	file1PathStr := fmt.Sprintf("%s%s/%s%s", volume_consts.LogsStorageDirpath, string(enclaveUuid), testUserService1Uuid, volume_consts.Filetype)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

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
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}
}

func TestStreamUserServiceLogsPerWeek_EmptyLogLines(t *testing.T) {
	expectedAmountLogLines := 0

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var emptyFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	logLinesStr := ""

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()
	formattedWeekNum := fmt.Sprintf("%02d", startingWeek)
	file1PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, string(enclaveUuid), testUserService1Uuid, volume_consts.Filetype)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
	}
}

func TestStreamUserServiceLogsPerWeek_WithLogsAcrossWeeks(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

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

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()

	week3logLinesStr := strings.Join(week3logLines, "\n") + "\n"
	week4logLinesStr := strings.Join(week4logLines, "\n")

	formattedWeekFour := fmt.Sprintf("%02d", 4)
	week4filepath := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekFour, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week4, err := underlyingFs.Create(week4filepath)
	require.NoError(t, err)
	_, err = week4.WriteString(week4logLinesStr)
	require.NoError(t, err)

	formattedWeekThree := fmt.Sprintf("%02d", 3)
	week3filepath := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekThree, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week3, err := underlyingFs.Create(week3filepath)
	require.NoError(t, err)
	_, err = week3.WriteString(week3logLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClock(defaultYear, 4, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}

}

func TestStreamUserServiceLogsPerWeek_WithLogLineAcrossWeeks(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

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

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()

	week3logLinesStr := strings.Join(week3logLines, "\n") + "\n"
	week4logLinesStr := strings.Join(week4logLines, "\n") + "\n"

	formattedWeekFour := fmt.Sprintf("%02d", 4)
	week4filepath := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekFour, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week4, err := underlyingFs.Create(week4filepath)
	require.NoError(t, err)
	_, err = week4.WriteString(week4logLinesStr)
	require.NoError(t, err)

	formattedWeekThree := fmt.Sprintf("%02d", 3)
	week3filepath := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekThree, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week3, err := underlyingFs.Create(week3filepath)
	require.NoError(t, err)
	_, err = week3.WriteString(week3logLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClock(defaultYear, 4, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedFirstLogLine, serviceLogLines[0].GetContent())
	}
}

func TestStreamUserServiceLogsPerWeekReturnsTimestampedLogLines(t *testing.T) {
	expectedAmountLogLines := 3

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	timedLogLine1 := fmt.Sprintf("{\"log\":\"Starting feature 'centralized logs'\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)
	timedLogLine2 := fmt.Sprintf("{\"log\":\"Starting feature 'runs idempotently'\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)
	timedLogLine3 := fmt.Sprintf("{\"log\":\"The enclave was created\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)

	timestampedLogLines := []string{timedLogLine1, timedLogLine2, timedLogLine3}
	timestampedLogLinesStr := strings.Join(timestampedLogLines, "\n") + "\n"

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()

	formattedWeekNum := fmt.Sprintf("%02d", startingWeek)
	filepath := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file, err := underlyingFs.Create(filepath)
	require.NoError(t, err)
	_, err = file.WriteString(timestampedLogLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClock(defaultYear, startingWeek, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	expectedTime, err := time.Parse(utcFormat, defaultUTCTimestampStr)
	require.NoError(t, err)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perWeekStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		for _, logLine := range serviceLogLines {
			require.Equal(t, expectedTime, logLine.GetTimestamp())
		}
	}
}

func TestStreamUserServiceLogsPerFileReturnsTimestampedLogLines(t *testing.T) {
	expectedAmountLogLines := 3

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	timedLogLine1 := fmt.Sprintf("{\"log\":\"Starting feature 'centralized logs'\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)
	timedLogLine2 := fmt.Sprintf("{\"log\":\"Starting feature 'runs idempotently'\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)
	timedLogLine3 := fmt.Sprintf("{\"log\":\"The enclave was created\", \"timestamp\":\"%v\"}", defaultUTCTimestampStr)

	timestampedLogLines := []string{timedLogLine1, timedLogLine2, timedLogLine3}
	timestampedLogLinesStr := strings.Join(timestampedLogLines, "\n") + "\n"

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()

	filepath := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file, err := underlyingFs.Create(filepath)
	require.NoError(t, err)
	_, err = file.WriteString(timestampedLogLinesStr)
	require.NoError(t, err)

	perFileStreamStrategy := stream_logs_strategy.NewPerFileStreamLogsStrategy()

	expectedTime, err := time.Parse(utcFormat, defaultUTCTimestampStr)
	require.NoError(t, err)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perFileStreamStrategy,
	)
	require.NoError(t, testEvaluationErr)

	for serviceUuid, serviceLogLines := range receivedUserServiceLogsByUuid {
		expectedAmountLogLines, found := expectedServiceAmountLogLinesByServiceUuid[serviceUuid]
		require.True(t, found)
		require.Equal(t, expectedAmountLogLines, len(serviceLogLines))
		require.Equal(t, expectedTime, serviceLogLines[0].GetTimestamp())
	}
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
	underlyingFs volume_filesystem.VolumeFilesystem,
	streamStrategy stream_logs_strategy.StreamLogsStrategy,
) (map[service.ServiceUUID][]logline.LogLine, error) {
	ctx := context.Background()

	receivedServiceLogsByUuid := map[service.ServiceUUID][]logline.LogLine{}

	for serviceUuid := range expectedServiceAmountLogLinesByServiceUuid {
		receivedServiceLogsByUuid[serviceUuid] = []logline.LogLine{}
	}

	kurtosisBackend := backend_interface.NewMockKurtosisBackend(t)

	// no log file management is done in these tests so values for logFileManager aren't important
	mockTime := logs_clock.NewMockLogsClock(0, 0, 0)
	fileLayout := file_layout.NewPerWeekFileLayout(mockTime)
	logFileManager := log_file_manager.NewLogFileManager(kurtosisBackend, underlyingFs, fileLayout, mockTime, 0)
	logsDatabaseClient := NewPersistentVolumeLogsDatabaseClient(kurtosisBackend, underlyingFs, logFileManager, streamStrategy)

	userServiceLogsByUuidChan, errChan, receivedCancelCtxFunc, err := logsDatabaseClient.StreamUserServiceLogs(ctx, enclaveUuid, userServiceUuids, logLinesFilters, shouldFollowLogs, defaultShouldReturnAllLogs, defaultNumLogLines)
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
			if !isChanOpen && len(userServiceLogsByUuidChan) == 0 {
				shouldReceiveStream = false
				break
			}
			if isChanOpen && streamErr != nil {
				return nil, stacktrace.Propagate(streamErr, "Receiving streaming error.")
			}
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

func createFilledPerFileFilesystem() volume_filesystem.VolumeFilesystem {
	logLines := []string{logLine1, logLine2, logLine3a, logLine3b, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")

	file1PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file2PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService2Uuid, volume_consts.Filetype)
	file3PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService3Uuid, volume_consts.Filetype)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	file1, _ := mapFs.Create(file1PathStr)
	_, _ = file1.WriteString(logLinesStr)

	file2, _ := mapFs.Create(file2PathStr)
	_, _ = file2.WriteString(logLinesStr)

	file3, _ := mapFs.Create(file3PathStr)
	_, _ = file3.WriteString(logLinesStr)

	return mapFs
}

func createFilledPerWeekFilesystem(week int) volume_filesystem.VolumeFilesystem {
	logLines := []string{logLine1, logLine2, logLine3a, logLine3b, logLine4, logLine5, logLine6, logLine7, logLine8}

	logLinesStr := strings.Join(logLines, "\n")
	// %02d to format week num with leading zeros so 1-9 are converted to 01-09 for %V format
	formattedWeekNum := fmt.Sprintf("%02d", week)
	file1PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file2PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService2Uuid, volume_consts.Filetype)
	file3PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService3Uuid, volume_consts.Filetype)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	file1, _ := mapFs.Create(file1PathStr)
	_, _ = file1.WriteString(logLinesStr)

	file2, _ := mapFs.Create(file2PathStr)
	_, _ = file2.WriteString(logLinesStr)

	file3, _ := mapFs.Create(file3PathStr)
	_, _ = file3.WriteString(logLinesStr)

	return mapFs
}

func createEmptyPerFileFilesystem() volume_filesystem.VolumeFilesystem {
	file1PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file2PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService2Uuid, volume_consts.Filetype)
	file3PathStr := fmt.Sprintf(volume_consts.PerFileFmtStr, volume_consts.LogsStorageDirpath, testEnclaveUuid, testUserService3Uuid, volume_consts.Filetype)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	_, _ = mapFs.Create(file1PathStr)
	_, _ = mapFs.Create(file2PathStr)
	_, _ = mapFs.Create(file3PathStr)

	return mapFs
}

func createEmptyPerWeekFilesystem(week int) volume_filesystem.VolumeFilesystem {
	// %02d to format week num with leading zeros so 1-9 are converted to 01-09 for %V format
	formattedWeekNum := fmt.Sprintf("%02d", week)
	file1PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	file2PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService2Uuid, volume_consts.Filetype)
	file3PathStr := fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekNum, testEnclaveUuid, testUserService3Uuid, volume_consts.Filetype)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	_, _ = mapFs.Create(file1PathStr)
	_, _ = mapFs.Create(file2PathStr)
	_, _ = mapFs.Create(file3PathStr)

	return mapFs
}
