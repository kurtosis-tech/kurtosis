package log_file_creator

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/logs_clock"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
)

// LogFileCreator is responsible for creating file paths for services across all enclaves.
type LogFileCreator struct {
	time logs_clock.LogsClock

	kurtosisBackend backend_interface.KurtosisBackend

	filesystem volume_filesystem.VolumeFilesystem
}

func NewLogFileCreator(
	time logs_clock.LogsClock,
	kurtosisBackend backend_interface.KurtosisBackend,
	filesystem volume_filesystem.VolumeFilesystem) *LogFileCreator {
	return &LogFileCreator{
		time:            time,
		kurtosisBackend: kurtosisBackend,
		filesystem:      filesystem,
	}
}

// CreateLogFilePathsIdempotently creates three log files for every service across all running enclaves.
// The first is a file with the name ending in the uuid of the service.
// The other two file paths are symlinks to the uuid file, but the file names end with the shortened uuid and service name.
// The file paths are created idempotently meaning if they already exist, new ones are not created.
// If files exist for the shortened uuid and service name files, and they are not symlinks, they are removed and symlink files
// are created.
func (creator *LogFileCreator) CreateLogFilePathsIdempotently() {
	//- engine would query all the enclaves
	//- engine would query all the services in that enclave
	//- collect all the enclave uuids and service names/uuids/shortened uuids
	//
	//- determine the current year and time
	//
	//- engine would create a file for the uuid
	//- idempotently
	//- engine would create a symlink to symlink the shortened uuid to the uuid filepath
	//- if an actual file already exists, remove it and create symlink
	//- engine would create a symlink to symlink the name to the uuid filepath
	//- if an actual file already exists, remove it and create symlink
}
