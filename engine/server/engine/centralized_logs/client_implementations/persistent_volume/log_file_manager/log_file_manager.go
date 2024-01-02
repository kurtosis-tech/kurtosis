package log_file_manager

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
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"strconv"
	"time"
)

const (
	oneWeek = 7 * 24 * time.Hour
)

// LogFileManager is responsible for creating and removing log files from filesystem.
type LogFileManager struct {
	kurtosisBackend backend_interface.KurtosisBackend

	filesystem volume_filesystem.VolumeFilesystem

	time logs_clock.LogsClock
}

func NewLogFileManager(
	kurtosisBackend backend_interface.KurtosisBackend,
	filesystem volume_filesystem.VolumeFilesystem,
	time logs_clock.LogsClock) *LogFileManager {
	return &LogFileManager{
		kurtosisBackend: kurtosisBackend,
		filesystem:      filesystem,
		time:            time,
	}
}

// StartLogFileManagement initiates logic for managing log files in the filesystem
func (manager *LogFileManager) StartLogFileManagement(ctx context.Context) {
	// Schedule thread for removing log files beyond retention period
	go func() {
		logrus.Debugf("Scheduling log removal for log retention every '%v' hours...", volume_consts.RemoveLogsWaitHours)
		manager.RemoveLogsBeyondRetentionPeriod()

		logRemovalTicker := time.NewTicker(volume_consts.RemoveLogsWaitHours)
		for range logRemovalTicker.C {
			logrus.Debug("Attempting to remove old log file paths...")
			manager.RemoveLogsBeyondRetentionPeriod()
		}
	}()

	// Schedule thread for creating log files
	go func() {
		// TODO: Remove this when moving away from persistent volume logs db
		// Creating log file paths on an interval is a hack to prevent duplicate logs from being stored by the log aggregator
		// The LogsAggregator is configured to write logs to three different log file paths, one for uuid, service name, and shortened uuid
		// This is so that the logs are retrievable by each identifier even when enclaves are stopped. More context on this here: https://github.com/kurtosis-tech/kurtosis/pull/1213
		// To prevent storing duplicate logs, the CreateLogFiles will ensure that the service name and short uuid log files are just symlinks to the uuid log file path
		logFileCreatorTicker := time.NewTicker(volume_consts.CreateLogsWaitMinutes)

		logrus.Debugf("Scheduling log file path creation every '%v' minutes...", volume_consts.CreateLogsWaitMinutes)
		for range logFileCreatorTicker.C {
			logrus.Trace("Creating log file paths...")
			err := manager.CreateLogFiles(ctx)
			if err != nil {
				logrus.Errorf("An error occurred attempting to create log file paths: %v", err)
			} else {
				logrus.Trace("Successfully created log file paths.")
			}
		}
	}()
}

// CreateLogFiles creates three log files for every service across all running enclaves.
// The first is a file with the name ending in the uuid of the service.
// The other two file paths are symlinks to the uuid file, ending with the shortened uuid and service name respectively.
// If files exist for the shortened uuid and service name files, but they are not symlinks, they are removed and symlink files
// are created to prevent duplicate log storage.
func (manager *LogFileManager) CreateLogFiles(ctx context.Context) error {
	var err error

	year, week := manager.time.Now().ISOWeek()

	enclaveToServicesMap, err := manager.getEnclaveAndServiceInfo(ctx)
	if err != nil {
		// already wrapped with propagate
		return err
	}

	for enclaveUuid, serviceRegistrations := range enclaveToServicesMap {
		for _, serviceRegistration := range serviceRegistrations {
			serviceUuidStr := string(serviceRegistration.GetUUID())
			serviceNameStr := string(serviceRegistration.GetName())
			serviceShortUuidStr := uuid_generator.ShortenedUUIDString(serviceUuidStr)

			serviceUuidFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceUuidStr)
			if err = manager.createLogFileIdempotently(serviceUuidFilePathStr); err != nil {
				return err
			}

			serviceNameFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceNameStr)
			if err = manager.createSymlinkLogFile(serviceUuidFilePathStr, serviceNameFilePathStr); err != nil {
				return err
			}
			logrus.Tracef("Created symlinked log file: '%v'", serviceNameFilePathStr)

			serviceShortUuidFilePathStr := getFilepathStr(year, week, string(enclaveUuid), serviceShortUuidStr)
			if err = manager.createSymlinkLogFile(serviceUuidFilePathStr, serviceShortUuidFilePathStr); err != nil {
				return err
			}
			logrus.Tracef("Created symlinked log file: '%v'", serviceShortUuidFilePathStr)
		}
	}

	return nil
}

// RemoveLogsBeyondRetentionPeriod implements the Job cron interface. It removes logs a week older than the log retention period.
func (manager *LogFileManager) RemoveLogsBeyondRetentionPeriod() {
	// compute the next oldest week
	year, weekToRemove := manager.time.Now().Add(time.Duration(-volume_consts.LogRetentionPeriodInWeeks) * oneWeek).ISOWeek()

	// remove directory for that week
	oldLogsDirPath := getLogsDirPathForWeek(year, weekToRemove)
	if err := manager.filesystem.RemoveAll(oldLogsDirPath); err != nil {
		logrus.Warnf("An error occurred removing old logs at the following path '%v': %v\n", oldLogsDirPath, err)
	}
	logrus.Debugf("Removed logs beyond retention period at the following path: '%v'", oldLogsDirPath)
}

func (manager *LogFileManager) RemoveAllLogs() error {
	// only removes logs for this year because Docker prevents all logs from base logs storage file path
	year, _ := manager.time.Now().ISOWeek()
	if err := manager.filesystem.RemoveAll(getLogsDirPathForYear(year)); err != nil {
		return stacktrace.Propagate(err, "An error occurred attempting to remove all logs.")
	}
	return nil
}

func (manager *LogFileManager) RemoveEnclaveLogs(enclaveUuid string) error {
	currentTime := manager.time.Now()
	for i := 0; i < volume_consts.LogRetentionPeriodInWeeks; i++ {
		year, week := currentTime.Add(time.Duration(-i) * oneWeek).ISOWeek()
		enclaveLogsDirPathForWeek := getEnclaveLogsDirPath(year, week, enclaveUuid)
		if err := manager.filesystem.RemoveAll(enclaveLogsDirPathForWeek); err != nil {
			return stacktrace.Propagate(err, "An error occurred attempting to remove logs for enclave '%v' logs at the following path: %v", enclaveUuid, enclaveLogsDirPathForWeek)
		}
	}
	return nil
}

func (manager *LogFileManager) getEnclaveAndServiceInfo(ctx context.Context) (map[enclave.EnclaveUUID][]*service.ServiceRegistration, error) {
	enclaveToServicesMap := map[enclave.EnclaveUUID][]*service.ServiceRegistration{}

	enclaves, err := manager.kurtosisBackend.GetEnclaves(ctx, &enclave.EnclaveFilters{UUIDs: nil, Statuses: nil})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get all enclaves from kurtosis backend.")
	}
	for enclaveUuid := range enclaves {
		var serviceRegistrations []*service.ServiceRegistration

		enclaveServices, err := manager.kurtosisBackend.GetUserServices(ctx, enclaveUuid, &service.ServiceFilters{Names: nil, UUIDs: nil, Statuses: nil})
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

func (manager *LogFileManager) createLogFileIdempotently(logFilePath string) error {
	var err error
	if _, err = manager.filesystem.Stat(logFilePath); os.IsNotExist(err) {
		if _, err = manager.filesystem.Create(logFilePath); err != nil {
			return stacktrace.Propagate(err, "An error occurred creating a log file path at '%v'", logFilePath)
		}
		logrus.Tracef("Created log file: '%v'", logFilePath)
		return nil
	}
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred checking if log file path at '%v' existed.", logFilePath)
	}
	return nil
}

func (manager *LogFileManager) createSymlinkLogFile(targetLogFilePath, symlinkLogFilePath string) error {
	// remove existing log files that could be storing logs at this path
	if err := manager.filesystem.RemoveAll(symlinkLogFilePath); err != nil && !errors.IsNotFound(err) {
		return stacktrace.Propagate(err, "An error occurred attempting to remove an existing log file at the symlink file path '%v'.", symlinkLogFilePath)
	}
	// replace with symlink
	if err := manager.filesystem.Symlink(targetLogFilePath, symlinkLogFilePath); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a symlink file path '%v' for target file path '%v'.", targetLogFilePath, targetLogFilePath)
	}
	return nil
}

// TODO: Implement a FilePath Builder to centralize log file path creation across the entire module
// creates a filepath of format /<filepath_base>/year/week/<enclave>/serviceIdentifier.<filetype>
func getFilepathStr(year, week int, enclaveUuid, serviceIdentifier string) string {
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf(volume_consts.PerWeekFilePathFmtStr, volume_consts.LogsStorageDirpath, strconv.Itoa(year), formattedWeekNum, enclaveUuid, serviceIdentifier, volume_consts.Filetype)
}

// creates a directory path of format /<filepath_base>/year/week/<enclave>/
func getEnclaveLogsDirPath(year, week int, enclaveUuid string) string {
	logsDirPathForYearAndWeek := getLogsDirPathForWeek(year, week)
	return fmt.Sprintf("%s%s/", logsDirPathForYearAndWeek, enclaveUuid)
}

// creates a directory path of format /<filepath_base>/year/week/
func getLogsDirPathForWeek(year, week int) string {
	logsDirPathForYear := getLogsDirPathForYear(year)
	formattedWeekNum := fmt.Sprintf("%02d", week)
	return fmt.Sprintf("%s%s/", logsDirPathForYear, formattedWeekNum)
}

// creates a directory path of format /<filepath_base>/year/
func getLogsDirPathForYear(year int) string {
	return fmt.Sprintf("%s%s/", volume_consts.LogsStorageDirpath, strconv.Itoa(year))
}
