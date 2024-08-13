package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"time"
)

// this interface should support operations by the following clients:
// - logs aggregator
//   - needs to know how to store files
//
// - log file manager
//   - needs to know what to name the log files it creates
//   - needs to log file paths to remove based on retention period, but also all of them
//
// - stream logs strategy
//   - needs to know what
type LogFileLayout interface {
	// GetLogFileLayoutFormat Returns a string representation the "format" that files are laid out in
	// Formats are composed:
	// - "/" - representing a nested directory
	// - "<enclaveUuid>" - representing where an enclave uuid is inserted
	// - "<serviceUuid>" - representing where a service uuid is inserted
	// - time formats specified by strftime https://cplusplus.com/reference/ctime/strftime/
	// - any other ascii text
	GetLogFileLayoutFormat() string

	// GetLogFilePath??

	// GetLogFilePaths Retrieves a list of filepaths [filesystem] for [serviceUuid] in [enclaveUuid]
	// If [retentionPeriodIntervals] is set to -1, retrieves all filepaths from the currentTime till [retentionPeriod]
	// If [retentionPeriodIntervals] is positive, retrieves all filepaths within the range [currentTime - retentionPeriod] and [currentTime - (retentionPeriodIntervals) * retentionPeriod]
	GetLogFilePaths(filesystem volume_filesystem.VolumeFilesystem, retentionPeriod time.Duration, retentionPeriodIntervals int, enclaveUuid, serviceUuid string) ([]string, error)

	// rename to "GetLogFilePaths"BasedOnRetentoin
	// - should support getting log file paths within time period
	// - theoretically should support getting all file log file paths
	// following Philosophy of Software Design by John Ousterhoust - small interface, deep modules
	//
}
