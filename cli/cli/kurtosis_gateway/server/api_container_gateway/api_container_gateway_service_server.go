package api_container_gateway

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
)

const (
	localHostIpStr = "127.0.0.1"
)

type ApiContainerGatewayServiceServer struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	// Id of enclave the API container is running in
	enclaveId string
	// Client for the api container we'll be connecting too
	remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	// Provides connections to Kurtosis objectis in cluster
	connectionProvider *connection.GatewayConnectionProvider

	// ServiceMap and mutex to protect it
	mutex                               *sync.Mutex
	userServiceGuidToLocalConnectionMap map[string]*runningLocalServiceConnection
}

type runningLocalServiceConnection struct {
	localPublicServicePorts map[string]*kurtosis_core_rpc_api_bindings.Port
	localPublicIp           string
	kurtosisConnection      connection.GatewayConnectionToKurtosis
}

func NewEnclaveApiContainerGatewayServer(connectionProvider *connection.GatewayConnectionProvider, remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient, enclaveId string) (resultCoreGatewayServerService *ApiContainerGatewayServiceServer, resultGatewayCloseFunc func()) {
	// Start out with 0 connections to user services
	userServiceToLocalConnectionMap := map[string]*runningLocalServiceConnection{}

	closeGatewayFunc := func() {
		// Stop any port forwarding
		for _, runningLocalServiceConnection := range resultCoreGatewayServerService.userServiceGuidToLocalConnectionMap {
			runningLocalServiceConnection.kurtosisConnection.Stop()
		}
	}

	return &ApiContainerGatewayServiceServer{
		remoteApiContainerClient:            remoteApiContainerClient,
		connectionProvider:                  connectionProvider,
		mutex:                               &sync.Mutex{},
		userServiceGuidToLocalConnectionMap: userServiceToLocalConnectionMap,
		enclaveId:                           enclaveId,
	}, closeGatewayFunc
}

func (service *ApiContainerGatewayServiceServer) LoadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadModuleArgs) (*kurtosis_core_rpc_api_bindings.LoadModuleResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.LoadModule(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) UnloadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnloadModuleArgs) (*kurtosis_core_rpc_api_bindings.UnloadModuleResponse, error) {
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
func (service *ApiContainerGatewayServiceServer) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.RegisterService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil

}
func (service *ApiContainerGatewayServiceServer) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	cleanUpServiceConnection := true
	cleanUpService := true
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StartService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}
	defer func() {
		if cleanUpService {
			destroyEnclaveArgs := &kurtosis_core_rpc_api_bindings.RemoveServiceArgs{ServiceId: args.GetServiceId()}
			if _, err := service.remoteApiContainerClient.RemoveService(ctx, destroyEnclaveArgs); err != nil {
				logrus.Errorf("Connecting to the service running in the remote cluster failed, expected to be able to cleanup the created service, but an error occurred calling the backend to remove the service we created:\v%v", err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the service with id '%v'", args.GetServiceId())
			}
		}
	}()
	// Do not connect to services with no ports specified
	if len(args.PrivatePorts) == 0 {
		return remoteApiContainerResponse, nil
	}
	serviceGuid := remoteApiContainerResponse.GetServiceGuid()

	runningServiceConnection, err := service.startRunningConnectionForKurtosisService(serviceGuid, args.PrivatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to start a local connection to service '%v', instead a non-nil error was returned", args.GetServiceId())
	}
	defer func() {
		if cleanUpServiceConnection {
			service.idempotentKillRunningConnectionForServiceGuid(serviceGuid)
		}
	}()

	// Overwrite PublicPorts and PublicIp fields
	remoteApiContainerResponse.PublicIpAddr = runningServiceConnection.localPublicIp
	remoteApiContainerResponse.PublicPorts = runningServiceConnection.localPublicServicePorts

	cleanUpService = false
	cleanUpServiceConnection = false
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}
	cleanUpConnection := true
	for serviceId, serviceInfo := range remoteApiContainerResponse.ServiceInfo {
		// Get the running connection if it's available, start one if there is no running connection
		serviceGuid := serviceInfo.GetServiceGuid()
		var runningLocalConnection *runningLocalServiceConnection
		runningLocalConnection, isFound := service.userServiceGuidToLocalConnectionMap[serviceGuid]
		if !isFound && len(serviceInfo.PrivatePorts) > 0 {
			runningLocalConnection, err = service.startRunningConnectionForKurtosisService(serviceGuid, serviceInfo.PrivatePorts)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Expected to be able to start a local connection to kurtosis service '%v', instead a non-nil error was returned", serviceId)
			}
			defer func() {
				if cleanUpConnection {
					service.idempotentKillRunningConnectionForServiceGuid(serviceGuid)
				}
			}()
		}
		// Overwrite PublicPorts and PublicIp fields
		serviceInfo.MaybePublicPorts = runningLocalConnection.localPublicServicePorts
		serviceInfo.MaybePublicIpAddr = runningLocalConnection.localPublicIp
	}

	cleanUpConnection = false
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*kurtosis_core_rpc_api_bindings.RemoveServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.RemoveService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	removedServiceGuid := remoteApiContainerResponse.GetServiceGuid()
	// Kill the connection if we can
	service.idempotentKillRunningConnectionForServiceGuid(removedServiceGuid)

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

func (service *ApiContainerGatewayServiceServer) GetModules(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetModulesArgs) (*kurtosis_core_rpc_api_bindings.GetModulesResponse, error) {
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

func (service *ApiContainerGatewayServiceServer) PauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.PauseServiceArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.PauseService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container method from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) UnpauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnpauseServiceArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.UnpauseService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

// startRunningConnectionForKurtosisService starts a port forwarding process from kernel assigned local ports to the remote service ports specified
// If privatePortsFromApi is empty, an error is thrown
func (service *ApiContainerGatewayServiceServer) startRunningConnectionForKurtosisService(serviceGuid string, privatePortsFromApi map[string]*kurtosis_core_rpc_api_bindings.Port) (*runningLocalServiceConnection, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if len(privatePortsFromApi) == 0 {
		return nil, stacktrace.NewError("Expected Kurtosis service to have private ports specified for port forwarding, instead no ports were provided")
	}
	remotePrivatePortSpecs := map[string]*port_spec.PortSpec{}
	for portSpecId, coreApiPort := range privatePortsFromApi {
		if coreApiPort.GetProtocol() != kurtosis_core_rpc_api_bindings.Port_TCP {
			logrus.Warnf(
				"Will not be able to forward service port with id '%v' for service with guid '%v' in enclave '%v'. "+
					"The protocol of this port is '%v', but only '%v' is supported",
				portSpecId,
				serviceGuid,
				service.enclaveId,
				coreApiPort.GetProtocol(),
				kurtosis_core_rpc_api_bindings.Port_TCP.String(),
			)
			continue
		}
		portNumberUint16 := uint16(coreApiPort.GetNumber())
		remotePortSpec, err := port_spec.NewPortSpec(portNumberUint16, port_spec.PortProtocol_TCP)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to create port spec describing remote port '%v', instead a non-nil error was returned", portSpecId)
		}
		remotePrivatePortSpecs[portSpecId] = remotePortSpec
	}

	// Start listening
	serviceConnection, err := service.connectionProvider.ForUserService(service.enclaveId, serviceGuid, remotePrivatePortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to start a local connection service with guid '%v' in enclave '%v', instead a non-nil error was returned", serviceGuid, service.enclaveId)
	}
	cleanUpConnection := true
	defer func() {
		if cleanUpConnection {
			serviceConnection.Stop()
		}
	}()

	localPublicApiPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, privateApiPort := range privatePortsFromApi {
		localPortSpec, found := serviceConnection.GetLocalPorts()[portId]
		if !found {
			return nil, stacktrace.NewError("Service requested private port '%v', but no local forwarded port was made for it", portId)
		}
		localPublicApiPorts[portId] = &kurtosis_core_rpc_api_bindings.Port{
			Number:   uint32(localPortSpec.GetNumber()),
			Protocol: privateApiPort.Protocol,
		}
	}

	runingLocalServiceConnection := &runningLocalServiceConnection{
		kurtosisConnection:      serviceConnection,
		localPublicServicePorts: localPublicApiPorts,
		localPublicIp:           localHostIpStr,
	}

	// Store information about our running gateway
	service.userServiceGuidToLocalConnectionMap[serviceGuid] = runingLocalServiceConnection
	cleanUpMapEntry := true
	defer func() {
		if cleanUpMapEntry {
			delete(service.userServiceGuidToLocalConnectionMap, serviceGuid)
		}
	}()

	cleanUpMapEntry = false
	cleanUpConnection = false
	return runingLocalServiceConnection, nil
}

func (service *ApiContainerGatewayServiceServer) idempotentKillRunningConnectionForServiceGuid(serviceGuid string) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	runningLocalConnection, isRunning := service.userServiceGuidToLocalConnectionMap[serviceGuid]
	// Nothing running, nothing to kill
	if !isRunning {
		return
	}

	// Close up the connection
	runningLocalConnection.kurtosisConnection.Stop()
	// delete the entry for the serve
	delete(service.userServiceGuidToLocalConnectionMap, serviceGuid)
}
