package log_file_creator

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_consts"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
)

// LogFileCreator is responsible for creating the necessary file paths for service logs across all enclaves.
// Context:
// The LogsAggregator is configured to write logs to three different log file paths, one for uuid, service name, and shortened uuid.
// This is so that the logs are retrievable by each identifier even when enclaves are stopped.
// (More context on this here: https://github.com/kurtosis-tech/kurtosis/pull/1213)
// To prevent storing duplicate logs, the LogFileCreator will ensure that the service name and short uuid log files are just
// symlinks to the uuid log file path.
type LogFileCreator struct {
	kurtosisBackend backend_interface.KurtosisBackend

	filesystem volume_filesystem.VolumeFilesystem

	time logs_clock.LogsClock
}

func NewLogFileCreator(
	kurtosisBackend backend_interface.KurtosisBackend,
	filesystem volume_filesystem.VolumeFilesystem,
	time logs_clock.LogsClock) *LogFileCreator {
	return &LogFileCreator{
		kurtosisBackend: kurtosisBackend,
		filesystem:      filesystem,
		time:            time,
	}
}

// CreateLogFiles creates three log files for every service across all running enclaves.
// The first is a file with the name ending in the uuid of the service.
// The other two file paths are symlinks to the uuid file, ending with the shortened uuid and service name respectively.
// The file paths are created idempotently meaning if they already exist, new ones are not created.
// If files exist for the shortened uuid and service name files, but they are not symlinks, they are removed and symlink files
// are created to prevent duplicate log storage.
func (creator *LogFileCreator) CreateLogFiles(ctx context.Context) error {
	var err error

	// get year and time for the file paths
	year, week := creator.time.Now().ISOWeek()

	// collect enclave and service info for the file paths
	enclaveToServicesMap, err := creator.getEnclaveAndServiceInfo(ctx)
	if err != nil {
		// already wrapped with propagate
		return err
	}

	for enclaveUuid, serviceRegistrations := range enclaveToServicesMap {
		for _, serviceRegistration := range serviceRegistrations {
			serviceUuidStr := string(serviceRegistration.GetUUID())
			serviceNameStr := string(serviceRegistration.GetName())
			serviceShortUuidStr := uuid_generator.ShortenedUUIDString(serviceUuidStr)

			// create log file for the uuid idempotently
			serviceUuidFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceUuidStr)
			_, err = creator.filesystem.Open(serviceUuidFilePathStr)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred creating log file path '%v'.", serviceUuidFilePathStr)
			}

			logrus.Info("CREATING SYMLINKED FILES")
			// create symlinks for name and shortened uuid log files
			serviceNameFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceNameStr)
			err = creator.createSymlinkLogFile(serviceUuidFilePathStr, serviceNameFilePathStr)
			if err != nil {
				return err
			}

			serviceShortUuidFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceShortUuidStr)
			err = creator.createSymlinkLogFile(serviceUuidFilePathStr, serviceShortUuidFilePathStr)
			if err != nil {
				return err
			}
			logrus.Info("SUCCESSFULLY CREATED SYMLINKED FILES")
		}
	}

	return nil
}

func (creator *LogFileCreator) getEnclaveAndServiceInfo(ctx context.Context) (map[enclave.EnclaveUUID][]*service.ServiceRegistration, error) {
	enclaveToServicesMap := map[enclave.EnclaveUUID][]*service.ServiceRegistration{}

	// collect info of all services in enclaves
	enclaves, err := creator.kurtosisBackend.GetEnclaves(ctx, &enclave.EnclaveFilters{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get all enclaves from kurtosis backend.")
	}
	for enclaveUuid, _ := range enclaves {
		var serviceRegistrations []*service.ServiceRegistration

		enclaveServices, err := creator.kurtosisBackend.GetUserServices(ctx, enclaveUuid, &service.ServiceFilters{})
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred while trying to get user services for enclave '%v' from kurtosis backend.", enclaveUuid)
		}
		for _, serviceInfo := range enclaveServices {
			serviceRegistrations = append(serviceRegistrations, serviceInfo.GetRegistration())
		}

		enclaveToServicesMap[enclaveUuid] = serviceRegistrations
	}
	return enclaveToServicesMap, nil
}

func (creator *LogFileCreator) createSymlinkLogFile(targetLogFilePath, symlinkLogFilePath string) error {
	// remove existing log files that could be storing logs at this path
	if err := creator.filesystem.RemoveAll(symlinkLogFilePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred attempting to remove an existing log file at the symlink file path '%v'.", symlinkLogFilePath)
	}
	// replace with symlink
	if err := creator.filesystem.Symlink(targetLogFilePath, symlinkLogFilePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a symlink file path '%v' for target file path '%v'.", targetLogFilePath, targetLogFilePath)
	}
	return nil
}

// creates a filepath of format /<filepath_base>/year/week/<enclave>/serviceIdentifier.<filetype>
func getFilepathStr(year, week int, enclaveUuid, serviceIdentifier string) string {
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), strconv.Itoa(week), enclaveUuid, serviceIdentifier, volume_consts.Filetype)
}
