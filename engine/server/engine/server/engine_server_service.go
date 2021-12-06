package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/server/engine/enclave_manager"
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
}

func NewEngineServerService(imageVersionTag string, enclaveManager *enclave_manager.EnclaveManager) *EngineServerService {
	service := &EngineServerService{
		imageVersionTag:                  imageVersionTag,
		enclaveManager:                   enclaveManager,
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
		args.ShouldPublishAllPorts,
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
	if err := service.enclaveManager.StopEnclave(ctx, enclaveId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
	}
	return &emptypb.Empty{}, nil
}

func (service *EngineServerService) DestroyEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs) (*emptypb.Empty, error) {

	if err := service.enclaveManager.DestroyEnclave(ctx, args.EnclaveId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying enclave with ID '%v':", args.EnclaveId)
	}

	return &emptypb.Empty{}, nil
}

func (service *EngineServerService) Clean (ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CleanArgs) (*kurtosis_engine_rpc_api_bindings.CleanResponse, error){
	artifactIDs , err := service.enclaveManager.Clean(ctx, args.ShouldCleanAll)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}
	response := &kurtosis_engine_rpc_api_bindings.CleanResponse{Artifacts: artifactIDs}
	return response, nil
}
