package stream_logs_strategy

import (
	"bufio"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/file_layout"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"time"
)

const (
	retentionPeriodInWeeksForTesting = 5
)

func TestIsWithinRetentionPeriod(t *testing.T) {
	// this is the 36th week of the year
	jsonLogLine := map[string]string{
		"timestamp": "2023-09-06T00:35:15-04:00",
	}

	// week 41 would put the log line outside the retention period
	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 41, 0)
	strategy := NewStreamLogsStrategyImpl(mockTime, convertWeeksToDuration(retentionPeriodInWeeksForTesting), file_layout.NewPerWeekFileLayout(mockTime))

	timestamp, err := parseTimestampFromJsonLogLine(jsonLogLine)
	require.NoError(t, err)
	logLine := logline.NewLogLine("", *timestamp)

	isWithinRetentionPeriod, err := strategy.isWithinRetentionPeriod(logLine)

	require.NoError(t, err)
	require.False(t, isWithinRetentionPeriod)
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

func convertWeeksToDuration(retentionPeriodInWeeks int) time.Duration {
	const hoursInWeek = 7 * 24 // 7 days * 24 hours
	return time.Duration(retentionPeriodInWeeks*hoursInWeek) * time.Hour
}
