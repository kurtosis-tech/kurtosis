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

func (service *EngineServerService) GetUserServiceLogs(
	args *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsArgs,
	stream kurtosis_engine_rpc_api_bindings.EngineService_GetUserServiceLogsServer,
) error {

	enclaveId := enclave.EnclaveID(args.GetEnclaveId())
	userServiceGuidStrSet := args.GetServiceGuidSet()
	requestedUserServiceGuids := make(map[user_service.ServiceGUID]bool, len(userServiceGuidStrSet))

	for userServiceGuidStr := range userServiceGuidStrSet {
		userServiceGuid := user_service.ServiceGUID(userServiceGuidStr)
		requestedUserServiceGuids[userServiceGuid] = true
	}

	if service.logsDatabaseClient == nil {
		return stacktrace.NewError("It's not possible to return user service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		userServiceLogsByServiceGuidChan chan map[user_service.ServiceGUID][]centralized_logs.LogLine
		errChan          chan error
		cancelStreamFunc func()
		err              error
	)

	if err := service.reportAnyMissingGuids(enclaveId, requestedUserServiceGuids, stream); err != nil {
		return stacktrace.Propagate(err, "An error occurred reporting missing user service GUIDs for enclave '%v' and requested user service GUIDs '%+v'", enclaveId, requestedUserServiceGuids)
	}

	if args.FollowLogs {
		userServiceLogsByServiceGuidChan, errChan, cancelStreamFunc, err = service.logsDatabaseClient.StreamUserServiceLogs(stream.Context(), enclaveId, requestedUserServiceGuids)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred streaming user service logs for GUIDs '%+v' in enclave with ID '%v'", requestedUserServiceGuids, enclaveId)
		}
	} else {
		userServiceLogsByServiceGuidChan, errChan, cancelStreamFunc, err = service.logsDatabaseClient.GetUserServiceLogs(stream.Context(), enclaveId, requestedUserServiceGuids)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred streaming user service logs for GUIDs '%+v' in enclave with ID '%v'", requestedUserServiceGuids, enclaveId)
		}
	}
	defer cancelStreamFunc()

	for {
		select {
		//stream case
		case userServiceLogsByServiceGuid, isChanOpen := <-userServiceLogsByServiceGuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				break
			}

			// We also have to do this check on every stream send because the GUIDs could be added in the logs DB at some point
			newExistingServiceGuids, err := service.logsDatabaseClient.FilterExistingServiceGuids(stream.Context(), enclaveId, requestedUserServiceGuids)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service GUIDs from the log client for enclave '%v' and for the requested GUIDs '%+v'", enclaveId, requestedUserServiceGuids)
			}

			notFoundUserServiceGuids := getNotFoundUserServiceGuids(requestedUserServiceGuids, newExistingServiceGuids)

			getUserServiceLogsResponse := newUserLogsResponse(requestedUserServiceGuids, userServiceLogsByServiceGuid, notFoundUserServiceGuids)
			if err := stream.Send(getUserServiceLogsResponse); err != nil {
				return stacktrace.Propagate(err, "An error occurred sending the stream logs for user service logs response '%+v'", getUserServiceLogsResponse)
			}
		//client cancel ctx case
		case <-stream.Context().Done():
			logrus.Debug("The user service logs stream has done")
			return nil
		//error from logs database case
		case err := <-errChan:
			return stacktrace.Propagate(err,"An error occurred streaming user service logs")
		}
	}

}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (service *EngineServerService) reportAnyMissingGuids(
	enclaveId enclave.EnclaveID,
	requestedUserServiceGuids map[user_service.ServiceGUID]bool,
	stream kurtosis_engine_rpc_api_bindings.EngineService_GetUserServiceLogsServer,
) error {
	existingServiceGuids, err := service.logsDatabaseClient.FilterExistingServiceGuids(stream.Context(), enclaveId, requestedUserServiceGuids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service GUIDs from the log client for enclave '%v' and for the requested GUIDs '%+v'", enclaveId, requestedUserServiceGuids)
	}

	notFoundUserServiceGuids := getNotFoundUserServiceGuids(requestedUserServiceGuids, existingServiceGuids)

	if len(notFoundUserServiceGuids) == 0 {
		//there is nothing to report
		return nil
	}

	emptyUserServiceLogsByServiceGuid := map[user_service.ServiceGUID][]centralized_logs.LogLine{}
	for serviceGuid := range requestedUserServiceGuids {
		emptyUserServiceLogsByServiceGuid[serviceGuid] = []centralized_logs.LogLine{}
	}

	getUserServiceLogsResponse := newUserLogsResponse(requestedUserServiceGuids, emptyUserServiceLogsByServiceGuid, notFoundUserServiceGuids)
	if err := stream.Send(getUserServiceLogsResponse); err != nil {
		return stacktrace.Propagate(err, "An error occurred sending the stream logs for user service logs response '%+v'", getUserServiceLogsResponse)
	}

	return nil
}

func newUserLogsResponse(
	requestedUserServiceGuids map[user_service.ServiceGUID]bool,
	userServiceLogsByUserServiceGuid map[user_service.ServiceGUID][]centralized_logs.LogLine,
	notFoundUserServiceGuids map[string]bool,
) *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse {
	userServiceLogLinesByGuid := make(map[string]*kurtosis_engine_rpc_api_bindings.LogLine, len(userServiceLogsByUserServiceGuid))

	for userServiceGuid := range requestedUserServiceGuids {
		userServiceGuidStr := string(userServiceGuid)
		userServiceLogLines, found := userServiceLogsByUserServiceGuid[userServiceGuid]
		if !found {
			userServiceLogLinesByGuid[userServiceGuidStr] = &kurtosis_engine_rpc_api_bindings.LogLine{}
		}
		logLines := newRPCBindingsLogLineFromLogLines(userServiceLogLines)
		userServiceLogLinesByGuid[userServiceGuidStr] = logLines
	}

	getUserServiceLogsResponse := &kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse{
		UserServiceLogsByUserServiceGuid: userServiceLogLinesByGuid,
		NotFoundUserServiceGuidSet: notFoundUserServiceGuids,
	}
	return getUserServiceLogsResponse
}

func newRPCBindingsLogLineFromLogLines(logLines []centralized_logs.LogLine) *kurtosis_engine_rpc_api_bindings.LogLine {

	logLinesStr := make([]string, len(logLines))

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
	}

	rpcBindingsLogLines := &kurtosis_engine_rpc_api_bindings.LogLine{Line: logLinesStr}

	return rpcBindingsLogLines
}

func getNotFoundUserServiceGuids(
	requestedUserServiceGuids map[user_service.ServiceGUID]bool,
	existingServiceGuids map[user_service.ServiceGUID]bool,
) map[string]bool {
	notFoundUserServiceGuids := map[string]bool{}

	for requestedUserServiceGuid := range requestedUserServiceGuids {
		if _, found := existingServiceGuids[requestedUserServiceGuid]; !found {
			requestedUserServiceGuidStr := string(requestedUserServiceGuid)
			notFoundUserServiceGuids[requestedUserServiceGuidStr] = true
		}
	}

	return notFoundUserServiceGuids
}
