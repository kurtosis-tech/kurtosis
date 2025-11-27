package centralized_logs

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/centralized_logs/logline"
)

type LogsDatabaseClient interface {
	StreamUserServiceLogs(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		userServiceUuids map[service.ServiceUUID]bool,
		conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
		shouldFollowLogs bool,
		shouldReturnAllLogs bool, // if true, stream all log lines
		numLogLines uint32, // if [shouldReturnAllLogs] is false, stream that only the last [numLogLines]
	) (
		chan map[service.ServiceUUID][]logline.LogLine,
		chan error,
		context.CancelFunc,
		error,
	)

	FilterExistingServiceUuids(
		ctx context.Context,
		enclaveUuid enclave.EnclaveUUID,
		userServiceUuids map[service.ServiceUUID]bool,
	) (
		map[service.ServiceUUID]bool,
		error,
	)

	StartLogFileManagement(ctx context.Context)

	RemoveEnclaveLogs(enclaveUuid string) error

	RemoveAllLogs() error
}
