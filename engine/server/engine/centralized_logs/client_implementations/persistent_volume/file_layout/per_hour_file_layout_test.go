package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetLogFilePath(t *testing.T) {
	currentTime := logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 1)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	expectedFilepath := "/var/log/kurtosis/2024/01/1/01/test-enclave/test-user-service-1.json"
	now := currentTime.Now()
	actualFilePath := fileLayout.GetLogFilePath(now, testEnclaveUuid, testUserService1Uuid)
	require.Equal(t, expectedFilepath, actualFilePath)
}

func TestGetLogFilePathsWithHourlyRetention(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 5)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 2).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 3).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 4).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 5).Now(), testEnclaveUuid, testUserService1Uuid)

	createFilepaths(t, filesystem, []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	})

	expectedLogFilePaths := []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	}

	retentionPeriod := 6 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithHourlyRetentionAcrossDays(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 2, 2)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 1, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 1, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 1, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 2, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 2, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, 2, 2).Now(), testEnclaveUuid, testUserService1Uuid)

	createFilepaths(t, filesystem, []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	})

	expectedLogFilePaths := []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	}

	retentionPeriod := 6 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithHourlyRetentionAcrossWeeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClockPerHour(defaultYear, 18, 1, 2)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 17, 0, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 17, 0, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 17, 0, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 18, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 18, 1, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(defaultYear, 18, 1, 2).Now(), testEnclaveUuid, testUserService1Uuid)

	createFilepaths(t, filesystem, []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	})

	expectedLogFilePaths := []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	}

	retentionPeriod := 6 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithHourlyRetentionAcrossYears(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 2)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 2).Now(), testEnclaveUuid, testUserService1Uuid)

	createFilepaths(t, filesystem, []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	})

	expectedLogFilePaths := []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFourFp,
		hourFiveFp,
	}

	retentionPeriod := 6 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}

}

func TestSundayIsConvertedFromStrftimeToGolangTime(t *testing.T) {
	expectedFilepath := "/var/log/kurtosis/2024/02/7/05/test-enclave/test-user-service-1.json"

	mockTime := logs_clock.NewMockLogsClockPerHour(2024, 2, 0, 5)
	fileLayout := NewPerHourFileLayout(mockTime, volume_consts.LogsStorageDirpath)

	actualFilePath := fileLayout.GetLogFilePath(mockTime.Now(), testEnclaveUuid, testUserService1Uuid)
	require.Equal(t, expectedFilepath, actualFilePath)
}

//func TestGetLogFilePathsWithHourlyRetentionReturnsCorrectPathsIfHoursMissingInBetween(t *testing.T) {
//	filesystem := volume_filesystem.NewMockedVolumeFilesystem()
//
//	currentTime := logs_clock.NewMockLogsClockPerHour(defaultYear, defaultWeek, defaultDay, 1)
//	fileLayout := NewPerWeekFileLayout(currentTime)
//
//	// ../week/enclave uuid/service uuid.json
//	week52filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 0, 0).Now(), testEnclaveUuid, testUserService1Uuid)
//	week1filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
//	week3filepath := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerDay(defaultYear, 3, 0).Now(), testEnclaveUuid, testUserService1Uuid)
//
//	_, _ = filesystem.Create(week52filepath)
//	_, _ = filesystem.Create(week1filepath)
//	_, _ = filesystem.Create(week3filepath)
//	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekDuration
//	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)
//
//	require.NoError(t, err)
//	require.Len(t, logFilePaths, 1)
//	require.Equal(t, week3filepath, logFilePaths[0]) // should only return week 3 because week 2 is missing
//}

func createFilepaths(t *testing.T, filesystem volume_filesystem.VolumeFilesystem, filepaths []string) {
	for _, path := range filepaths {
		_, err := filesystem.Create(path)
		require.NoError(t, err)
	}
}

func TestDurationToHour(t *testing.T) {

}
