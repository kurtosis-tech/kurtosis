package file_layout

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
	"time"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	retentionPeriodInWeeksForTesting = 5
	oneWeekInHours                   = 7 * 24 * time.Hour

	defaultYear = 2023
	defaultDay  = 0 // sunday
)

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

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

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

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

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

	mockTime := logs_clock.NewMockLogsClock(2016, currentWeek, 1)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

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

	expectedLogFilePaths := []string{
		week52filepath,
		week1filepath,
		week2filepath,
	}

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := 3 * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

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

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

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

	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)
	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Len(t, logFilePaths, 1)
	require.Equal(t, week3filepath, logFilePaths[0]) // should only return week 3 because week 2 is missing
}

func TestGetLogFilePathsReturnsCorrectPathsIfCurrentWeekHasNoLogsYet(t *testing.T) {
	// currently in week 3
	currentWeek := 3
	mockTime := logs_clock.NewMockLogsClock(defaultYear, currentWeek, defaultDay)

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

	fileLayout := NewPerWeekFileLayout(mockTime)
	retentionPeriod := retentionPeriodInWeeksForTesting * oneWeekInHours
	logFilePaths, err := fileLayout.GetLogFilePaths(filesystem, retentionPeriod, -1, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.Equal(t, len(expectedLogFilePaths), len(logFilePaths))
	for i, filePath := range expectedLogFilePaths {
		require.Equal(t, filePath, logFilePaths[i])
	}
}

func getWeekFilepathStr(year, week int) string {
	// %02d to format week num with leading zeros so 1-9 are converted to 01-09 for %V format
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, testEnclaveUuid, testUserService1Uuid, volume_consts.Filetype)
}
