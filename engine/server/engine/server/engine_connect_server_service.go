package server

import (
	"connectrpc.com/connect"
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

var (
	logRetentionFeatureReleaseTime = time.Date(2023, 9, 7, 13, 0, 0, 0, time.UTC)
)

type EngineConnectServerService struct {
	// The version tag of the engine server image, so it can report its own version
	imageVersionTag string

	enclaveManager *enclave_manager.EnclaveManager

	// The protected user ID for metrics analytics purpose
	metricsUserID string

	// User consent to send metrics
	didUserAcceptSendingMetrics bool

	// The clients for consuming container logs from the logs' database server

	// per week pulls logs from enclaves created post log retention feature
	perWeekLogsDatabaseClient centralized_logs.LogsDatabaseClient

	// per file pulls logs from enclaves created pre log retention feature
	// TODO: remove once users are fully migrated to log retention/new log schema
	perFileLogsDatabaseClient centralized_logs.LogsDatabaseClient

	logFileManager *log_file_manager.LogFileManager
}

func NewEngineConnectServerService(
	imageVersionTag string,
	enclaveManager *enclave_manager.EnclaveManager,
	metricsUserId string,
	didUserAcceptSendingMetrics bool,
	perWeekLogsDatabaseClient centralized_logs.LogsDatabaseClient,
	perFileLogsDatabaseClient centralized_logs.LogsDatabaseClient,
	logFileManager *log_file_manager.LogFileManager,
) *EngineConnectServerService {
	service := &EngineConnectServerService{
		imageVersionTag:             imageVersionTag,
		enclaveManager:              enclaveManager,
		metricsUserID:               metricsUserId,
		didUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		perWeekLogsDatabaseClient:   perWeekLogsDatabaseClient,
		perFileLogsDatabaseClient:   perFileLogsDatabaseClient,
		logFileManager:              logFileManager,
	}
	return service
}

func (service *EngineConnectServerService) GetEngineInfo(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse], error) {
	result := &kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse{
		EngineVersion: service.imageVersionTag,
	}
	return connect.NewResponse(result), nil
}

func (service *EngineConnectServerService) CreateEnclave(ctx context.Context, connectArgs *connect.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	args := connectArgs.Msg
	logrus.Debugf("args: %+v", args)
	apiContainerLogLevel, err := logrus.ParseLevel(args.GetApiContainerLogLevel())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.ApiContainerLogLevel)
	}

	isProduction := false
	if args.GetMode() == kurtosis_engine_rpc_api_bindings.EnclaveMode_PRODUCTION {
		isProduction = true
	}
	enclaveInfo, err := service.enclaveManager.CreateEnclave(
		ctx,
		service.imageVersionTag,
		args.GetApiContainerVersionTag(),
		apiContainerLogLevel,
		args.GetEnclaveName(),
		isProduction,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with name '%v'", args.EnclaveName)
	}

	response := &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
		EnclaveInfo: enclaveInfo,
	}

	return connect.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetEnclaves(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	infoForEnclaves, err := service.enclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	response := &kurtosis_engine_rpc_api_bindings.GetEnclavesResponse{EnclaveInfo: infoForEnclaves}
	return connect.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetExistingAndHistoricalEnclaveIdentifiers(context.Context, *connect.Request[emptypb.Empty]) (*connect.Response[kurtosis_engine_rpc_api_bindings.GetExistingAndHistoricalEnclaveIdentifiersResponse], error) {
	allIdentifiers, err := service.enclaveManager.GetExistingAndHistoricalEnclaveIdentifiers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching enclave identifiers")
	}
	response := &kurtosis_engine_rpc_api_bindings.GetExistingAndHistoricalEnclaveIdentifiersResponse{AllIdentifiers: allIdentifiers}
	return connect.NewResponse(response), nil
}

func (service *EngineConnectServerService) StopEnclave(ctx context.Context, connectArgs *connect.Request[kurtosis_engine_rpc_api_bindings.StopEnclaveArgs]) (*connect.Response[emptypb.Empty], error) {
	args := connectArgs.Msg
	enclaveIdentifier := args.EnclaveIdentifier

	if err := service.enclaveManager.StopEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (service *EngineConnectServerService) DestroyEnclave(ctx context.Context, connectArgs *connect.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect.Response[emptypb.Empty], error) {
	args := connectArgs.Msg
	enclaveIdentifier := args.EnclaveIdentifier

	if err := service.enclaveManager.DestroyEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying enclave with identifier '%v':", args.EnclaveIdentifier)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (service *EngineConnectServerService) Clean(ctx context.Context, connectArgs *connect.Request[kurtosis_engine_rpc_api_bindings.CleanArgs]) (*connect.Response[kurtosis_engine_rpc_api_bindings.CleanResponse], error) {
	args := connectArgs.Msg
	removedEnclaveUuidsAndNames, err := service.enclaveManager.Clean(ctx, args.GetShouldCleanAll())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}
	if args.GetShouldCleanAll() {
		if err = service.logFileManager.RemoveAllLogs(); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred removing all logs.")
		}
	}
	response := &kurtosis_engine_rpc_api_bindings.CleanResponse{RemovedEnclaveNameAndUuids: removedEnclaveUuidsAndNames}
	return connect.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetServiceLogs(ctx context.Context, connectArgs *connect.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs], stream *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]) error {

	args := connectArgs.Msg
	enclaveIdentifier := args.GetEnclaveIdentifier()
	enclaveUuid, err := service.enclaveManager.GetEnclaveUuidForEnclaveIdentifier(context.Background(), enclaveIdentifier)

	contextWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	if err != nil {
		logrus.Errorf("An error occurred while fetching uuid for enclave '%v'. This could happen if the enclave has been deleted. Treating it as UUID", enclaveIdentifier)
		enclaveUuid = enclave.EnclaveUUID(enclaveIdentifier)
	}
	serviceUuidStrSet := args.GetServiceUuidSet()
	requestedServiceUuids := make(map[user_service.ServiceUUID]bool, len(serviceUuidStrSet))
	shouldFollowLogs := args.GetFollowLogs()
	shouldReturnAllLogs := args.GetReturnAllLogs()
	numLogLines := args.GetNumLogLines()

	for serviceUuidStr := range serviceUuidStrSet {
		serviceUuid := user_service.ServiceUUID(serviceUuidStr)
		requestedServiceUuids[serviceUuid] = true
	}

	if service.perWeekLogsDatabaseClient == nil || service.perFileLogsDatabaseClient == nil {
		return stacktrace.NewError("It's not possible to return service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
		errChan                      chan error
		cancelCtxFunc                func()
	)

	notFoundServiceUuids, err := service.reportAnyMissingUuidsAndGetNotFoundUuidsList(contextWithCancel, enclaveUuid, requestedServiceUuids, stream)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reporting missing user service UUIDs for enclave '%v' and requested service UUIDs '%+v'", enclaveUuid, requestedServiceUuids)
	}

	conjunctiveLogLineFilters, err := newConjunctiveLogLineFiltersFromGRPCLogLineFilters(args.GetConjunctiveFilters())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the conjunctive log line filters from the GRPC's conjunctive log line filters '%+v'", args.GetConjunctiveFilters())
	}

	// get enclave creation time to determine strategy to pull logs
	enclaveCreationTime, err := service.getEnclaveCreationTime(ctx, enclaveUuid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to get the enclave creation time to determine how to pull logs.")
	}
	logsDatabaseClient := service.getLogsDatabaseClient(enclaveCreationTime)

	serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, err = logsDatabaseClient.StreamUserServiceLogs(
		contextWithCancel,
		enclaveUuid,
		requestedServiceUuids,
		conjunctiveLogLineFilters,
		shouldFollowLogs,
		shouldReturnAllLogs,
		numLogLines)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred streaming service logs for UUIDs '%+v' in enclave with ID '%v' using filters '%v+' "+
				"and with should follow logs value as '%v'",
			requestedServiceUuids,
			enclaveUuid,
			conjunctiveLogLineFilters,
			shouldFollowLogs,
		)
	}
	defer func() {
		if cancelCtxFunc != nil {
			cancelCtxFunc()
		}
	}()

	for {
		select {
		//stream case
		case serviceLogsByServiceUuid, isChanOpen := <-serviceLogsByServiceUuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				logrus.Debug("Exiting the stream loop after receiving a close signal from the service logs by service UUID channel")
				return nil
			}

			getServiceLogsResponse := newLogsResponse(requestedServiceUuids, serviceLogsByServiceUuid, notFoundServiceUuids)
			if err := stream.Send(getServiceLogsResponse); err != nil {
				return stacktrace.Propagate(err, "An error occurred sending the stream logs for service logs response '%+v'", getServiceLogsResponse)
			}
		//client cancel ctx case
		case <-contextWithCancel.Done():
			logrus.Debug("The user service logs stream has done")
			return nil
		//error from logs database case
		case err, isChanOpen := <-errChan:
			if isChanOpen {
				logrus.Debug("Exiting the stream because an error from the logs database client was received through the error chan.")
				return stacktrace.Propagate(err, "An error occurred streaming user service logs.")
			}
			logrus.Debug("Exiting the stream loop after receiving a close signal from the error chan")
			return nil
		}
	}
}

func (service *EngineConnectServerService) Close() error {
	if err := service.enclaveManager.Close(); err != nil {
		return stacktrace.Propagate(err, "An error occurred closing the enclave manager")
	}
	return nil
}

func (service *EngineConnectServerService) reportAnyMissingUuidsAndGetNotFoundUuidsList(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	requestedServiceUuids map[user_service.ServiceUUID]bool,
	stream *connect.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse],
) (map[string]bool, error) {
	// doesn't matter which logs client is used here
	existingServiceUuids, err := service.perWeekLogsDatabaseClient.FilterExistingServiceUuids(ctx, enclaveUuid, requestedServiceUuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service UUIDs from the log client for enclave '%v' and for the requested UUIDs '%+v'", enclaveUuid, requestedServiceUuids)
	}

	notFoundServiceUuids := getNotFoundServiceUuidsAndEmptyServiceLogsMap(requestedServiceUuids, existingServiceUuids)

	if len(notFoundServiceUuids) == 0 {
		//there is nothing to report
		return notFoundServiceUuids, nil
	}

	emptyServiceLogsByServiceUuid := map[user_service.ServiceUUID][]logline.LogLine{}

	getServiceLogsResponse := newLogsResponse(requestedServiceUuids, emptyServiceLogsByServiceUuid, notFoundServiceUuids)
	if err := stream.Send(getServiceLogsResponse); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred sending the stream logs for service logs response '%+v'", getServiceLogsResponse)
	}

	return notFoundServiceUuids, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func newLogsResponse(
	requestedServiceUuids map[user_service.ServiceUUID]bool,
	serviceLogsByServiceUuid map[user_service.ServiceUUID][]logline.LogLine,
	notFoundServiceUuids map[string]bool,
) *kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse {
	serviceLogLinesByUuid := make(map[string]*kurtosis_engine_rpc_api_bindings.LogLine, len(serviceLogsByServiceUuid))
	for serviceUuid := range requestedServiceUuids {
		serviceUuidStr := string(serviceUuid)
		_, isInNotFoundUuidList := notFoundServiceUuids[serviceUuidStr]
		serviceLogLines, found := serviceLogsByServiceUuid[serviceUuid]
		// should continue in the not-found-UUID list
		if !found && isInNotFoundUuidList {
			continue
		}
		// there is no new log lines but is a found UUID, so it has to be included in the service logs map
		if !found && !isInNotFoundUuidList {
			serviceLogLinesByUuid[serviceUuidStr] = &kurtosis_engine_rpc_api_bindings.LogLine{
				Line: nil,
			}
		}
		//Remove the service's UUID from the initial not found list, if it was returned from the logs database
		//This could happen because some services could send the first log line several minutes after the bootstrap
		if found && isInNotFoundUuidList {
			delete(notFoundServiceUuids, serviceUuidStr)
		}

		logLines := newRPCBindingsLogLineFromLogLines(serviceLogLines)
		serviceLogLinesByUuid[serviceUuidStr] = logLines
	}

	getServiceLogsResponse := &kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse{
		ServiceLogsByServiceUuid: serviceLogLinesByUuid,
		NotFoundServiceUuidSet:   notFoundServiceUuids,
	}
	return getServiceLogsResponse
}

func newRPCBindingsLogLineFromLogLines(logLines []logline.LogLine) *kurtosis_engine_rpc_api_bindings.LogLine {

	logLinesStr := make([]string, len(logLines))

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
	}

	rpcBindingsLogLines := &kurtosis_engine_rpc_api_bindings.LogLine{Line: logLinesStr}

	return rpcBindingsLogLines
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

func newConjunctiveLogLineFiltersFromGRPCLogLineFilters(
	grpcLogLineFilters []*kurtosis_engine_rpc_api_bindings.LogLineFilter,
) (logline.ConjunctiveLogLineFilters, error) {
	var conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters

	for _, grpcLogLineFilter := range grpcLogLineFilters {
		var logLineFilter *logline.LogLineFilter
		operator := grpcLogLineFilter.GetOperator()
		filterTextPattern := grpcLogLineFilter.GetTextPattern()
		switch operator {
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_TEXT:
			logLineFilter = logline.NewDoesContainTextLogLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_TEXT:
			logLineFilter = logline.NewDoesNotContainTextLogLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_MATCH_REGEX:
			logLineFilter = logline.NewDoesContainMatchRegexLogLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX:
			logLineFilter = logline.NewDoesNotContainMatchRegexLogLineFilter(filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in GRPC filter '%v'; this is a bug in Kurtosis", operator, grpcLogLineFilter)
		}
		conjunctiveLogLineFilters = append(conjunctiveLogLineFilters, *logLineFilter)
	}

	return conjunctiveLogLineFilters, nil
}

// If the enclave was created prior to log retention, return the per file logs client
func (service *EngineConnectServerService) getLogsDatabaseClient(enclaveCreationTime time.Time) centralized_logs.LogsDatabaseClient {
	if enclaveCreationTime.After(logRetentionFeatureReleaseTime) {
		return service.perWeekLogsDatabaseClient
	} else {
		return service.perFileLogsDatabaseClient
	}
}

func (service *EngineConnectServerService) getEnclaveCreationTime(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (time.Time, error) {
	enclaves, err := service.enclaveManager.GetEnclaves(ctx)
	if err != nil {
		return time.Time{}, err
	}

	enclaveObj, found := enclaves[string(enclaveUuid)]
	if !found {
		return time.Time{}, stacktrace.NewError("Engine could not find enclave '%v'", enclaveUuid)
	}

	timestamp := enclaveObj.GetCreationTime()
	if timestamp == nil {
		return time.Time{}, stacktrace.NewError("An error occurred getting the creation time for enclave '%v'. This is a bug in Kurtosis", enclaveUuid)
	}
	return timestamp.AsTime(), nil
}
