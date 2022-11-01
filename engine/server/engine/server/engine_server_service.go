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
	ctx context.Context,
	args *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsArgs,
) (*kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse, error) {
	enclaveId := enclave.EnclaveID(args.GetEnclaveId())
	userServiceGuidStrSet := args.GetServiceGuidSet()
	requestedUserServiceGuids := make(map[user_service.ServiceGUID]bool, len(userServiceGuidStrSet))

	for userServiceGuidStr := range userServiceGuidStrSet {
		userServiceGuid := user_service.ServiceGUID(userServiceGuidStr)
		requestedUserServiceGuids[userServiceGuid] = true
	}

	if service.logsDatabaseClient == nil {
		return nil, stacktrace.NewError("It's not possible to return user service logs because there is not logs database client; this is bug in Kurtosis")
	}

	userServiceLogsByUserServiceGuid, err := service.logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, requestedUserServiceGuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs for GUIDs '%+v' in enclave with ID '%v'", requestedUserServiceGuids, enclaveId)
	}

	getUserServiceLogsResponse := newUserLogsResponseFromUserServiceLogsByGuid(requestedUserServiceGuids, userServiceLogsByUserServiceGuid)

	return getUserServiceLogsResponse, nil

}

func (service *EngineServerService) StreamUserServiceLogs(
	args *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsArgs,
	stream kurtosis_engine_rpc_api_bindings.EngineService_StreamUserServiceLogsServer,
) error {

	enclaveId := enclave.EnclaveID(args.GetEnclaveId())
	userServiceGuidStrSet := args.GetServiceGuidSet()
	requestedUserServiceGuids := make(map[user_service.ServiceGUID]bool, len(userServiceGuidStrSet))

	for userServiceGuidStr := range userServiceGuidStrSet {
		userServiceGuid := user_service.ServiceGUID(userServiceGuidStr)
		requestedUserServiceGuids[userServiceGuid] = true
	}

	if service.logsDatabaseClient == nil {
		return stacktrace.NewError("It's not possible to return user service logs because there is not logs database client; this is bug in Kurtosis")
	}

	userServiceLogsByServiceGuidChan, errChan, cancelLogsDatabaseStreamFunc, err := service.logsDatabaseClient.StreamUserServiceLogs(stream.Context(), enclaveId, requestedUserServiceGuids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred streaming user service logs for GUIDs '%+v' in enclave with ID '%v'", requestedUserServiceGuids, enclaveId)
	}

	//this channel will produce a signal when the logs database read stream thread has finished
	readLogsDatabaseHasFinishedSignaller := make(chan struct{})

	go runStreamUserServiceLogsRoutine(
		requestedUserServiceGuids,
		userServiceLogsByServiceGuidChan,
		errChan,
		stream,
		readLogsDatabaseHasFinishedSignaller,
	)

	go runStreamCancellationRoutine(
		stream,
		errChan,
		readLogsDatabaseHasFinishedSignaller,
		cancelLogsDatabaseStreamFunc,
	)

	return nil
}

func runStreamUserServiceLogsRoutine(
	requestedUserServiceGuids map[user_service.ServiceGUID]bool,
	userServiceLogsByServiceGuidChan chan map[user_service.ServiceGUID][]centralized_logs.LogLine,
	errChan chan error,
	stream kurtosis_engine_rpc_api_bindings.EngineService_StreamUserServiceLogsServer,
	readLogsDatabaseHasFinishedSignaller chan struct{},
) {
	for {
		userServiceLogsByServiceGuid, isChanOpen := <-userServiceLogsByServiceGuidChan
		//If the channel is closed means that the logs database client won't continue sending streams
		if !isChanOpen {
			break
		}

		getUserServiceLogsResponse := newUserLogsResponseFromUserServiceLogsByGuid(requestedUserServiceGuids, userServiceLogsByServiceGuid)
		if err := stream.Send(getUserServiceLogsResponse); err != nil {
			errChan <- stacktrace.Propagate(err, "An error occurred sending the stream logs for user service logs response '%+v'", getUserServiceLogsResponse)
		}
	}

	readLogsDatabaseHasFinishedSignaller <- struct{}{}
}

func runStreamCancellationRoutine(
	stream kurtosis_engine_rpc_api_bindings.EngineService_StreamUserServiceLogsServer,
	errChan chan error,
	readLogsDatabaseHasFinishedSignaller chan struct{},
	cancelLogsDatabaseStreamFunc func(),
) {
	defer cancelLogsDatabaseStreamFunc()

	for  {
		select {
		case <-stream.Context().Done():
			logrus.Debug("The user service logs stream has done")
			return
		case err := <- errChan:
			logrus.Errorf("An error occurred streaming user service logs. Err:\n%v", err)
			return
		case <-readLogsDatabaseHasFinishedSignaller:
			logrus.Debug("Received the read logs database has finished signal")
			return
		}
	}
}

func newUserLogsResponseFromUserServiceLogsByGuid(requestedUserServiceGuids map[user_service.ServiceGUID]bool, userServiceLogsByUserServiceGuid map[user_service.ServiceGUID][]centralized_logs.LogLine) *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse {
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

	getUserServiceLogsResponse := &kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse{UserServiceLogsByUserServiceGuid: userServiceLogLinesByGuid}
	return getUserServiceLogsResponse
}

func newRPCBindingsLogLineFromLogLines(logLines []centralized_logs.LogLine) *kurtosis_engine_rpc_api_bindings.LogLine {

	logLinesStr := make([]string, len(logLines))

	for _, logLine := range logLines {
		logLinesStr = append(logLinesStr, logLine.GetContent())
	}

	rpcBindingsLogLines := &kurtosis_engine_rpc_api_bindings.LogLine{Line: logLinesStr}

	return rpcBindingsLogLines
}