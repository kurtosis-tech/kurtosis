package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EngineServerService struct {
	// The version tag of the engine server image, so it can report its own version
	imageVersionTag string

	enclaveManager *enclave_manager.EnclaveManager

	metricsClient client.MetricsClient

	//The protected user ID for metrics analytics purpose
	metricsUserID string

	//User consent to send metrics
	didUserAcceptSendingMetrics bool

	//The client for consuming container logs from the logs' database server
	logsDatabaseClient centralized_logs.LogsDatabaseClient
}

func NewEngineServerService(
	imageVersionTag string,
	enclaveManager *enclave_manager.EnclaveManager,
	metricsClient client.MetricsClient,
	metricsUserId string,
	didUserAcceptSendingMetrics bool,
	logsDatabaseClient centralized_logs.LogsDatabaseClient,
) *EngineServerService {
	service := &EngineServerService{
		imageVersionTag:             imageVersionTag,
		enclaveManager:              enclaveManager,
		metricsClient:               metricsClient,
		metricsUserID:               metricsUserId,
		didUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		logsDatabaseClient:          logsDatabaseClient,
	}
	return service
}

func (service *EngineServerService) GetEngineInfo(ctx context.Context, empty *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	result := &kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse{
		EngineVersion: service.imageVersionTag,
	}
	return result, nil
}

func (service *EngineServerService) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error) {
	if err := service.metricsClient.TrackCreateEnclave(args.EnclaveId); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking create enclave event\n%v", err)
	}

	apiContainerLogLevel, err := logrus.ParseLevel(args.ApiContainerLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.ApiContainerLogLevel)
	}

	enclaveInfo, err := service.enclaveManager.CreateEnclave(
		ctx,
		args.ApiContainerVersionTag,
		apiContainerLogLevel,
		args.EnclaveId,
		args.IsPartitioningEnabled,
		service.metricsUserID,
		service.didUserAcceptSendingMetrics,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with ID '%v'", args.EnclaveId)
	}

	response := &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
		EnclaveInfo: enclaveInfo,
	}

	return response, nil
}

func (service *EngineServerService) GetEnclaves(ctx context.Context, _ *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	infoForEnclaves, err := service.enclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	response := &kurtosis_engine_rpc_api_bindings.GetEnclavesResponse{EnclaveInfo: infoForEnclaves}
	return response, nil
}

func (service *EngineServerService) StopEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.StopEnclaveArgs) (*emptypb.Empty, error) {
	enclaveId := args.EnclaveId

	if err := service.metricsClient.TrackStopEnclave(enclaveId); err != nil {
		//We don't want to interrupt user's flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking stop enclave event\n%v", err)
	}

	if err := service.enclaveManager.StopEnclave(ctx, enclaveId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
	}

	return &emptypb.Empty{}, nil
}

func (service *EngineServerService) DestroyEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs) (*emptypb.Empty, error) {
	enclaveId := args.EnclaveId

	if err := service.metricsClient.TrackDestroyEnclave(enclaveId); err != nil {
		//We don't want to interrupt user's flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking destroy enclave event\n%v", err)
	}

	if err := service.enclaveManager.DestroyEnclave(ctx, enclaveId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying enclave with ID '%v':", args.EnclaveId)
	}

	return &emptypb.Empty{}, nil
}

func (service *EngineServerService) Clean(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CleanArgs) (*kurtosis_engine_rpc_api_bindings.CleanResponse, error) {
	enclaveIDs, err := service.enclaveManager.Clean(ctx, args.ShouldCleanAll)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}

	response := &kurtosis_engine_rpc_api_bindings.CleanResponse{RemovedEnclaveIds: enclaveIDs}
	return response, nil
}

func (service *EngineServerService) GetServiceLogs(
	args *kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs,
	stream kurtosis_engine_rpc_api_bindings.EngineService_GetServiceLogsServer,
) error {

	enclaveId := enclave.EnclaveID(args.GetEnclaveId())
	serviceGuidStrSet := args.GetServiceGuidSet()
	requestedServiceGuids := make(map[user_service.ServiceGUID]bool, len(serviceGuidStrSet))
	shouldFollowLogs := args.FollowLogs

	for serviceGuidStr := range serviceGuidStrSet {
		serviceGuid := user_service.ServiceGUID(serviceGuidStr)
		requestedServiceGuids[serviceGuid] = true
	}

	if service.logsDatabaseClient == nil {
		return stacktrace.NewError("It's not possible to return service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		serviceLogsByServiceGuidChan chan map[user_service.ServiceGUID][]centralized_logs.LogLine
		errChan                      chan error
		cancelStreamFunc             func()
		err                          error
	)

	notFoundServiceGuids, err := service.reportAnyMissingGuidsAndGetNotFoundGuidsList(enclaveId, requestedServiceGuids, stream)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reporting missing user service GUIDs for enclave '%v' and requested service GUIDs '%+v'", enclaveId, requestedServiceGuids)
	}

	conjunctiveLogLineFilters, err := newConjunctiveLogLineFiltersGRPC(args.GetConjunctiveFilters())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the conjunctive log line filters from the GRPC's conjunctive log line filters '%+v'", args.GetConjunctiveFilters())
	}

	if shouldFollowLogs {
		serviceLogsByServiceGuidChan, errChan, cancelStreamFunc, err = service.logsDatabaseClient.StreamUserServiceLogs(stream.Context(), enclaveId, requestedServiceGuids, conjunctiveLogLineFilters)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred streaming service logs for GUIDs '%+v' in enclave with ID '%v'", requestedServiceGuids, enclaveId)
		}
	} else {
		serviceLogsByServiceGuidChan, errChan, cancelStreamFunc, err = service.logsDatabaseClient.GetUserServiceLogs(stream.Context(), enclaveId, requestedServiceGuids, conjunctiveLogLineFilters)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred streaming service logs for GUIDs '%+v' in enclave with ID '%v'", requestedServiceGuids, enclaveId)
		}
	}
	defer cancelStreamFunc()

	for {
		select {
		//stream case
		case serviceLogsByServiceGuid, isChanOpen := <-serviceLogsByServiceGuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				break
			}

			getServiceLogsResponse := newLogsResponse(requestedServiceGuids, serviceLogsByServiceGuid, notFoundServiceGuids)
			if err := stream.Send(getServiceLogsResponse); err != nil {
				return stacktrace.Propagate(err, "An error occurred sending the stream logs for service logs response '%+v'", getServiceLogsResponse)
			}
			if !shouldFollowLogs {
				logrus.Debug("User requested to not follow the logs, so the logs stream is closed after sending all the logs created at this point")
				return nil
			}
		//client cancel ctx case
		case <-stream.Context().Done():
			logrus.Debug("The user service logs stream has done")
			return nil
		//error from logs database case
		case err := <-errChan:
			return stacktrace.Propagate(err, "An error occurred streaming user service logs")
		}
	}

}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (service *EngineServerService) reportAnyMissingGuidsAndGetNotFoundGuidsList(
	enclaveId enclave.EnclaveID,
	requestedServiceGuids map[user_service.ServiceGUID]bool,
	stream kurtosis_engine_rpc_api_bindings.EngineService_GetServiceLogsServer,
) (map[string]bool, error) {
	existingServiceGuids, err := service.logsDatabaseClient.FilterExistingServiceGuids(stream.Context(), enclaveId, requestedServiceGuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service GUIDs from the log client for enclave '%v' and for the requested GUIDs '%+v'", enclaveId, requestedServiceGuids)
	}

	notFoundServiceGuids := getNotFoundServiceGuidsAndEmptyServiceLogsMap(requestedServiceGuids, existingServiceGuids)

	if len(notFoundServiceGuids) == 0 {
		//there is nothing to report
		return notFoundServiceGuids, nil
	}

	emptyServiceLogsByServiceGuid := map[user_service.ServiceGUID][]centralized_logs.LogLine{}

	getServiceLogsResponse := newLogsResponse(requestedServiceGuids, emptyServiceLogsByServiceGuid, notFoundServiceGuids)
	if err := stream.Send(getServiceLogsResponse); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred sending the stream logs for service logs response '%+v'", getServiceLogsResponse)
	}

	return notFoundServiceGuids, nil
}

func newLogsResponse(
	requestedServiceGuids map[user_service.ServiceGUID]bool,
	serviceLogsByServiceGuid map[user_service.ServiceGUID][]centralized_logs.LogLine,
	notFoundServiceGuids map[string]bool,
) *kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse {
	serviceLogLinesByGuid := make(map[string]*kurtosis_engine_rpc_api_bindings.LogLine, len(serviceLogsByServiceGuid))

	for serviceGuid := range requestedServiceGuids {
		serviceGuidStr := string(serviceGuid)
		_, isInNotFoundGuidList := notFoundServiceGuids[serviceGuidStr]
		serviceLogLines, found := serviceLogsByServiceGuid[serviceGuid]
		// should continue in the not-found-GUID list
		if !found && isInNotFoundGuidList {
			continue
		}
		// there is no new log lines but is a found GUID, so it has to be included in the service logs map
		if !found && !isInNotFoundGuidList {
			serviceLogLinesByGuid[serviceGuidStr] = &kurtosis_engine_rpc_api_bindings.LogLine{
				Line: nil,
			}
		}
		//Remove the service's GUID from the initial not found list, if it was returned from the logs database
		//This could happen because some services could send the first log line several minutes after the bootstrap
		if found && isInNotFoundGuidList {
			delete(notFoundServiceGuids, serviceGuidStr)
		}

		logLines := newRPCBindingsLogLineFromLogLines(serviceLogLines)
		serviceLogLinesByGuid[serviceGuidStr] = logLines
	}

	getServiceLogsResponse := &kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse{
		ServiceLogsByServiceGuid: serviceLogLinesByGuid,
		NotFoundServiceGuidSet:   notFoundServiceGuids,
	}
	return getServiceLogsResponse
}

func newRPCBindingsLogLineFromLogLines(logLines []centralized_logs.LogLine) *kurtosis_engine_rpc_api_bindings.LogLine {

	logLinesStr := make([]string, len(logLines))

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
	}

	rpcBindingsLogLines := &kurtosis_engine_rpc_api_bindings.LogLine{Line: logLinesStr}

	return rpcBindingsLogLines
}

func getNotFoundServiceGuidsAndEmptyServiceLogsMap(
	requestedServiceGuids map[user_service.ServiceGUID]bool,
	existingServiceGuids map[user_service.ServiceGUID]bool,
) map[string]bool {
	notFoundServiceGuids := map[string]bool{}

	for requestedServiceGuid := range requestedServiceGuids {
		if _, found := existingServiceGuids[requestedServiceGuid]; !found {
			requestedServiceGuidStr := string(requestedServiceGuid)
			notFoundServiceGuids[requestedServiceGuidStr] = true
		}
	}

	return notFoundServiceGuids
}

func newConjunctiveLogLineFiltersGRPC(
	logLineFilters []*kurtosis_engine_rpc_api_bindings.LogLineFilter,
) (centralized_logs.ConjunctiveLogLineFilters, error) {
	var lokiLogLineFilters []centralized_logs.LokiLineFilter

	for _, logLineFilter := range logLineFilters {
		var lokiLogLineFilter *centralized_logs.LokiLineFilter
		operator := logLineFilter.GetOperator()
		filterTextPattern := logLineFilter.GetTextPattern()
		switch operator {
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_TEXT:
			lokiLogLineFilter = centralized_logs.NewDoesContainTextLokiLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_TEXT:
			lokiLogLineFilter = centralized_logs.NewDoesNotContainTextLokiLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_MATCH_REGEX:
			lokiLogLineFilter = centralized_logs.NewDoesContainMatchRegexLokiLineFilter(filterTextPattern)
		case kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX:
			lokiLogLineFilter = centralized_logs.NewDoesNotContainMatchRegexLokiLineFilter(filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		lokiLogLineFilters = append(lokiLogLineFilters, *lokiLogLineFilter)
	}

	logPipeline := centralized_logs.NewLokiLogPipeline(lokiLogLineFilters)

	return logPipeline, nil
}
