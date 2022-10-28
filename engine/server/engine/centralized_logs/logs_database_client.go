package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
)

type LogsDatabaseClient interface {
	GetUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveID,
		userServiceGuids map[service.ServiceGUID]bool,
	) (
		userServiceLogsByServiceGuid map[service.ServiceGUID][]string,
		err error,
	)
	StreamUserServiceLogs(
		ctx context.Context,
		enclaveID enclave.EnclaveID,
		userServiceGuids map[service.ServiceGUID]bool,
	) (
		userServiceLogsByServiceGuidChan chan map[service.ServiceGUID][]string,
		errChan chan error,
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
