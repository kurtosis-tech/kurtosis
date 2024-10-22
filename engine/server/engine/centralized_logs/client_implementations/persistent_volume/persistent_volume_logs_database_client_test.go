package persistent_volume

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/file_layout"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/persistent_volume_helpers"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/stream_logs_strategy"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
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
	defaultHour  = 5
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

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, startingWeek, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	underlyingFs := createFilledFilesystem(perWeekFileLayout, mockTime.Now())
	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHour_WithFilters(t *testing.T) {
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

	mockTime := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, defaultHour)
	perHourFileLayout := file_layout.NewPerHourFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	underlyingFs := createFilledFilesystem(perHourFileLayout, mockTime.Now())
	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, startingWeek, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	underlyingFs := createEmptyFilesystem(perWeekFileLayout, mockTime.Now())
	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHour_NoLogsFromPersistentVolume(t *testing.T) {
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

	mockTime := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, defaultHour)
	perHourFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	underlyingFs := createEmptyFilesystem(perHourFileLayout, mockTime.Now())
	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, startingWeek, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	file1PathStr := perWeekFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHour_ThousandsOfLogLinesSuccessfulExecution(t *testing.T) {
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
	mockTime := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, defaultHour)
	perHourFileLayout := file_layout.NewPerHourFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	file1PathStr := perHourFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, startingWeek, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	file1PathStr := perWeekFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHour_EmptyLogLines(t *testing.T) {
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
	mockTime := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, defaultHour)
	perHourFileLayout := file_layout.NewPerHourFileLayout(mockTime, volume_consts.LogsStorageDirpath)
	file1PathStr := perHourFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file1, err := underlyingFs.Create(file1PathStr)
	require.NoError(t, err)
	_, err = file1.WriteString(logLinesStr)
	require.NoError(t, err)

	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		emptyFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, 4, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	week3logLinesStr := strings.Join(week3logLines, "\n") + "\n"
	week4logLinesStr := strings.Join(week4logLines, "\n")

	week4filepath := perWeekFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	week4, err := underlyingFs.Create(week4filepath)
	require.NoError(t, err)
	_, err = week4.WriteString(week4logLinesStr)
	require.NoError(t, err)

	week3Time := logs_clock.NewMockLogsClockPerDay(defaultYear, 3, defaultDay)
	week3filepath := perWeekFileLayout.GetLogFilePath(week3Time.Now(), testEnclaveUuid, testUserService1Uuid)
	week3, err := underlyingFs.Create(week3filepath)
	require.NoError(t, err)
	_, err = week3.WriteString(week3logLinesStr)
	require.NoError(t, err)

	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHour_WithLogsAcrossHours(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	hour4logLines := []string{
		logLine5,
		logLine6,
		logLine7,
		logLine8}
	hour3logLines := []string{
		logLine1,
		logLine2,
		logLine3a,
		logLine3b,
		logLine4}

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()
	hour4Time := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, 4)
	perHourFileLayout := file_layout.NewPerHourFileLayout(hour4Time, volume_consts.LogsStorageDirpath)

	hour3logLinesStr := strings.Join(hour3logLines, "\n") + "\n"
	hour4logLinesStr := strings.Join(hour4logLines, "\n")

	hour4filepath := perHourFileLayout.GetLogFilePath(hour4Time.Now(), testEnclaveUuid, testUserService1Uuid)
	hour4, err := underlyingFs.Create(hour4filepath)
	require.NoError(t, err)
	_, err = hour4.WriteString(hour4logLinesStr)
	require.NoError(t, err)

	hour3Time := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, 3)
	hour3filepath := perHourFileLayout.GetLogFilePath(hour3Time.Now(), testEnclaveUuid, testUserService1Uuid)
	hour3, err := underlyingFs.Create(hour3filepath)
	require.NoError(t, err)
	_, err = hour3.WriteString(hour3logLinesStr)
	require.NoError(t, err)

	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(hour4Time, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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
	week4filepath := fmt.Sprintf(file_layout.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekFour, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week4, err := underlyingFs.Create(week4filepath)
	require.NoError(t, err)
	_, err = week4.WriteString(week4logLinesStr)
	require.NoError(t, err)

	formattedWeekThree := fmt.Sprintf("%02d", 3)
	week3filepath := fmt.Sprintf(file_layout.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(defaultYear), formattedWeekThree, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
	week3, err := underlyingFs.Create(week3filepath)
	require.NoError(t, err)
	_, err = week3.WriteString(week3logLinesStr)
	require.NoError(t, err)

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, 4, defaultDay)
	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath))

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

func TestStreamUserServiceLogsPerHour_WithLogLineAcrossHours(t *testing.T) {
	expectedAmountLogLines := 8

	expectedServiceAmountLogLinesByServiceUuid := map[service.ServiceUUID]int{
		testUserService1Uuid: expectedAmountLogLines,
	}

	var logLinesFilters []logline.LogLineFilter

	userServiceUuids := map[service.ServiceUUID]bool{
		testUserService1Uuid: true,
	}

	expectedFirstLogLine := "Starting feature 'centralized logs'"

	hour4logLines := []string{
		logLine3b,
		logLine4,
		logLine5,
		logLine6,
		logLine7,
		logLine8}
	hour3logLines := []string{
		logLine1,
		logLine2,
		logLine3a}

	underlyingFs := volume_filesystem.NewMockedVolumeFilesystem()
	hour4Time := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, 4)
	perHourFileLayout := file_layout.NewPerHourFileLayout(hour4Time, volume_consts.LogsStorageDirpath)

	hour3logLinesStr := strings.Join(hour3logLines, "\n") + "\n"
	hour4logLinesStr := strings.Join(hour4logLines, "\n")

	hour4filepath := perHourFileLayout.GetLogFilePath(hour4Time.Now(), testEnclaveUuid, testUserService1Uuid)
	hour4, err := underlyingFs.Create(hour4filepath)
	require.NoError(t, err)
	_, err = hour4.WriteString(hour4logLinesStr)
	require.NoError(t, err)

	hour3Time := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, 3)
	hour3filepath := perHourFileLayout.GetLogFilePath(hour3Time.Now(), testEnclaveUuid, testUserService1Uuid)
	hour3, err := underlyingFs.Create(hour3filepath)
	require.NoError(t, err)
	_, err = hour3.WriteString(hour3logLinesStr)
	require.NoError(t, err)

	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(hour4Time, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, startingWeek, defaultDay)
	perWeekFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	filepath := perWeekFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file, err := underlyingFs.Create(filepath)
	require.NoError(t, err)
	_, err = file.WriteString(timestampedLogLinesStr)
	require.NoError(t, err)

	perWeekStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perWeekFileLayout)

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

func TestStreamUserServiceLogsPerHourReturnsTimestampedLogLines(t *testing.T) {
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
	mockTime := logs_clock.NewMockLogsClockPerHour(defaultYear, startingWeek, defaultDay, defaultHour)
	perHourFileLayout := file_layout.NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	filepath := perHourFileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	file, err := underlyingFs.Create(filepath)
	require.NoError(t, err)
	_, err = file.WriteString(timestampedLogLinesStr)
	require.NoError(t, err)

	perHourStreamStrategy := stream_logs_strategy.NewStreamLogsStrategyImpl(mockTime, persistent_volume_helpers.ConvertWeeksToDuration(retentionPeriodInWeeksForTesting), perHourFileLayout)

	expectedTime, err := time.Parse(utcFormat, defaultUTCTimestampStr)
	require.NoError(t, err)

	receivedUserServiceLogsByUuid, testEvaluationErr := executeStreamCallAndGetReceivedServiceLogLines(
		t,
		logLinesFilters,
		userServiceUuids,
		expectedServiceAmountLogLinesByServiceUuid,
		doNotFollowLogs,
		underlyingFs,
		perHourStreamStrategy,
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

	// no log file management is done in these tests so values provided to logFileManager aren't important
	logFileManager := log_file_manager.NewLogFileManager(nil, nil, nil, nil, time.Duration(0), "")
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

func createFilledFilesystem(fileLayout file_layout.LogFileLayout, time time.Time) volume_filesystem.VolumeFilesystem {
	logLines := []string{logLine1, logLine2, logLine3a, logLine3b, logLine4, logLine5, logLine6, logLine7, logLine8}
	logLinesStr := strings.Join(logLines, "\n")

	file1PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService1Uuid)
	file2PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService2Uuid)
	file3PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService3Uuid)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	file1, _ := mapFs.Create(file1PathStr)
	_, _ = file1.WriteString(logLinesStr)

	file2, _ := mapFs.Create(file2PathStr)
	_, _ = file2.WriteString(logLinesStr)

	file3, _ := mapFs.Create(file3PathStr)
	_, _ = file3.WriteString(logLinesStr)

	return mapFs
}

func createEmptyFilesystem(fileLayout file_layout.LogFileLayout, time time.Time) volume_filesystem.VolumeFilesystem {
	file1PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService1Uuid)
	file2PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService2Uuid)
	file3PathStr := fileLayout.GetLogFilePath(time, testEnclaveUuid, testUserService3Uuid)

	mapFs := volume_filesystem.NewMockedVolumeFilesystem()

	_, _ = mapFs.Create(file1PathStr)
	_, _ = mapFs.Create(file2PathStr)
	_, _ = mapFs.Create(file3PathStr)

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
