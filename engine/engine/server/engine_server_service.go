package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager"
)

type EngineServerService struct {
	// This embedding is required by gRPC
	kurtosis_engine_rpc_api_bindings.UnimplementedEngineServiceServer

	enclaveManager *enclave_manager.EnclaveManager
}

func NewEngineServerService(enclaveManager *enclave_manager.EnclaveManager) *EngineServerService {
	service := &EngineServerService{
		enclaveManager: enclaveManager,
	}
	return service
}

func (service *EngineServerService) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error){

}
