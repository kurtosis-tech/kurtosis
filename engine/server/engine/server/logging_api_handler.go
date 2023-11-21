package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"golang.org/x/net/websocket"

	api "github.com/kurtosis-tech/kurtosis/api/golang/logging/kurtosis_logging_api_bindings"
)

type LoggingRuntime struct {
	// The version tag of the engine server image, so it can report its own version
	ImageVersionTag string

	EnclaveManager *enclave_manager.EnclaveManager

	// The protected user ID for metrics analytics purpose
	MetricsUserID string

	// User consent to send metrics
	DidUserAcceptSendingMetrics bool

	// The clients for consuming container logs from the logs' database server

	// per week pulls logs from enclaves created post log retention feature
	PerWeekLogsDatabaseClient centralized_logs.LogsDatabaseClient

	// per file pulls logs from enclaves created pre log retention feature
	// TODO: remove once users are fully migrated to log retention/new log schema
	PerFileLogsDatabaseClient centralized_logs.LogsDatabaseClient

	LogFileManager *log_file_manager.LogFileManager

	MetricsClient metrics_client.MetricsClient
}

func sendErrorCode(ctx echo.Context, code int, message string) error {
	err := ctx.JSON(code, message)
	return err
}

type LogStreamer struct {
	serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
	errChan                      chan error
	cancelCtxFunc                func()
	requestedServiceUuids        []user_service.ServiceUUID
	notFoundServiceUuids         []string
}

func (engine LoggingRuntime) GetEnclavesEnclaveIdentifierLogs(ctx echo.Context, enclaveIdentifier api.EnclaveIdentifier, params api.GetEnclavesEnclaveIdentifierLogsParams) error {
	streamer, err := engine.getLogStreamer(
		ctx,
		enclaveIdentifier,
		utils.MapList(params.ServiceUuidSet, func(x string) user_service.ServiceUUID { return user_service.ServiceUUID(x) }),
		params.FollowLogs,
		params.ReturnAllLogs,
		utils.MapPointer(params.NumLogLines, func(x int) uint32 { return uint32(x) }),
		params.ConjunctiveFilters,
	)
	if err != nil {
		logrus.Error(err)
		return sendErrorCode(ctx, http.StatusInternalServerError, "Failed to setup log streaming")
	}

	if ctx.IsWebSocket() {
		return streamer.streamWithWebsocket(ctx)
	} else {
		return streamer.streamHTTP(ctx)
	}

}

func (engine LoggingRuntime) GetEnclavesEnclaveIdentifierServicesServiceUuidLogs(ctx echo.Context, enclaveIdentifier api.EnclaveIdentifier, serviceUuid api.ServiceUuid, params api.GetEnclavesEnclaveIdentifierServicesServiceUuidLogsParams) error {
	serviceUuidStrSet := []user_service.ServiceUUID{user_service.ServiceUUID(serviceUuid)}
	streamer, err := engine.getLogStreamer(
		ctx,
		enclaveIdentifier,
		serviceUuidStrSet,
		params.FollowLogs,
		params.ReturnAllLogs,
		utils.MapPointer(params.NumLogLines, func(x int) uint32 { return uint32(x) }),
		params.ConjunctiveFilters,
	)
	if err != nil {
		logrus.Error(err)
		return sendErrorCode(ctx, http.StatusInternalServerError, "Failed to setup log streaming")
	}

	if ctx.IsWebSocket() {
		return streamer.streamWithWebsocket(ctx)
	} else {
		return streamer.streamHTTP(ctx)
	}
}

func (engine LoggingRuntime) getLogStreamer(
	ctx echo.Context,
	enclaveIdentifier api.EnclaveIdentifier,
	serviceUuidList []user_service.ServiceUUID,
	maybeShouldFollowLogs *bool,
	maybeShouldReturnAllLogs *bool,
	maybeNumLogLines *uint32,
	maybeFilters *[]api.LogLineFilter,
) (*LogStreamer, error) {
	enclaveUuid, err := engine.EnclaveManager.GetEnclaveUuidForEnclaveIdentifier(context.Background(), enclaveIdentifier)
	if err != nil {
		logrus.Errorf("An error occurred while fetching uuid for enclave '%v'. This could happen if the enclave has been deleted. Treating it as UUID", enclaveIdentifier)
		return nil, err
	}

	requestedServiceUuids := make(map[user_service.ServiceUUID]bool, len(serviceUuidList))
	shouldFollowLogs := utils.DerefWith(maybeShouldFollowLogs, false)
	shouldReturnAllLogs := utils.DerefWith(maybeShouldReturnAllLogs, false)
	numLogLines := utils.DerefWith(maybeNumLogLines, 100)
	filters := utils.DerefWith(maybeFilters, []api.LogLineFilter{})
	context := ctx.Request().Context()

	for _, serviceUuidStr := range serviceUuidList {
		serviceUuid := user_service.ServiceUUID(serviceUuidStr)
		requestedServiceUuids[serviceUuid] = true
	}

	if engine.PerWeekLogsDatabaseClient == nil || engine.PerFileLogsDatabaseClient == nil {
		return nil, stacktrace.NewError("It's not possible to return service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
		errChan                      chan error
		cancelCtxFunc                func()
	)

	notFoundServiceUuids, err := engine.reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(context, enclaveUuid, requestedServiceUuids)
	if err != nil {
		return nil, err
	}

	conjunctiveLogLineFilters, err := fromHttpLogLineFilters(filters)
	if err != nil {
		return nil, err
	}

	// get enclave creation time to determine strategy to pull logs
	enclaveCreationTime, err := engine.getEnclaveCreationTime(context, enclaveUuid)
	if err != nil {
		return nil, err
	}
	logsDatabaseClient := engine.getLogsDatabaseClient(enclaveCreationTime)

	serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, err = logsDatabaseClient.StreamUserServiceLogs(
		context,
		enclaveUuid,
		requestedServiceUuids,
		conjunctiveLogLineFilters,
		shouldFollowLogs,
		shouldReturnAllLogs,
		uint32(numLogLines))
	if err != nil {
		return nil, err
	}

	return &LogStreamer{
		serviceLogsByServiceUuidChan: serviceLogsByServiceUuidChan,
		errChan:                      errChan,
		cancelCtxFunc:                cancelCtxFunc,
		notFoundServiceUuids:         notFoundServiceUuids,
		requestedServiceUuids:        serviceUuidList,
	}, nil
}

func (streamer LogStreamer) close() {
	streamer.cancelCtxFunc()
}

func (streamer LogStreamer) streamWithWebsocket(ctx echo.Context) error {
	defer streamer.close()

	logrus.Debugf("Starting log stream using Websocket on services: %s", streamer.requestedServiceUuids)
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		for {
			select {
			//stream case
			case serviceLogsByServiceUuid, isChanOpen := <-streamer.serviceLogsByServiceUuidChan:
				//If the channel is closed means that the logs database client won't continue sending streams
				if !isChanOpen {
					logrus.Debug("Exiting the stream loop after receiving a close signal from the service logs by service UUID channel")
					return
				}

				getServiceLogsResponse := newLogsResponseHttp(streamer.requestedServiceUuids, serviceLogsByServiceUuid, streamer.notFoundServiceUuids)
				err := websocket.JSON.Send(ws, getServiceLogsResponse)
				if err != nil {
					ctx.Logger().Error(err)
				}

			//error from logs database case
			case err, isChanOpen := <-streamer.errChan:
				if isChanOpen {
					logrus.Error(err)
					logrus.Debug("Exiting the stream because an error from the logs database client was received through the error chan.")
					return
				}
				logrus.Debug("Exiting the stream loop after receiving a close signal from the error chan")
				return
			}
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())
	return nil
}

func (streamer LogStreamer) streamHTTP(ctx echo.Context) error {
	defer streamer.close()

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	ctx.Response().WriteHeader(http.StatusOK)
	enc := json.NewEncoder(ctx.Response())

	logrus.Debugf("Starting log stream using HTTP on services: %s", streamer.requestedServiceUuids)
	for {
		select {
		//stream case
		case serviceLogsByServiceUuid, isChanOpen := <-streamer.serviceLogsByServiceUuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				logrus.Debug("Exiting the stream loop after receiving a close signal from the service logs by service UUID channel")
				return nil
			}

			getServiceLogsResponse := newLogsResponseHttp(streamer.requestedServiceUuids, serviceLogsByServiceUuid, streamer.notFoundServiceUuids)
			if err := enc.Encode(getServiceLogsResponse); err != nil {
				return err
			}
			ctx.Response().Flush()

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

func (service *LoggingRuntime) reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	requestedServiceUuids map[user_service.ServiceUUID]bool,
) ([]string, error) {
	// doesn't matter which logs client is used here
	existingServiceUuids, err := service.PerWeekLogsDatabaseClient.FilterExistingServiceUuids(ctx, enclaveUuid, requestedServiceUuids)
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

// If the enclave was created prior to log retention, return the per file logs client
func (service *LoggingRuntime) getLogsDatabaseClient(enclaveCreationTime time.Time) centralized_logs.LogsDatabaseClient {
	if enclaveCreationTime.After(logRetentionFeatureReleaseTime) {
		return service.PerWeekLogsDatabaseClient
	} else {
		return service.PerFileLogsDatabaseClient
	}
}

func (service *LoggingRuntime) getEnclaveCreationTime(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (time.Time, error) {
	enclaves, err := service.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return time.Time{}, err
	}

	enclaveObj, found := enclaves[string(enclaveUuid)]
	if !found {
		return time.Time{}, stacktrace.NewError("Engine could not find enclave '%v'", enclaveUuid)
	}

	timestamp := enclaveObj.CreationTime
	return timestamp, nil
}

func fromHttpLogLineFilters(
	logLineFilters []api.LogLineFilter,
) (logline.ConjunctiveLogLineFilters, error) {
	var conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters

	for _, logLineFilter := range logLineFilters {
		var filter *logline.LogLineFilter
		operator := logLineFilter.Operator
		filterTextPattern := logLineFilter.TextPattern
		switch operator {
		case api.DOESCONTAINTEXT:
			filter = logline.NewDoesContainTextLogLineFilter(filterTextPattern)
		case api.DOESNOTCONTAINTEXT:
			filter = logline.NewDoesNotContainTextLogLineFilter(filterTextPattern)
		case api.DOESCONTAINMATCHREGEX:
			filter = logline.NewDoesContainMatchRegexLogLineFilter(filterTextPattern)
		case api.DOESNOTCONTAINMATCHREGEX:
			filter = logline.NewDoesNotContainMatchRegexLogLineFilter(filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in GRPC filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		conjunctiveLogLineFilters = append(conjunctiveLogLineFilters, *filter)
	}

	return conjunctiveLogLineFilters, nil
}

func (service *EngineRuntime) reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	requestedServiceUuids map[user_service.ServiceUUID]bool,
) ([]string, error) {
	// doesn't matter which logs client is used here
	existingServiceUuids, err := service.PerWeekLogsDatabaseClient.FilterExistingServiceUuids(ctx, enclaveUuid, requestedServiceUuids)
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

func newLogsResponseHttp(
	requestedServiceUuids []user_service.ServiceUUID,
	serviceLogsByServiceUuid map[user_service.ServiceUUID][]logline.LogLine,
	initialNotFoundServiceUuids []string,
) *api.ServiceLogs {
	serviceLogLinesByUuid := make(map[string]api.LogLine, len(serviceLogsByServiceUuid))
	notFoundServiceUuids := make([]string, len(initialNotFoundServiceUuids))
	for _, serviceUuid := range requestedServiceUuids {
		serviceUuidStr := string(serviceUuid)
		isInNotFoundUuidList := slices.Contains(initialNotFoundServiceUuids, serviceUuidStr)
		serviceLogLines, found := serviceLogsByServiceUuid[serviceUuid]
		// should continue in the not-found-UUID list
		if !found && isInNotFoundUuidList {
			notFoundServiceUuids = append(notFoundServiceUuids, serviceUuidStr)
		}

		// there is no new log lines but is a found UUID, so it has to be included in the service logs map
		if !found && !isInNotFoundUuidList {
			serviceLogLinesByUuid[serviceUuidStr] = api.LogLine{
				Line:      []string{},
				Timestamp: time.Now(),
			}
		}

		logLines := newHttpBindingsLogLineFromLogLines(serviceLogLines)
		serviceLogLinesByUuid[serviceUuidStr] = logLines
	}

	response := &api.ServiceLogs{
		NotFoundServiceUuidSet:   &notFoundServiceUuids,
		ServiceLogsByServiceUuid: &serviceLogLinesByUuid,
	}
	return response
}

func newHttpBindingsLogLineFromLogLines(logLines []logline.LogLine) api.LogLine {
	logLinesStr := make([]string, len(logLines))
	var logTimestamp time.Time

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
		logTimestamp = logLine.GetTimestamp()
	}

	return api.LogLine{Line: logLinesStr, Timestamp: logTimestamp}

}
