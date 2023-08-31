package stream_logs_strategy

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
)

type StreamLogsStrategy interface {
	StreamLogs(
		ctx context.Context,
		fs volume_filesystem.VolumeFilesystem,
		logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
		streamErrChan chan error,
		enclaveUuid enclave.EnclaveUUID,
		serviceUuid service.ServiceUUID,
		conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
		shouldFollowLogs bool,
	)
}
