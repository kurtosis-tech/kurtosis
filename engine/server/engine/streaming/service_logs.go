package streaming

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_http"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_logline"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"

	api_type "github.com/kurtosis-tech/kurtosis/api/golang/v2/http_rest/api_types"
)

var (
	defaultNumberOfLogLines uint32 = 100
)

type ServiceLogStreamer struct {
	serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
	errChan                      chan error
	cancelCtxFunc                func()
	requestedServiceUuids        []user_service.ServiceUUID
	notFoundServiceUuids         []string
}

func NewServiceLogStreamer(
	ctx context.Context,
	enclaveManager *enclave_manager.EnclaveManager,
	enclaveIdentifier api_type.EnclaveIdentifier,
	logsDatabaseClient centralized_logs.LogsDatabaseClient,
	serviceUuidList []user_service.ServiceUUID,
	maybeShouldFollowLogs *bool,
	maybeShouldReturnAllLogs *bool,
	maybeNumLogLines *uint32,
	maybeFilters *[]api_type.LogLineFilter,
) (*ServiceLogStreamer, error) {
	enclaveUuid, err := enclaveManager.GetEnclaveUuidForEnclaveIdentifier(ctx, enclaveIdentifier)
	if err != nil {
		logrus.Errorf("An error occurred while fetching uuid for enclave '%v'. This could happen if the enclave has been deleted. Treating it as UUID", enclaveIdentifier)
		return nil, err
	}

	requestedServiceUuids := make(map[user_service.ServiceUUID]bool, len(serviceUuidList))
	shouldFollowLogs := utils.DerefWith(maybeShouldFollowLogs, false)
	shouldReturnAllLogs := utils.DerefWith(maybeShouldReturnAllLogs, false)
	numLogLines := utils.DerefWith(maybeNumLogLines, defaultNumberOfLogLines)
	filters := utils.DerefWith(maybeFilters, []api_type.LogLineFilter{})

	for _, serviceUuidStr := range serviceUuidList {
		serviceUuid := user_service.ServiceUUID(serviceUuidStr)
		requestedServiceUuids[serviceUuid] = true
	}

	if logsDatabaseClient == nil {
		return nil, stacktrace.NewError("It's not possible to return service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
		errChan                      chan error
		cancelCtxFunc                func()
	)

	notFoundServiceUuids, err := reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(ctx, enclaveUuid, logsDatabaseClient, requestedServiceUuids)
	if err != nil {
		return nil, err
	}

	conjunctiveLogLineFilters, err := to_logline.ToLoglineLogLineFilters(filters)
	if err != nil {
		return nil, err
	}

	serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, err = logsDatabaseClient.StreamUserServiceLogs(
		ctx,
		enclaveUuid,
		requestedServiceUuids,
		conjunctiveLogLineFilters,
		shouldFollowLogs,
		shouldReturnAllLogs,
		uint32(numLogLines))
	if err != nil {
		return nil, err
	}

	return &ServiceLogStreamer{
		serviceLogsByServiceUuidChan: serviceLogsByServiceUuidChan,
		errChan:                      errChan,
		cancelCtxFunc:                cancelCtxFunc,
		notFoundServiceUuids:         notFoundServiceUuids,
		requestedServiceUuids:        serviceUuidList,
	}, nil
}

func (streamer ServiceLogStreamer) GetRequestedServiceUuids() []user_service.ServiceUUID {
	return streamer.requestedServiceUuids
}

func (streamer ServiceLogStreamer) Close() {
	streamer.cancelCtxFunc()
}

func (streamer ServiceLogStreamer) Consume(consumer func(*api_type.ServiceLogs) error) error {
	for {
		select {
		//stream case
		case serviceLogsByServiceUuid, isChanOpen := <-streamer.serviceLogsByServiceUuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				logrus.Debug("Exiting the stream loop after receiving a close signal from the service logs by service UUID channel")
				return nil
			}

			serviceLogsResponse := to_http.ToHttpServiceLogs(streamer.requestedServiceUuids, serviceLogsByServiceUuid, streamer.notFoundServiceUuids)
			err := consumer(serviceLogsResponse)
			if err != nil {
				return err
			}

		//error from logs database case
		case err, isChanOpen := <-streamer.errChan:
			if isChanOpen {
				logrus.Error(err)
				logrus.Debug("Exiting the stream because an error from the logs database client was received through the error chan.")
				return nil
			}
			logrus.Debug("Exiting the stream loop after receiving a close signal from the error chan")
			return nil
		}
	}
}

// =============================================================================================================================================
// ============================================== Helper Functions =============================================================================
// =============================================================================================================================================

func reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	logsDatabaseClient centralized_logs.LogsDatabaseClient,
	requestedServiceUuids map[user_service.ServiceUUID]bool,
) ([]string, error) {
	// doesn't matter which logs client is used here
	existingServiceUuids, err := logsDatabaseClient.FilterExistingServiceUuids(ctx, enclaveUuid, requestedServiceUuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service UUIDs from the log client for enclave '%v' and for the requested UUIDs '%+v'", enclaveUuid, requestedServiceUuids)
	}

	notFoundServiceUuidsMap := getNotFoundServiceUuidsAndEmptyServiceLogsMap(requestedServiceUuids, existingServiceUuids)
	var notFoundServiceUuids []string
	for service := range notFoundServiceUuidsMap {
		notFoundServiceUuids = append(notFoundServiceUuids, service)
	}
	return notFoundServiceUuids, nil
}

func getNotFoundServiceUuidsAndEmptyServiceLogsMap(
	requestedServiceUuids map[user_service.ServiceUUID]bool,
	existingServiceUuids map[user_service.ServiceUUID]bool,
) map[string]bool {
	notFoundServiceUuids := map[string]bool{}

	for requestedServiceUuid := range requestedServiceUuids {
		if _, found := existingServiceUuids[requestedServiceUuid]; !found {
			requestedServiceUuidStr := string(requestedServiceUuid)
			notFoundServiceUuids[requestedServiceUuidStr] = true
		}
	}

	return notFoundServiceUuids
}
