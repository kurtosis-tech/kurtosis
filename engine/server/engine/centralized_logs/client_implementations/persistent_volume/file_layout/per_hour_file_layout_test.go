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

func TestGetLogFilePathsWithHourlyRetentionReturnsCorrectPathsIfHoursMissingInBetween(t *testing.T) {
	filesystem := volume_filesystem.NewMockedVolumeFilesystem()

	currentTime := logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 2)
	fileLayout := NewPerHourFileLayout(currentTime, volume_consts.LogsStorageDirpath)

	hourZeroFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 21).Now(), testEnclaveUuid, testUserService1Uuid)
	hourOneFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 22).Now(), testEnclaveUuid, testUserService1Uuid)
	hourTwoFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 52, 0, 23).Now(), testEnclaveUuid, testUserService1Uuid)
	hourThreeFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2023, 1, 1, 3).Now(), testEnclaveUuid, testUserService1Uuid)
	hourFiveFp := fileLayout.GetLogFilePath(logs_clock.NewMockLogsClockPerHour(2024, 1, 1, 2).Now(), testEnclaveUuid, testUserService1Uuid)

	createFilepaths(t, filesystem, []string{
		hourZeroFp,
		hourOneFp,
		hourTwoFp,
		hourThreeFp,
		hourFiveFp,
	})

	retentionPeriod := 6 * time.Hour // this would return all filepaths, but hour three is missing
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)
	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, hourFiveFp, logFilePaths[0]) // should only return hour 5 3 because hour 4 is missing
}

func TestTimeToWeekDayHour(t *testing.T) {
	tests := []struct {
		name         string
		inputTime    time.Time
		expectedYear int
		expectedWeek int
		expectedDay  int
		expectedHour int
	}{
		{
			name:         "Midweek Wednesday 14:00",
			inputTime:    time.Date(2023, 10, 18, 14, 0, 0, 0, time.UTC),
			expectedYear: 2023,
			expectedWeek: 42,
			expectedDay:  3,
			expectedHour: 14,
		},
		{
			name:         "Sunday midnight",
			inputTime:    time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC),
			expectedYear: 2023,
			expectedWeek: 41,
			expectedDay:  7, // Sunday should be converted to 7
			expectedHour: 0,
		},
		{
			name:         "Monday 9:30",
			inputTime:    time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedWeek: 1,
			expectedDay:  1,
			expectedHour: 9,
		},
		{
			name:         "Saturday afternoon",
			inputTime:    time.Date(2024, 10, 19, 15, 0, 0, 0, time.UTC),
			expectedYear: 2024,
			expectedWeek: 42,
			expectedDay:  6,
			expectedHour: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, week, day, hour := TimeToWeekDayHour(tt.inputTime)
			if year != tt.expectedYear || week != tt.expectedWeek || day != tt.expectedDay || hour != tt.expectedHour {
				t.Errorf("TimeToWeekDayHour(%v) = (%d, %d, %d, %d); expected (%d, %d, %d, %d)",
					tt.inputTime, year, week, day, hour, tt.expectedYear, tt.expectedWeek, tt.expectedDay, tt.expectedHour)
			}
		})
	}
}

func TestDurationToHours(t *testing.T) {
	tests := []struct {
		name          string
		inputDuration time.Duration
		expectedHours int
	}{
		{
			name:          "Zero duration",
			inputDuration: 0,
			expectedHours: 0,
		},
		{
			name:          "One hour duration",
			inputDuration: time.Hour,
			expectedHours: 1,
		},
		{
			name:          "Fractional hour duration",
			inputDuration: 90 * time.Minute, // 1.5 hours
			expectedHours: 2,                // should round up
		},
		{
			name:          "More than one day",
			inputDuration: 25 * time.Hour,
			expectedHours: 25,
		},
		{
			name:          "Negative duration",
			inputDuration: -time.Hour,
			expectedHours: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DurationToHours(tt.inputDuration)
			if result != tt.expectedHours {
				t.Errorf("DurationToHours(%v) = %d; expected %d", tt.inputDuration, result, tt.expectedHours)
			}
		})
	}
}

func createFilepaths(t *testing.T, filesystem volume_filesystem.VolumeFilesystem, filepaths []string) {
	for _, path := range filepaths {
		_, err := filesystem.Create(path)
		require.NoError(t, err)
	}
}
