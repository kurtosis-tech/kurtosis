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
	// This embedding is required by gRPC
	kurtosis_engine_rpc_api_bindings.UnimplementedEngineServiceServer

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
		logsDatabaseClient: logsDatabaseClient,
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
	userServiceGuidsStr := args.GetServiceGuidSet()
	userServiceGuids := make(map[user_service.ServiceGUID]bool, len(userServiceGuidsStr))

	for userServiceGuidStr := range userServiceGuidsStr {
		userServiceGuid := user_service.ServiceGUID(userServiceGuidStr)
		userServiceGuids[userServiceGuid] = true
	}

	userServiceLogsByUserServiceGuid, err := service.logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs for GUIDs '%+v' in enclave with ID '%v'", userServiceGuids, enclaveId)
	}

	userServiceLogLinesByUserServiceGuid := make(map[string]*kurtosis_engine_rpc_api_bindings.LogLine, len(userServiceLogsByUserServiceGuid))

	for userServiceGuid, logLinesStr := range userServiceLogsByUserServiceGuid {
		userServiceGuidStr := string(userServiceGuid)
		logLine := &kurtosis_engine_rpc_api_bindings.LogLine{Line: logLinesStr}
		userServiceLogLinesByUserServiceGuid[userServiceGuidStr] = logLine
	}

	response := &kurtosis_engine_rpc_api_bindings.GetUserServiceLogsResponse{UserServiceLogsByUserServiceGuid: userServiceLogLinesByUserServiceGuid}
	return response, nil

}
