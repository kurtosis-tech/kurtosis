package stream_logs_strategy

import (
	"bufio"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/file_layout"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/stretchr/testify/require"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	retentionPeriodInWeeksForTesting = 5

	defaultYear = 2023
	defaultDay  = 0 // sunday
)

// TODO: migrate GetLogFilePaths tests to FileLayout interface when it is fully merged
// for now, leave them duplicated so there's an extra layer of testing as the migration happens
func TestGetLogFilePaths(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week12filepath := getWeekFilepathStr(defaultYear, 12)
	week13filepath := getWeekFilepathStr(defaultYear, 13)
	week14filepath := getWeekFilepathStr(defaultYear, 14)
	week15filepath := getWeekFilepathStr(defaultYear, 15)
	week16filepath := getWeekFilepathStr(defaultYear, 16)
	week17filepath := getWeekFilepathStr(defaultYear, 17)

	_, _ = filesystem.Create(week12filepath)
	_, _ = filesystem.Create(week13filepath)
	_, _ = filesystem.Create(week14filepath)
	_, _ = filesystem.Create(week15filepath)
	_, _ = filesystem.Create(week16filepath)
	_, _ = filesystem.Create(week17filepath)

	currentWeek := 17

	expectedLogFilePaths := []string{
		week13filepath,
		week14filepath,
		week15filepath,
		week16filepath,
		week17filepath,
	}

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYear(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week50filepath := getWeekFilepathStr(defaultYear-1, 50)
	week51filepath := getWeekFilepathStr(defaultYear-1, 51)
	week52filepath := getWeekFilepathStr(defaultYear-1, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	currentWeek := 2

	expectedLogFilePaths := []string{
		week50filepath,
		week51filepath,
		week52filepath,
		week1filepath,
		week2filepath,
	}

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYearWith53Weeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// According to ISOWeek, 2015 has 53 weeks
	week52filepath := getWeekFilepathStr(2015, 52)
	week53filepath := getWeekFilepathStr(2015, 53)
	week1filepath := getWeekFilepathStr(2016, 1)
	week2filepath := getWeekFilepathStr(2016, 2)
	week3filepath := getWeekFilepathStr(2016, 3)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week53filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)
	_, _ = filesystem.Create(week3filepath)

	currentWeek := 3

	expectedLogFilePaths := []string{
		week52filepath,
		week53filepath,
		week1filepath,
		week2filepath,
		week3filepath,
	}

	mockTime := logs_clock.NewMockLogsClockPerDay(2016, currentWeek, 1)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithDiffRetentionPeriod(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear-1, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	currentWeek := 2
	retentionPeriod := 3

	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriod, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsAllAvailableWeeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear-1, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	// should return existing file paths even though log files going all the back to retention period don't exist
	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}

	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Less(t, len(logFilePaths), retentionPeriodInWeeksForTesting)
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsCorrectPathsIfWeeksMissingInBetween(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear, 0)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week3filepath := getWeekFilepathStr(defaultYear, 3)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week3filepath)

	currentWeek := 3

	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, week3filepath, logFilePaths[0]) // should only return week 3 because week 2 is missing
}

func TestGetLogFilePathsReturnsCorrectPathsIfCurrentWeekHasNoLogsYet(t *testing.T) {
	// currently in week 3
	currentWeek := 3
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)

	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	// ../week/enclave uuid/service uuid.json
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	// no logs for week current week exist yet
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	// should return week 1 and 2 logs, even though no logs for current week yet
	expectedLogFilePaths := []string{
		week1filepath,
		week2filepath,
	}

	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriodInWeeksForTesting, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestIsWithinRetentionPeriod(t *testing.T) {
	// this is the 36th week of the year
	jsonLogLine := map[string]string{
		"timestamp": "2023-09-06T00:35:15-04:00",
	}

	// week 41 would put the log line outside the retention period
	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 41, 0)
	strategy := NewPerWeekStreamLogsStrategy(mockTime, retentionPeriodInWeeksForTesting)

	timestamp, err := parseTimestampFromJsonLogLine(jsonLogLine)
	require.NoError(t, err)
	logLine := logline.NewLogLine("", *timestamp)

	isWithinRetentionPeriod, err := strategy.isWithinRetentionPeriod(logLine)

	require.NoError(t, err)
	require.False(t, isWithinRetentionPeriod)
}

func getWeekFilepathStr(year, week int) string {
	// %02d to format week num with leading zeros so 1-9 are converted to 01-09 for %V format
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf(file_layout.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}

func TestGetCompleteJsonLogString(t *testing.T) {
	logLine1 := "{\"log\":\"Starting feature 'runs idempotently'\"}"
	logLine2a := "{\"log\":\"Starting feature 'apic "
	logLine2b := "idempotently'\"}"

	logs := strings.Join([]string{logLine1, logLine2a, logLine2b}, string(volume_consts.NewLineRune))
	logsReader := bufio.NewReader(strings.NewReader(logs))

	var jsonLogStr string
	var err error

	// First read
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.NoError(t, err)
	require.Equal(t, logLine1, jsonLogStr)

	// Second read
	logLine2 := "{\"log\":\"Starting feature 'apic idempotently'\"}"
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.Equal(t, logLine2, jsonLogStr)
}

func TestGetCompleteJsonLogStringAcrossManyCompleteLines(t *testing.T) {
	logLine1 := "{\"log\":\"Starting feature 'files manager'\"}"
	logLine2 := "{\"log\":\"The enclave was created\"}"
	logLine3 := "{\"log\":\"User service started\"}"
	logLine4 := "{\"log\":\"The data have being loaded\"}"

	logs := strings.Join([]string{logLine1, logLine2, logLine3, logLine4}, string(volume_consts.NewLineRune))
	logsReader := bufio.NewReader(strings.NewReader(logs))

	var jsonLogStr string
	var err error

	// First read
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.NoError(t, err)
	require.Equal(t, logLine1, jsonLogStr)

	// Second read
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.NoError(t, err)
	require.Equal(t, logLine2, jsonLogStr)

	// Fourth read
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.NoError(t, err)
	require.Equal(t, logLine3, jsonLogStr)

	// Last read
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.Equal(t, logLine4, jsonLogStr)
}

func TestGetCompleteJsonLogStringAcrossManyBrokenLines(t *testing.T) {
	logLine1a := "{\"log\":\"Starting"
	logLine1b := " feature "
	logLine1c := "'runs "
	logLine1d := "idempotently'\"}"

	logs := strings.Join([]string{logLine1a, logLine1b, logLine1c, logLine1d}, string(volume_consts.NewLineRune))
	logsReader := bufio.NewReader(strings.NewReader(logs))

	var jsonLogStr string
	var err error

	logLine1 := "{\"log\":\"Starting feature 'runs idempotently'\"}"
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.Equal(t, logLine1, jsonLogStr)
}

func TestGetCompleteJsonLogStringWithNoValidJsonEnding(t *testing.T) {
	logLine1 := "{\"log\":\"Starting idempotently'\""

	logsReader := bufio.NewReader(strings.NewReader(logLine1))

	var jsonLogStr string
	var err error

	// this will end up in an infinite loop, bc [getCompleteJsonLogString] keeps looping till it finds EOF or complete json
	jsonLogStr, err = getCompleteJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.Equal(t, logLine1, jsonLogStr)
}

func TestGetJsonLogString(t *testing.T) {
	logLine1 := "{\"log\":\"Starting feature 'centralized logs'\"}"
	logLine2 := "{\"log\":\"Starting feature 'runs idempotently'\"}"
	logLine3a := "{\"log\":\"Starting feature 'apic "
	logLine3b := "idempotently'\"}"

	logs := strings.Join([]string{logLine1, logLine2, logLine3a, logLine3b}, string(volume_consts.NewLineRune))
	logsReader := bufio.NewReader(strings.NewReader(logs))

	var jsonLogStr string
	var isComplete bool
	var err error

	// First read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.NoError(t, err)
	require.True(t, isComplete)
	require.Equal(t, logLine1, jsonLogStr)

	// Second read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.NoError(t, err)
	require.True(t, isComplete)
	require.Equal(t, logLine2, jsonLogStr)

	// Third read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.NoError(t, err)
	require.False(t, isComplete)
	require.Equal(t, logLine3a, jsonLogStr)

	// Last read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.True(t, isComplete)
	require.Equal(t, logLine3b, jsonLogStr)
}

func TestGetJsonLogStringWithEOFAndNoNewLine(t *testing.T) {
	logLine1a := "{\"log\":\"Starting feature 'apic "
	logLine1b := "idempotently'\"}"

	logs := logLine1a + "\n" + logLine1b
	logsReader := bufio.NewReader(strings.NewReader(logs))

	var jsonLogStr string
	var isComplete bool
	var err error

	// First read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.NoError(t, err)
	require.False(t, isComplete)
	require.Equal(t, logLine1a, jsonLogStr)

	// Second read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.True(t, isComplete)
	require.Equal(t, logLine1b, jsonLogStr)
}

func TestGetJsonLogStringWithEOFAndNoValidJsonEnding(t *testing.T) {
	logLine1 := "{\"log\":\"Starting feature 'centralized logs'\""

	logsReader := bufio.NewReader(strings.NewReader(logLine1))

	var jsonLogStr string
	var isComplete bool
	var err error

	// First read
	jsonLogStr, isComplete, err = getJsonLogString(logsReader)
	require.Error(t, err)
	require.ErrorIs(t, io.EOF, err)
	require.False(t, isComplete)
	require.Equal(t, logLine1, jsonLogStr)
}

func TestParseTimestampFromJsonLogLineReturnsTime(t *testing.T) {
	timestampStr := "2023-09-06T00:35:15Z" // utc timestamp
	jsonLogLine := map[string]string{
		"timestamp": timestampStr,
	}

	expectedTime, err := time.Parse(time.RFC3339, timestampStr)
	require.NoError(t, err)

	actualTime, err := parseTimestampFromJsonLogLine(jsonLogLine)

	require.NoError(t, err)
	require.Equal(t, expectedTime, *actualTime)
}

func TestParseTimestampFromJsonLogLineWithOffsetReturnsTime(t *testing.T) {
	timestampStr := "2023-09-06T00:35:15-04:00" // utc timestamp with offset '-4:00'
	jsonLogLine := map[string]string{
		"timestamp": timestampStr,
	}

	expectedTime, err := time.Parse(time.RFC3339, timestampStr)
	require.NoError(t, err)

	actualTime, err := parseTimestampFromJsonLogLine(jsonLogLine)

	require.NoError(t, err)
	require.Equal(t, expectedTime, *actualTime)
}

func TestParseTimestampFromJsonLogLineWithIncorrectlyFormattedTimeReturnsError(t *testing.T) {
	timestampStr := "2023-09-06" // not UTC formatted timestamp str
	jsonLogLine := map[string]string{
		"timestamp": timestampStr,
	}

	_, err := parseTimestampFromJsonLogLine(jsonLogLine)

	require.Error(t, err)
}

func TestParseTimestampFromJsonLogLineWithoutTimezoneReturnsError(t *testing.T) {
	timestampStr := "2023-09-06T00:35:15" // no utc timezone indicator or offset to indicate timezone
	jsonLogLine := map[string]string{
		"timestamp": timestampStr,
	}

	_, err := parseTimestampFromJsonLogLine(jsonLogLine)

	require.Error(t, err)
}

func TestParseTimestampFromJsonLogLineWithNoTimestampFieldReturnsError(t *testing.T) {
	jsonLogLine := map[string]string{}

	_, err := parseTimestampFromJsonLogLine(jsonLogLine)

	require.Error(t, err)
}
