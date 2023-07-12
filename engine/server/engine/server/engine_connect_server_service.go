package server

import (
	"context"
	"errors"
	connect_go "github.com/bufbuild/connect-go"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EngineConnectServerService struct {
	// The version tag of the engine server image, so it can report its own version
	imageVersionTag string

	enclaveManager *enclave_manager.EnclaveManager

	//The protected user ID for metrics analytics purpose
	metricsUserID string

	//User consent to send metrics
	didUserAcceptSendingMetrics bool

	//The client for consuming container logs from the logs' database server
	logsDatabaseClient centralized_logs.LogsDatabaseClient
}

func NewEngineConnectServerService(
	imageVersionTag string,
	enclaveManager *enclave_manager.EnclaveManager,
	metricsUserId string,
	didUserAcceptSendingMetrics bool,
	logsDatabaseClient centralized_logs.LogsDatabaseClient,
) *EngineConnectServerService {
	service := &EngineConnectServerService{
		imageVersionTag:             imageVersionTag,
		enclaveManager:              enclaveManager,
		metricsUserID:               metricsUserId,
		didUserAcceptSendingMetrics: didUserAcceptSendingMetrics,
		logsDatabaseClient:          logsDatabaseClient,
	}
	return service
}

func (service *EngineConnectServerService) GetEngineInfo(context.Context, *connect_go.Request[emptypb.Empty]) (*connect_go.Response[kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse], error) {
	logrus.Infof("HURRY CALLED HERE")
	result := &kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse{
		EngineVersion: service.imageVersionTag,
	}
	return connect_go.NewResponse(result), nil
}

func (service *EngineConnectServerService) CreateEnclave(ctx context.Context, connectArgs *connect_go.Request[kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs]) (*connect_go.Response[kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse], error) {
	args := connectArgs.Msg
	apiContainerLogLevel, err := logrus.ParseLevel(args.ApiContainerLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.ApiContainerLogLevel)
	}

	enclaveInfo, err := service.enclaveManager.CreateEnclave(
		ctx,
		args.ApiContainerVersionTag,
		apiContainerLogLevel,
		args.EnclaveName,
		args.IsPartitioningEnabled,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with name '%v'", args.EnclaveName)
	}

	response := &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
		EnclaveInfo: enclaveInfo,
	}

	return connect_go.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetEnclaves(ctx context.Context, _ *connect_go.Request[emptypb.Empty]) (*connect_go.Response[kurtosis_engine_rpc_api_bindings.GetEnclavesResponse], error) {
	infoForEnclaves, err := service.enclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	response := &kurtosis_engine_rpc_api_bindings.GetEnclavesResponse{EnclaveInfo: infoForEnclaves}
	return connect_go.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetExistingAndHistoricalEnclaveIdentifiers(context.Context, *connect_go.Request[emptypb.Empty]) (*connect_go.Response[kurtosis_engine_rpc_api_bindings.GetExistingAndHistoricalEnclaveIdentifiersResponse], error) {
	allIdentifiers, err := service.enclaveManager.GetExistingAndHistoricalEnclaveIdentifiers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching enclave identifiers")
	}
	response := &kurtosis_engine_rpc_api_bindings.GetExistingAndHistoricalEnclaveIdentifiersResponse{AllIdentifiers: allIdentifiers}
	return connect_go.NewResponse(response), nil
}

func (service *EngineConnectServerService) StopEnclave(ctx context.Context, connectArgs *connect_go.Request[kurtosis_engine_rpc_api_bindings.StopEnclaveArgs]) (*connect_go.Response[emptypb.Empty], error) {
	args := connectArgs.Msg
	enclaveIdentifier := args.EnclaveIdentifier

	if err := service.enclaveManager.StopEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
	}

	return connect_go.NewResponse(&emptypb.Empty{}), nil
}

func (service *EngineConnectServerService) DestroyEnclave(ctx context.Context, connectArgs *connect_go.Request[kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs]) (*connect_go.Response[emptypb.Empty], error) {
	args := connectArgs.Msg
	enclaveIdentifier := args.EnclaveIdentifier
	if err := service.enclaveManager.DestroyEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying enclave with identifier '%v':", args.EnclaveIdentifier)
	}

	return connect_go.NewResponse(&emptypb.Empty{}), nil
}

func (service *EngineConnectServerService) Clean(ctx context.Context, connectArgs *connect_go.Request[kurtosis_engine_rpc_api_bindings.CleanArgs]) (*connect_go.Response[kurtosis_engine_rpc_api_bindings.CleanResponse], error) {
	args := connectArgs.Msg
	removedEnclaveUuidsAndNames, err := service.enclaveManager.Clean(ctx, args.ShouldCleanAll)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}

	response := &kurtosis_engine_rpc_api_bindings.CleanResponse{RemovedEnclaveNameAndUuids: removedEnclaveUuidsAndNames}
	return connect_go.NewResponse(response), nil
}

func (service *EngineConnectServerService) GetServiceLogs(ctx context.Context, connectArgs *connect_go.Request[kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs], stream *connect_go.ServerStream[kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse]) error {
	return connect_go.NewError(connect_go.CodeUnimplemented, errors.New("engine_api.EngineService.GetServiceLogs is not implemented"))
}
