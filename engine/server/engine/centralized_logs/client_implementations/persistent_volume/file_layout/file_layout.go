package file_layout

import (
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"time"
)

type LogFileLayout interface {
	// GetLogFileLayoutFormat returns a string representation the "format" that files are laid out in
	// Formats are composed:
	// - "/" - representing a nested directory
	// - "{{ enclaveUuid }}" - representing where an enclave uuid is inserted
	// - "{{ serviceUuid }}" - representing where a service uuid is inserted
	// - time formats specified by strftime https://cplusplus.com/reference/ctime/strftime/
	// - any other ascii text
	GetLogFileLayoutFormat() string

	// GetLogFilePath gets the log file path for [serviceUuid] in [enclaveUuid] at [time]
	GetLogFilePath(time time.Time, enclaveUuid, serviceUuid string) string

	// GetLogFilePaths retrieves a list of filepaths [filesystem] for [serviceUuid] in [enclaveUuid]
	// If [retentionPeriodIntervals] is set to -1, retrieves all filepaths from the currentTime till [retentionPeriod] in order
	// If [retentionPeriodIntervals] is positive, retrieves all filepaths within the range [currentTime - retentionPeriod] and [currentTime - (retentionPeriodIntervals) * retentionPeriod]
	// Returned filepaths sorted from oldest to most recent
	GetLogFilePaths(filesystem volume_filesystem.VolumeFilesystem, retentionPeriod time.Duration, retentionPeriodIntervals int, enclaveUuid, serviceUuid string) ([]string, error)
}
