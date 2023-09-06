package stream_logs_strategy

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"testing/fstest"
)

const (
	logsStorageDirpathForTests = "var/log/kurtosis/"

	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	defaultRetentionPeriodInWeeks = volume_consts.LogRetentionPeriodInWeeks

	defaultYear = 2023
	defaultDay  = 0 // sunday
)

func TestGetRetainedLogsFilePaths(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week12filepath := getWeekFilepathStr(defaultYear, 12)
	week13filepath := getWeekFilepathStr(defaultYear, 13)
	week14filepath := getWeekFilepathStr(defaultYear, 14)
	week15filepath := getWeekFilepathStr(defaultYear, 15)
	week16filepath := getWeekFilepathStr(defaultYear, 16)
	week17filepath := getWeekFilepathStr(defaultYear, 17)

	mapFS := &fstest.MapFS{
		week12filepath: {
			Data: []byte{},
		},
		week13filepath: {
			Data: []byte{},
		},
		week14filepath: {
			Data: []byte{},
		},
		week15filepath: {
			Data: []byte{},
		},
		week16filepath: {
			Data: []byte{},
		},
		week17filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 17

	expectedLogFilePaths := []string{
		week13filepath,
		week14filepath,
		week15filepath,
		week16filepath,
		week17filepath,
	}

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths := strategy.getRetainedLogsFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetRetainedLogsFilePathsAcrossNewYear(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week50filepath := getWeekFilepathStr(defaultYear-1, 50)
	week51filepath := getWeekFilepathStr(defaultYear-1, 51)
	week52filepath := getWeekFilepathStr(defaultYear-1, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	mapFS := &fstest.MapFS{
		week50filepath: {
			Data: []byte{},
		},
		week51filepath: {
			Data: []byte{},
		},
		week52filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 2

	expectedLogFilePaths := []string{
		week50filepath,
		week51filepath,
		week52filepath,
		week1filepath,
		week2filepath,
	}

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths := strategy.getRetainedLogsFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetRetainedLogsFilePathsWithDiffRetentionPeriod(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear-1, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	mapFS := &fstest.MapFS{
		week52filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 2
	retentionPeriod := 2

	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths := strategy.getRetainedLogsFilePaths(filesystem, retentionPeriod, testEnclaveUuid, testUserService1Uuid)

	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetRetainedLogsFilePathsReturnsErrorIfWeeksMissing(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear, 52)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	mapFS := &fstest.MapFS{
		week52filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths := strategy.getRetainedLogsFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.Less(t, len(logFilePaths), defaultRetentionPeriodInWeeks)
}

func TestGetRetainedLogsFilePathsReturnsCorrectPathsIfWeeksMissing(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week52filepath := getWeekFilepathStr(defaultYear, 0)
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week3filepath := getWeekFilepathStr(defaultYear, 3)

	mapFS := &fstest.MapFS{
		week52filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week3filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 3

	// should only return week 3
	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths := strategy.getRetainedLogsFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.Len(t, logFilePaths, 1)
	require.Equal(t, "/"+week3filepath, logFilePaths[0])
}

func TestIsWithinRetentionPeriod(t *testing.T) {
	// this is the 36th week of the yera
	jsonLogLine := map[string]string{
		"timestamp": "2023-09-06T00:35:15-04:00",
	}

	// week 41 would put the log line outside the retention period
	mockTime := logs_clock.NewMockLogsClock(2023, 41, 0)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)

	isWithinRetentionPeriod, err := strategy.isWithinRetentionPeriod(jsonLogLine)

	require.NoError(t, err)
	require.False(t, isWithinRetentionPeriod)
}

func getWeekFilepathStr(year, week int) string {
	return fmt.Sprintf(volume_consts.PerWeekFmtStr, logsStorageDirpathForTests, strconv.Itoa(year), strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}
