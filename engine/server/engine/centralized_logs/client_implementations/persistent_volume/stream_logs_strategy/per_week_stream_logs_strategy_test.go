package stream_logs_strategy

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"testing/fstest"
)

const (
	logsStorageDirpathForTests = "var/log/kurtosis/"

	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"
)

func TestGetRetainedLogsFilepaths(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week12filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "12", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week13filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "13", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week14filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "14", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week15filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "15", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week16filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "16", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week17filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "17", testEnclaveUuid, testUserService1Uuid, consts.Filetype)

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

	expectedLogFilepaths := []string{
		week17filepath,
		week16filepath,
		week15filepath,
		week14filepath,
		week13filepath,
	}
	logFilepaths, err := getRetainedLogsFilepaths(filesystem, defaultRetentionPeriodInWeeks, currentWeek, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(expectedLogFilepaths, logFilepaths))
	require.Len(t, logFilepaths, defaultRetentionPeriodInWeeks+1)
}

func TestGetRetainedLogsFilepathsAcrossNewYear(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week52filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "52", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week53filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "53", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week0filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "0", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week1filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "1", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week2filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "2", testEnclaveUuid, testUserService1Uuid, consts.Filetype)

	mapFS := &fstest.MapFS{
		week52filepath: {
			Data: []byte{},
		},
		week53filepath: {
			Data: []byte{},
		},
		week0filepath: {
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

	expectedLogFilepaths := []string{
		week2filepath,
		week1filepath,
		week0filepath,
		week53filepath,
		week52filepath,
	}
	logFilepaths, err := getRetainedLogsFilepaths(filesystem, defaultRetentionPeriodInWeeks, currentWeek, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(expectedLogFilepaths, logFilepaths))
	require.Len(t, logFilepaths, defaultRetentionPeriodInWeeks+1)
}

func TestGetRetainedLogsFilepathsWithDiffRetentionPeriod(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week0filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "0", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week1filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "1", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week2filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "2", testEnclaveUuid, testUserService1Uuid, consts.Filetype)

	mapFS := &fstest.MapFS{
		week0filepath: {
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

	expectedLogFilepaths := []string{
		week2filepath,
		week1filepath,
		week0filepath,
	}
	logFilepaths, err := getRetainedLogsFilepaths(filesystem, retentionPeriod, currentWeek, testEnclaveUuid, testUserService1Uuid)

	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(expectedLogFilepaths, logFilepaths))
	require.Len(t, logFilepaths, retentionPeriod+1)
}

func TestGetRetainedLogsFilepathsReturnsErrorIfWeeksMissing(t *testing.T) {
	// ../week/enclave uuid/service uuid.json
	week0filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "0", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week1filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "1", testEnclaveUuid, testUserService1Uuid, consts.Filetype)
	week2filepath := fmt.Sprintf("%s%s%s/%s%s", logsStorageDirpathForTests, "2", testEnclaveUuid, testUserService1Uuid, consts.Filetype)

	mapFS := &fstest.MapFS{
		week0filepath: {
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

	logFilepaths, err := getRetainedLogsFilepaths(filesystem, defaultRetentionPeriodInWeeks, currentWeek, testEnclaveUuid, testUserService1Uuid)

	require.Error(t, err)
	require.Less(t, len(logFilepaths), defaultRetentionPeriodInWeeks+1)
}
