package centralized_logs

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"io"
)


//TODO This is a temporary hack until we get the real MockedKurtosisBackend created by Mockery, then we should remove this interface
type MinimalKurtosisBackend interface {
	GetUserServiceLogs(
		ctx context.Context,
		enclaveId enclave.EnclaveID,
		filters *service.ServiceFilters,
		shouldFollowLogs bool,
	) (
		successfulUserServiceLogs map[service.ServiceGUID]io.ReadCloser,
		erroredUserServiceGuids map[service.ServiceGUID]error,
		resultError error,
	)
}
