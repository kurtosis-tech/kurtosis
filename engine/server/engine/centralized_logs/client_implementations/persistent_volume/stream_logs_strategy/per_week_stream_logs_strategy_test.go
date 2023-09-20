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
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	defaultRetentionPeriodInWeeks = volume_consts.LogRetentionPeriodInWeeks

	defaultYear = 2023
	defaultDay  = 0 // sunday
)

func TestGetLogFilePaths(t *testing.T) {
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
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYear(t *testing.T) {
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
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYearWith53Weeks(t *testing.T) {
	// According to ISOWeek, 2015 has 53 weeks
	week52filepath := getWeekFilepathStr(2015, 52)
	week53filepath := getWeekFilepathStr(2015, 53)
	week1filepath := getWeekFilepathStr(2016, 1)
	week2filepath := getWeekFilepathStr(2016, 2)
	week3filepath := getWeekFilepathStr(2016, 3)

	mapFS := &fstest.MapFS{
		week52filepath: {
			Data: []byte{},
		},
		week53filepath: {
			Data: []byte{},
		},
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
		week3filepath: {
			Data: []byte{},
		},
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 3

	expectedLogFilePaths := []string{
		week52filepath,
		week53filepath,
		week1filepath,
		week2filepath,
		week3filepath,
	}

	mockTime := logs_clock.NewMockLogsClock(2016, currentWeek, 1)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithDiffRetentionPeriod(t *testing.T) {
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
	logFilePaths, err := strategy.getLogFilePaths(filesystem, retentionPeriod, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsAllAvailableWeeks(t *testing.T) {
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

	// should return existing file paths even though log files going all the back to retention period don't exist
	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Less(t, len(logFilePaths), defaultRetentionPeriodInWeeks)
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsCorrectPathsIfWeeksMissingInBetween(t *testing.T) {
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

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, "/"+week3filepath, logFilePaths[0]) // should only return week 3 because week 2 is missing
}

func TestGetLogFilePathsReturnsCorrectPathsIfCurrentWeekHasNoLogsYet(t *testing.T) {
	// currently in week 3
	currentWeek := 3
	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)

	// ../week/enclave uuid/service uuid.json
	week1filepath := getWeekFilepathStr(defaultYear, 1)
	week2filepath := getWeekFilepathStr(defaultYear, 2)

	// no logs for week current week exist yet
	mapFS := &fstest.MapFS{
		week1filepath: {
			Data: []byte{},
		},
		week2filepath: {
			Data: []byte{},
		},
	}

	// should return week 1 and 2 logs, even though no logs for current week yet
	expectedLogFilePaths := []string{
		week1filepath,
		week2filepath,
	}

	filesystem := volume_filesystem.NewMockedVolumeFilesystem(mapFS)
	strategy := NewPerWeekStreamLogsStrategy(mockTime)
	logFilePaths, err := strategy.getLogFilePaths(filesystem, defaultRetentionPeriodInWeeks, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, "/"+filePath, logFilePaths[i])
	}
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
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpathForTests, strconv.Itoa(year), strconv.Itoa(week), testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}
