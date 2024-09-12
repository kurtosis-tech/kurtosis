package persistent_volume

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/stream_logs_strategy"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/volume_filesystem"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	oneSenderAdded = 1
)

// persistentVolumeLogsDatabaseClient pulls logs from a Docker volume the engine is mounted to
type persistentVolumeLogsDatabaseClient struct {
	kurtosisBackend backend_interface.KurtosisBackend

	filesystem volume_filesystem.VolumeFilesystem

	logFileManager *log_file_manager.LogFileManager

	streamStrategy stream_logs_strategy.StreamLogsStrategy
}

func NewPersistentVolumeLogsDatabaseClient(
	kurtosisBackend backend_interface.KurtosisBackend,
	filesystem volume_filesystem.VolumeFilesystem,
	logFileManager *log_file_manager.LogFileManager,
	streamStrategy stream_logs_strategy.StreamLogsStrategy,
) *persistentVolumeLogsDatabaseClient {
	return &persistentVolumeLogsDatabaseClient{
		kurtosisBackend: kurtosisBackend,
		filesystem:      filesystem,
		logFileManager:  logFileManager,
		streamStrategy:  streamStrategy,
	}
}

func (client *persistentVolumeLogsDatabaseClient) StreamUserServiceLogs(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	userServiceUuids map[service.ServiceUUID]bool,
	conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters,
	shouldFollowLogs bool,
	shouldReturnAllLogs bool,
	numLogLines uint32,
) (
	chan map[service.ServiceUUID][]logline.LogLine,
	chan error,
	context.CancelFunc,
	error,
) {
	ctx, cancelCtxFunc := context.WithCancel(ctx)

	conjunctiveLogFiltersWithRegex, err := logline.NewConjunctiveLogFiltersWithRegex(conjunctiveLogLineFilters)
	if err != nil {
		cancelCtxFunc()
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating conjunctive log line filter with regex from filters '%+v'", conjunctiveLogLineFilters)
	}

	// this channel return an error if the stream fails at some point
	streamErrChan := make(chan error)

	// this channel will return the user service log lines by service UUID
	logLineSender := logline.NewLogLineSender()
	logsByKurtosisUserServiceUuidChan := logLineSender.GetLogsChannel()

	wgSenders := &sync.WaitGroup{}
	for serviceUuid := range userServiceUuids {
		wgSenders.Add(oneSenderAdded)
		go client.streamServiceLogLines(
			ctx,
			wgSenders,
			logLineSender,
			streamErrChan,
			enclaveUuid,
			serviceUuid,
			conjunctiveLogFiltersWithRegex,
			shouldFollowLogs,
			shouldReturnAllLogs,
			numLogLines,
		)
	}

	// this go routine handles the stream cancellation
	go func() {
		//wait for stream go routine to end
		wgSenders.Wait()

		// send all buffered log lines
		logLineSender.Flush()

		// wait until the channel has been fully read/empty before closing it
		closeChannelWhenEmpty(logsByKurtosisUserServiceUuidChan)
		close(streamErrChan)

		//then cancel the context
		cancelCtxFunc()
	}()

	return logsByKurtosisUserServiceUuidChan, streamErrChan, cancelCtxFunc, nil
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

func (client *persistentVolumeLogsDatabaseClient) StartLogFileManagement(ctx context.Context) {
	client.logFileManager.StartLogFileManagement(ctx)
}

func (client *persistentVolumeLogsDatabaseClient) RemoveEnclaveLogs(enclaveUuid string) error {
	return client.logFileManager.RemoveEnclaveLogs(enclaveUuid)
}

func (client *persistentVolumeLogsDatabaseClient) RemoveAllLogs() error {
	return client.logFileManager.RemoveAllLogs()
}

// ====================================================================================================
//
//	Private helper functions
//
// ====================================================================================================
func (client *persistentVolumeLogsDatabaseClient) streamServiceLogLines(
	ctx context.Context,
	wgSenders *sync.WaitGroup,
	logLineSender *logline.LogLineSender,
	streamErrChan chan error,
	enclaveUuid enclave.EnclaveUUID,
	serviceUuid service.ServiceUUID,
	conjunctiveLogLinesFiltersWithRegex []logline.LogLineFilterWithRegex,
	shouldFollowLogs bool,
	shouldReturnAllLogs bool,
	numLogLines uint32,
) {
	defer wgSenders.Done()
	client.streamStrategy.StreamLogs(
		ctx,
		client.filesystem,
		logLineSender,
		streamErrChan,
		enclaveUuid,
		serviceUuid,
		conjunctiveLogLinesFiltersWithRegex,
		shouldFollowLogs,
		shouldReturnAllLogs,
		numLogLines)
}

func closeChannelWhenEmpty(logsChan chan map[service.ServiceUUID][]logline.LogLine) {
	for {
		if len(logsChan) == 0 {
			close(logsChan)
			return
		}
	}
}
