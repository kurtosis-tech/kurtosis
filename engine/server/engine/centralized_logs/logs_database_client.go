package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
)

type LogsDatabaseClient interface {
	StreamUserServiceLogs(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		userServiceUuids map[service.ServiceUUID]bool,
		conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
		shouldFollowLogs bool,
	) (
		userServiceLogsByServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
		errChan chan error,
		cancelFunc context.CancelFunc,
		err error,
	)

	FilterExistingServiceUuids(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		userServiceUuids map[service.ServiceUUID]bool,
	) (
		map[service.ServiceUUID]bool,
		error,
	)
}
