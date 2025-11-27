package stream_logs_strategy

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/logline"
)

// This interface is for implementing new algorithms for streaming logs from an underlying Volume filesystem
// For ex. if the schema for storing logs files changes, a new StreamLogsStrategy should be implemented
// to pull from logs files based on that schema
type StreamLogsStrategy interface {
	StreamLogs(
		ctx context.Context,
		fs volume_filesystem.VolumeFilesystem,
		logLineSender *logline.LogLineSender,
		streamErrChan chan error,
		enclaveUuid enclave.EnclaveUUID,
		serviceUuid service.ServiceUUID,
		conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
		shouldFollowLogs bool,
		shouldReturnAllLogs bool,
		numLogLines uint32,
	)
}
