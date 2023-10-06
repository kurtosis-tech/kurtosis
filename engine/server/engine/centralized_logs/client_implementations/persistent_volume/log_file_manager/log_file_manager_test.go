package log_file_manager

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Uuid = "test-user-service-1"

	defaultDay = 0
)

func TestRemoveLogsBeyondRetentionPeriod(t *testing.T) {
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()
	mockKurtosisBackend := backend_interface.NewMockKurtosisBackend(t)

	week49filepath := getFilepathStr(2022, 49, testEnclaveUuid, testUserService1Uuid)
	week50filepath := getFilepathStr(2022, 50, testEnclaveUuid, testUserService1Uuid)
	week51filepath := getFilepathStr(2022, 51, testEnclaveUuid, testUserService1Uuid)
	week52filepath := getFilepathStr(2022, 52, testEnclaveUuid, testUserService1Uuid)
	week1filepath := getFilepathStr(2023, 1, testEnclaveUuid, testUserService1Uuid)
	week2filepath := getFilepathStr(2023, 2, testEnclaveUuid, testUserService1Uuid)

	_, _ = mockFs.Create(week49filepath)
	_, _ = mockFs.Create(week50filepath)
	_, _ = mockFs.Create(week51filepath)
	_, _ = mockFs.Create(week52filepath)
	_, _ = mockFs.Create(week1filepath)
	_, _ = mockFs.Create(week2filepath)

	currentWeek := 2

	mockTime := logs_clock.NewMockLogsClock(2023, currentWeek, defaultDay)
	logFileManager := NewLogFileManager(mockKurtosisBackend, mockFs, mockTime)

	// should remove week 49 logs
	logFileManager.RemoveLogsBeyondRetentionPeriod()

	_, err := mockFs.Stat(week49filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}
