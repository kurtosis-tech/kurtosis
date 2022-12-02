package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type LogPipeLine interface {
	PipeLineStringify() string
}

type LogsDatabaseClient interface {
	GetUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveID,
		userServiceGuids map[service.ServiceGUID]bool,
		logPipeLine LogPipeLine,
	) (
		userServiceLogsByServiceGuidChan chan map[service.ServiceGUID][]LogLine,
		errChan chan error,
		cancelStreamFunc func(),
		err error,
	)
	StreamUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveID,
		userServiceGuids map[service.ServiceGUID]bool,
		logPipeLine LogPipeLine,
	) (
		userServiceLogsByServiceGuidChan chan map[service.ServiceGUID][]LogLine,
		errChan chan error,
		cancelStreamFunc func(),
		err error,
	)
	FilterExistingServiceGuids(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		userServiceGuids map[service.ServiceGUID]bool,
	) (
		map[service.ServiceGUID]bool,
		error,
	)
}
