package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetLogFilePath(t *testing.T) {
	currentTime := logs_clock.NewMockLogsClock(2024, 1, 1, 1)
	fileLayout := NewPerHourFileLayout(currentTime)

	expectedFilepath := "/var/log/kurtosis/2024/01/01/01/test-enclave/test-user-service-1.json"
	now := currentTime.Now()
	actualFilePath := fileLayout.GetLogFilePath(now, testEnclaveUuid, testUserService1Uuid)
	require.Equal(t, expectedFilepath, actualFilePath)
}

func TestGetLogFilePathsWithHourlyRetention(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 5)
	fileLayout := NewPerHourFileLayout(currentTime)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 2).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 3).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 4).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, defaultDay, 5).Now(), testEnclaveUuid, testUserService1Uuid)

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

	currentTime := logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 2)
	fileLayout := NewPerHourFileLayout(currentTime)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 2).Now(), testEnclaveUuid, testUserService1Uuid)

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

	retentionPeriod := 5 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithHourlyRetentionAcrossWeeks(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 2)
	fileLayout := NewPerHourFileLayout(currentTime)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 0, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 0).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFourFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 1).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClock(defaultYear, defaultWeek, 1, 2).Now(), testEnclaveUuid, testUserService1Uuid)

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

	retentionPeriod := 5 * time.Hour // retention period of 6 hours should return all the file paths
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func TestGetLogFilePathsWithHourlyRetentionAcrossYears(t *testing.T) {

}

func createFilepaths(t *testing.T, filesystem volume_filesystem.VolumeFilesystem, filepaths []string) {
	for _, path := range filepaths {
		_, err := filesystem.Create(path)
		require.NoError(t, err)
	}
}

// test get log file paths per hour across year

// test get log file paths across days

// test get log file paths across weeks

// test get log file paths within the same day

// test getting log file paths across days
