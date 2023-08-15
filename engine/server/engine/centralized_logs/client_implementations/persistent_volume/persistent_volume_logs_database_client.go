package persistent_volume

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	// Location of logs on the filesystem of the engine
	logsFilepath = "var/log/kurtosis/logs.json"
)

// persistentVolumeLogsDatabaseClient pulls logs from a Docker volume the engine is mounted to
type persistentVolumeLogsDatabaseClient struct {
	kurtosisBackend backend_interface.KurtosisBackend
}

func NewPersistentVolumeLogsDatabaseClient(kurtosisBackend backend_interface.KurtosisBackend) *persistentVolumeLogsDatabaseClient {
	return &persistentVolumeLogsDatabaseClient{
		kurtosisBackend: kurtosisBackend,
	}
}

func (client *persistentVolumeLogsDatabaseClient) StreamUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclaves.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
	shouldFollowLogs bool,
) (
	chan map[service.ServiceUUID][]logline.LogLine,
	chan error,
	context.CancelFunc,
	error,
) {
	ctx, cancelCtxFunc := context.WithCancel(ctx)

	// create user service filers

	// create log filters

	// grab logs
	// return error if smth happens

	// create err chan
	// return err if anything happened

	// create go routine to stream logs for each requested service

	// create go routine to handle stream cancellation
	// wait for all senders to end
	// close all resources
	// cancel context

	// return everything
	return nil, nil, cancelCtxFunc, nil
}

func (client *persistentVolumeLogsDatabaseClient) FilterExistingServiceUuids(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
) (map[service.ServiceUUID]bool, error) {
	userServiceFilters := &service.ServiceFilters{
		Names:    nil,
		UUIDs:    userServiceUuids,
		Statuses: nil,
	}

	existingServicesByUuids, err := client.kurtosisBackend.GetUserServices(ctx, enclaveUuid, userServiceFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services for enclave with UUID '%v' and using filters '%+v'", enclaveUuid, userServiceFilters)
	}

	filteredServiceUuidsSet := map[service.ServiceUUID]bool{}
	for serviceUuid := range userServiceUuids {
		if _, found := existingServicesByUuids[serviceUuid]; found {
			filteredServiceUuidsSet[serviceUuid] = true
		}
	}
	return filteredServiceUuidsSet, nil
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func streamServiceLogLines(
	ctx context.Context,
	logsByKurtosisUserServiceUuidChan chan map[service.ServiceUUID][]logline.LogLine,
	streamErrChan chan error,
	// conjunctiveLogLines []LogLinesFilterWithRegex,
) {
	// for
	// return if context was canceled
	// read a log line
	// turn it into a log line object
	// filter the log line based on the conjunctive filter regex
	// send the log line
}
