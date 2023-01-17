package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type ConjunctiveLogLineFilters interface {
	GetConjunctiveLogLineFiltersString() string
}

type LogsDatabaseClient interface {
	GetUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveUUID,
		userServiceGuids map[service.ServiceGUID]bool,
		conjunctiveLogLineFilters ConjunctiveLogLineFilters,
	) (
		userServiceLogsByServiceGuidChan chan map[service.ServiceGUID][]LogLine,
		errChan chan error,
		cancelStreamFunc func(),
		err error,
	)
	StreamUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveUUID,
		userServiceGuids map[service.ServiceGUID]bool,
		conjunctiveLogLineFilters ConjunctiveLogLineFilters,
	) (
		userServiceLogsByServiceGuidChan chan map[service.ServiceGUID][]LogLine,
		errChan chan error,
		cancelStreamFunc func(),
		err error,
	)
	FilterExistingServiceGuids(
		ctx context.Context,
		enclaveId enclave.EnclaveUUID,
		userServiceGuids map[service.ServiceGUID]bool,
	) (
		map[service.ServiceGUID]bool,
		error,
	)
}
