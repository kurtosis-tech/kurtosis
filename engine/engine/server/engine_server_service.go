package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_api_version"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type EngineServerService struct {
	// This embedding is required by gRPC
	// kurtosis_engine_rpc_api_bindings.UnimplementedEngineServiceServer

	enclaveManager *enclave_manager.EnclaveManager
}

func NewEngineServerService(enclaveManager *enclave_manager.EnclaveManager) *EngineServerService {
	service := &EngineServerService{
		enclaveManager: enclaveManager,
	}
	return service
}

func (service *EngineServerService) GetEngineInfo(ctx context.Context, empty *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse, error) {
	result := &kurtosis_engine_rpc_api_bindings.GetEngineInfoResponse{
		EngineApiVersion: kurtosis_engine_api_version.KurtosisEngineApiVersion,
	}
	return result, nil
}

func (service *EngineServerService) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error) {

	apiContainerLogLevel, err := logrus.ParseLevel(args.ApiContainerLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.ApiContainerLogLevel)
	}

	enclave, err := service.enclaveManager.CreateEnclave(
		ctx,
		args.ApiContainerImage,
		apiContainerLogLevel,
		args.EnclaveId,
		args.IsPartitioningEnabled,
		args.ShouldPublishAllPorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with ID '%v'", args.EnclaveId)
	}

	response := &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
		NetworkId:                   enclave.GetNetworkId(),
		NetworkCidr:                 enclave.GetNetworkIpAndMask().String(),
		ApiContainerId:              enclave.GetApiContainerId(),
		ApiContainerIpInsideNetwork: enclave.GetApiContainerIpAddr().String(),
		ApiContainerHostIp:          enclave.GetApiContainerHostPortBinding().HostIP,
		ApiContainerHostPort:        enclave.GetApiContainerHostPortBinding().HostPort,
	}

	return response, nil
}

func (service *EngineServerService) GetEnclaves(ctx context.Context, empty *emptypb.Empty) (*kurtosis_engine_rpc_api_bindings.GetEnclavesResponse, error) {
	enclave, err := service.enclaveManager.GetEnclave(ctx, args.EnclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave with ID '%v'", args.EnclaveId)
	}

	response := &kurtosis_engine_rpc_api_bindings.GetEnclaveResponse{
		NetworkId:                   enclave.GetNetworkId(),
		NetworkCidr:                 enclave.GetNetworkIpAndMask().String(),
		ApiContainerId:              enclave.GetApiContainerId(),
		ApiContainerIpInsideNetwork: enclave.GetApiContainerIpAddr().String(),
		ApiContainerHostIp:          enclave.GetApiContainerHostPortBinding().HostIP,
		ApiContainerHostPort:        enclave.GetApiContainerHostPortBinding().HostPort,
	}

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
