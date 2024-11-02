package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	retentionPeriodInWeeksForTesting = 5

	defaultYear = 2023
	defaultWeek = 17
	defaultDay  = 0 // sunday
)

func TestGetLogFilePaths(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 17
	currentTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	week12filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 12, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week13filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 13, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week14filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 14, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week15filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 15, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week16filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 16, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week17filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 17, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week12filepath)
	_, _ = filesystem.Create(week13filepath)
	_, _ = filesystem.Create(week14filepath)
	_, _ = filesystem.Create(week15filepath)
	_, _ = filesystem.Create(week16filepath)
	_, _ = filesystem.Create(week17filepath)

	expectedLogFilePaths := []string{
		week13filepath,
		week14filepath,
		week15filepath,
		week16filepath,
		week17filepath,
	}

	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYear(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 2
	currentTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	// ../week/enclave uuid/service uuid.json
	week50filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear-1, 50, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week51filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear-1, 51, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear-1, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	expectedLogFilePaths := []string{
		week50filepath,
		week51filepath,
		week52filepath,
		week1filepath,
		week2filepath,
	}

	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsAcrossNewYearWith53Weeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 3
	currentTime := logs_clock.NewMockLogsClockPerDay(2016, currentWeek, 1)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	// According to ISOWeek, 2015 has 53 weeks
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2015, 51, 3).Now(), testEnclaveUuid, testUserService1Uuid)
	week53filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2015, 52, 3).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2016, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2016, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week3filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2016, 3, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week53filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)
	_, _ = filesystem.Create(week3filepath)

	expectedLogFilePaths := []string{
		week52filepath,
		week53filepath,
		week1filepath,
		week2filepath,
		week3filepath,
	}

	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithDiffRetentionPeriod(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 2
	mockTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	// ../week/enclave uuid/service uuid.json
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear-1, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}
	retentionPeriod := 3 * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsAllAvailableWeeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 2
	currentTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	// ../week/enclave uuid/service uuid.json
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear-1, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	// should return existing file paths even though log files going all the back to retention period don't exist
	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Less(t, len(logFilePaths), retentionPeriodInWeeksForTesting)
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsReturnsCorrectPathsIfWeeksMissingInBetween(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 3
	currentTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	// ../week/enclave uuid/service uuid.json
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 0, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week3filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 3, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week3filepath)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, week3filepath, logFilePaths[0]) // should only return week 3 because week 2 is missing
}

func TestGetLogFilePathsReturnsCorrectPathsIfCurrentWeekHasNoLogsYet(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentWeek := 3
	currentTime := logs_clock.NewMockLogsClockPerDay(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	// ../week/enclave uuid/service uuid.json
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	// no logs for week current week exist yet
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	// should return week 1 and 2 logs, even though no logs for current week yet
	expectedLogFilePaths := []string{
		week1filepath,
		week2filepath,
	}

	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsOneIntervalBeyondRetentionPeriod(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 2, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	week49filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 49, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week50filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 50, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week51filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 51, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week49filepath)
	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	retentionPeriod := 5 * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, 1, testEnclaveUuid, testUserService1Uuid)
	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, logFilePaths[0], week49filepath)
}

func TestGetLogFilePathsTwoIntervalBeyondRetentionPeriod(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 2, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	week48filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 48, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week49filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 49, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week50filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 50, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week51filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 51, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week48filepath)
	_, _ = filesystem.Create(week49filepath)
	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	expectedLogFilePaths := []string{
		week49filepath,
		week48filepath,
	}

	retentionPeriod := 5 * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, 2, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 2)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithNoPathsBeyondRetentionPeriod(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 2, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	week50filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 50, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week51filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 51, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	retentionPeriod := 5 * oneWeekDuration
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, 1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 0)
}

func TestGetLogFilePathsWithMissingPathBetweenIntervals(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	mockTime := logs_clock.NewMockLogsClockPerDay(2023, 2, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	week47filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 48, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week49filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 49, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week50filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 50, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week51filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 51, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2022, 52, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	week2filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(2023, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)

	_, _ = filesystem.Create(week47filepath)
	// 48 is missing
	_, _ = filesystem.Create(week49filepath)
	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week50filepath)
	_, _ = filesystem.Create(week51filepath)
	_, _ = filesystem.Create(week52filepath)
	_, _ = filesystem.Create(week1filepath)
	_, _ = filesystem.Create(week2filepath)

	expectedLogFilePaths := []string{
		week49filepath,
		week47filepath,
	}

	retentionPeriod := 5 * oneWeekDuration
	// the expected behavior is return all filepaths, even if some are missing
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, 3, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 2)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}
