package api_container_gateway

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
)

type ApiContainerGatewayServiceServer struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	// Client for the api container we'll be connecting too
	remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	// Provides connections to Kurtosis objectis in cluster
	connectionProvider *connection.GatewayConnectionProvider

	// ServiceMap
	mutex                           *sync.Mutex
	userServiceToLocalConnectionMap map[string]connection.GatewayConnectionToKurtosis
}

func NewEnclaveApiContainerGatewayServer(connectionProvider *connection.GatewayConnectionProvider, remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient) (resultCoreGatewayServerService *ApiContainerGatewayServiceServer, resultGatewayCloseFunc func()) {
	// Start out with 0 connections to user services
	userServiceToLocalConnectionMap := map[string]connection.GatewayConnectionToKurtosis{}

	closeGatewayFunc := func() {
		// Stop any port forwarding
		for _, conn := range resultCoreGatewayServerService.userServiceToLocalConnectionMap {
			conn.Stop()
		}
	}

	return &ApiContainerGatewayServiceServer{
		remoteApiContainerClient:        remoteApiContainerClient,
		connectionProvider:              connectionProvider,
		mutex:                           &sync.Mutex{},
		userServiceToLocalConnectionMap: userServiceToLocalConnectionMap,
	}, closeGatewayFunc
}

func (service *ApiContainerGatewayServiceServer) LoadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadModuleArgs) (*kurtosis_core_rpc_api_bindings.LoadModuleResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.LoadModule(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) UnloadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnloadModuleArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.UnloadModule(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) ExecuteModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteModuleArgs) (*kurtosis_core_rpc_api_bindings.ExecuteModuleResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ExecuteModule(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) GetModuleInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetModuleInfoArgs) (*kurtosis_core_rpc_api_bindings.GetModuleInfoResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetModuleInfo(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.RegisterService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil

}
func (service *ApiContainerGatewayServiceServer) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StartService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}
	serviceId := args.GetServiceId()
	servicePrivatePorts := args.PrivatePorts
	servicePrivatePortSpecs := getPortSpecsFromRpcPortBinding()
	service.connectionProvider.ForUserService(serviceId)

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetServices(ctx context.Context, args *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetServiceInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServiceInfoArgs) (*kurtosis_core_rpc_api_bindings.GetServiceInfoResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetServiceInfo(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.RemoveService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.Repartition(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ExecCommand(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.WaitForHttpGetEndpointAvailability(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.WaitForHttpPostEndpointAvailability(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetModules(ctx context.Context, args *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetModulesResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetModules(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) UploadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.UploadFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) StoreWebFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StoreWebFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StoreFilesArtifactFromService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

// Private methods
func getPortSpecsFromRpcPortBinding(rpcPortBindings map[string]*kurtosis_core_rpc_api_bindings.Port) map[string]*port_spec.PortSpec {
	portSpecs := map[string]*port_spec.PortSpec{}
	for portId, port := range rpcPortBindings {
		portSpec, err := getPortSpecFromRpcPortBinding(port)
		if err != nil {
			// handle error
		}
		portSpecs[portId] = portSpec
	}
	return portSpecs
}

func getPortSpecFromRpcPortBinding(port *kurtosis_core_rpc_api_bindings.Port) (*port_spec.PortSpec, error) {
	portProtocol := port.GetProtocol()

	portSpecNumber := uint16(port.GetNumber())
}
