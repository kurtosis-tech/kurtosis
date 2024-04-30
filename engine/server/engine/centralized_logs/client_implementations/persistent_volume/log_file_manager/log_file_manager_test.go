package log_file_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"testing"
)

const (
	testEnclaveUuid      = "test-enclave"
	testUserService1Name = "test-user-service-1"
	testUserService1Uuid = "0e10c199bb1a4094839c3ebd432b2c49"

	defaultDay = 0
)

func TestRemoveLogsBeyondRetentionPeriod(t *testing.T) {
	mockKurtosisBackend := backend_interface.NewMockKurtosisBackend(t)
	mockTime := logs_clock.NewMockLogsClock(2023, 2, defaultDay)

	// setup filesystem
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()
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

	logFileManager := NewLogFileManager(mockKurtosisBackend, mockFs, mockTime, 5)
	logFileManager.RemoveLogsBeyondRetentionPeriod() // should remove week 49 logs

	_, err := mockFs.Stat(week49filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestRemoveEnclaveLogs(t *testing.T) {
	mockKurtosisBackend := backend_interface.NewMockKurtosisBackend(t)
	mockTime := logs_clock.NewMockLogsClock(2022, 52, defaultDay)

	// setup filesystem
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()

	week51filepath := getFilepathStr(2022, 51, testEnclaveUuid, testUserService1Uuid)
	week52filepathDiffEnclave := getFilepathStr(2022, 52, "enclaveOne", "serviceTwo")
	week52filepath := getFilepathStr(2022, 52, testEnclaveUuid, testUserService1Uuid)
	week52filepathDiffService := getFilepathStr(2022, 52, testEnclaveUuid, "serviceThree")

	_, _ = mockFs.Create(week51filepath)
	_, _ = mockFs.Create(week52filepathDiffEnclave)
	_, _ = mockFs.Create(week52filepath)
	_, _ = mockFs.Create(week52filepathDiffService)

	logFileManager := NewLogFileManager(mockKurtosisBackend, mockFs, mockTime, 5)
	err := logFileManager.RemoveEnclaveLogs(testEnclaveUuid) // should remove only all log files for enclave one

	require.NoError(t, err)

	_, err = mockFs.Stat(week52filepathDiffEnclave)
	require.NoError(t, err) // logs should still exist for different enclave

	_, err = mockFs.Stat(week52filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = mockFs.Stat(week51filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = mockFs.Stat(week52filepathDiffService)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestRemoveAllLogs(t *testing.T) {
	mockKurtosisBackend := backend_interface.NewMockKurtosisBackend(t)
	mockTime := logs_clock.NewMockLogsClock(2022, 52, defaultDay)

	// setup filesystem
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()

	week51filepath := getFilepathStr(2022, 51, testEnclaveUuid, testUserService1Uuid)
	week52filepathDiffEnclave := getFilepathStr(2022, 52, "enclaveOne", "serviceTwo")
	week52filepath := getFilepathStr(2022, 52, testEnclaveUuid, testUserService1Uuid)
	week52filepathDiffService := getFilepathStr(2022, 52, testEnclaveUuid, "serviceThree")

	_, _ = mockFs.Create(week51filepath)
	_, _ = mockFs.Create(week52filepathDiffEnclave)
	_, _ = mockFs.Create(week52filepath)
	_, _ = mockFs.Create(week52filepathDiffService)

	logFileManager := NewLogFileManager(mockKurtosisBackend, mockFs, mockTime, 5)
	err := logFileManager.RemoveAllLogs()

	require.NoError(t, err)

	_, err = mockFs.Stat(week52filepathDiffEnclave)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = mockFs.Stat(week52filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = mockFs.Stat(week51filepath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = mockFs.Stat(week52filepathDiffService)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestCreateLogFiles(t *testing.T) {
	mockTime := logs_clock.NewMockLogsClock(2022, 52, defaultDay)
	mockFs := volume_filesystem.NewMockedVolumeFilesystem()

	// setup kurtosis backend
	ctx := context.Background()
	mockKurtosisBackend := backend_interface.NewMockKurtosisBackend(t)

	// mock enclave
	enclaveUuid := enclave.EnclaveUUID(testEnclaveUuid)
	enclaveStatus := enclave.EnclaveStatus_Running
	enclaveCreationTime := mockTime.Now() // time doesn't matter
	enclaveMap := map[enclave.EnclaveUUID]*enclave.Enclave{
		enclaveUuid: enclave.NewEnclave(enclaveUuid, testEnclaveUuid, enclaveStatus, &enclaveCreationTime, false),
	}

	mockKurtosisBackend.
		EXPECT().
		GetEnclaves(ctx, &enclave.EnclaveFilters{UUIDs: nil, Statuses: nil}).
		Return(enclaveMap, nil)

	serviceUuid := service.ServiceUUID(testUserService1Uuid)
	serviceRegistration := service.NewServiceRegistration(testUserService1Name, serviceUuid, enclaveUuid, net.IP{}, "")
	serviceContainer := container.NewContainer(container.ContainerStatus_Running, "", []string{}, []string{}, map[string]string{})
	servicesMap := map[service.ServiceUUID]*service.Service{
		serviceUuid: service.NewService(serviceRegistration, map[string]*port_spec.PortSpec{}, net.IP{}, map[string]*port_spec.PortSpec{}, serviceContainer),
	}
	mockKurtosisBackend.
		EXPECT().
		GetUserServices(ctx, enclaveUuid, &service.ServiceFilters{Names: nil, UUIDs: nil, Statuses: nil}).
		Return(servicesMap, nil)

	expectedServiceUuidFilePath := getFilepathStr(2022, 52, testEnclaveUuid, testUserService1Uuid)
	expectedServiceNameFilePath := getFilepathStr(2022, 52, testEnclaveUuid, testUserService1Name)
	expectedServiceShortUuidFilePath := getFilepathStr(2022, 52, testEnclaveUuid, uuid_generator.ShortenedUUIDString(testUserService1Uuid))

	logFileManager := NewLogFileManager(mockKurtosisBackend, mockFs, mockTime, 5)
	err := logFileManager.CreateLogFiles(ctx)
	require.NoError(t, err)

	_, err = mockFs.Stat(expectedServiceUuidFilePath)
	require.NoError(t, err)

	_, err = mockFs.Stat(expectedServiceNameFilePath)
	require.NoError(t, err)

	_, err = mockFs.Stat(expectedServiceShortUuidFilePath)
	require.NoError(t, err)
}
