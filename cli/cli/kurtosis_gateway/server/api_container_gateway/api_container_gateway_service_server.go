package api_container_gateway

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
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

func (service *ApiContainerGatewayServiceServer) ExecuteStartosisScript(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteStartosisScriptArgs) (*kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ExecuteStartosisScript(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) ExecuteStartosisModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteStartosisModuleArgs) (*kurtosis_core_rpc_api_bindings.ExecuteStartosisResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ExecuteStartosisModule(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) StartServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServicesArgs) (*kurtosis_core_rpc_api_bindings.StartServicesResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	failedServicesPool := map[string]string{}
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StartServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}
	shouldRemoveServices := map[string]bool{}
	for serviceID := range remoteApiContainerResponse.GetSuccessfulServiceIdsToServiceInfo() {
		shouldRemoveServices[serviceID] = true
	}
	defer func() {
		for serviceIDStr := range shouldRemoveServices {
			removeServiceArgs := &kurtosis_core_rpc_api_bindings.RemoveServiceArgs{ServiceId: serviceIDStr}
			if _, err := service.remoteApiContainerClient.RemoveService(ctx, removeServiceArgs); err != nil {
				err = stacktrace.Propagate(err,
					"Connecting to the service running in the remote cluster failed, expected to be able to cleanup the created service, but an error occurred calling the backend to remove the service we created. "+
						"ACTION REQUIRED: You'll need to manually remove the service with id '%v'", serviceIDStr)
				failedServicesPool[serviceIDStr] = err.Error()
			}
		}
	}()
	successfulServices := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	// Add failed services to failed services pool
	for serviceIDStr, errStr := range remoteApiContainerResponse.GetFailedServiceIdsToError() {
		failedServicesPool[serviceIDStr] = errStr
	}

	// Write over the PublicIp and Public Ports fields so the service can be accessed through local port forwarding
	for serviceIDStr, serviceInfo := range remoteApiContainerResponse.GetSuccessfulServiceIdsToServiceInfo() {
		if err := service.writeOverServiceInfoFieldsWithLocalConnectionInformation(serviceInfo); err != nil {
			err = stacktrace.Propagate(err, "Expected to be able to write over service info fields for service '%v', instead a non-nil error was returned", serviceIDStr)
			failedServicesPool[serviceIDStr] = err.Error()
		}
		successfulServices[serviceIDStr] = serviceInfo
	}

	// Do not remove successful services
	for serviceIDStr := range successfulServices {
		delete(shouldRemoveServices, serviceIDStr)
	}
	startServicesResp := binding_constructors.NewStartServicesResponse(successfulServices, failedServicesPool)
	return startServicesResp, nil
}

func (service *ApiContainerGatewayServiceServer) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}
	for serviceId, serviceInfo := range remoteApiContainerResponse.ServiceInfo {
		if err := service.writeOverServiceInfoFieldsWithLocalConnectionInformation(serviceInfo); err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to write over service info fields for service '%v', instead a non-nil error was returned", serviceId)
		}
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*kurtosis_core_rpc_api_bindings.RemoveServiceResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
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

func (service *ApiContainerGatewayServiceServer) DefineFact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.DefineFactArgs) (*kurtosis_core_rpc_api_bindings.DefineFactResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.DefineFact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetFactValues(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetFactValuesArgs) (*kurtosis_core_rpc_api_bindings.GetFactValuesResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetFactValues(ctx, args)
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

func (service *ApiContainerGatewayServiceServer) RenderTemplatesToFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.RenderTemplatesToFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned")
	}

	return remoteApiContainerResponse, nil
}

// writeOverServiceInfoFieldsWithLocalConnectionInformation overwites the `MaybePublicPorts` and `MaybePublicIpAdrr` fields to connect to local ports forwarding requests to private ports in Kubernetes
// Only TCP Private Ports are forwarded
func (service *ApiContainerGatewayServiceServer) writeOverServiceInfoFieldsWithLocalConnectionInformation(serviceInfo *kurtosis_core_rpc_api_bindings.ServiceInfo) error {
	// If the service has no private ports, then don't overwrite any of the service info fields
	if len(serviceInfo.PrivatePorts) == 0 {
		return nil
	}

	serviceGuid := serviceInfo.GetServiceGuid()
	var localConnErr error
	var runningLocalConnection *runningLocalServiceConnection
	cleanUpConnection := true
	runningLocalConnection, isFound := service.userServiceGuidToLocalConnectionMap[serviceGuid]
	if !isFound {
		runningLocalConnection, localConnErr = service.startRunningConnectionForKurtosisService(serviceGuid, serviceInfo.PrivatePorts)
		if localConnErr != nil {
			return stacktrace.Propagate(localConnErr, "Expected to be able to start a local connection to Kurtosis service '%v', instead a non-nil error was returned", serviceGuid)
		}
		defer func() {
			if cleanUpConnection {
				service.idempotentKillRunningConnectionForServiceGuid(serviceGuid)
			}
		}()
	}
	serviceInfo.MaybePublicPorts = runningLocalConnection.localPublicServicePorts
	serviceInfo.MaybePublicIpAddr = runningLocalConnection.localPublicIp

	cleanUpConnection = false
	return nil
}

// startRunningConnectionForKurtosisService starts a port forwarding process from kernel assigned local ports to the remote service ports specified
// If privatePortsFromApi is empty, an error is thrown
func (service *ApiContainerGatewayServiceServer) startRunningConnectionForKurtosisService(serviceGuid string, privatePortsFromApi map[string]*kurtosis_core_rpc_api_bindings.Port) (*runningLocalServiceConnection, error) {
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
		// Skip the private remote port if no public local port is forwarding to it
		if !found {
			continue
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
