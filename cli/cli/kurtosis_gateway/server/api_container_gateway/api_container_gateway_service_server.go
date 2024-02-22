package api_container_gateway

import (
	"context"
	"io"
	"sync"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/connection"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_gateway/server/common"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	localHostIpStr                            = "127.0.0.1"
	errorCallingRemoteApiContainerFromGateway = "Expected to be able to call the remote api container from the gateway, instead a non nil err was returned"
)

type ApiContainerGatewayServiceServer struct {
	// ID of enclave the API container is running in
	enclaveId string
	// Client for the api container we'll be connecting too
	remoteApiContainerClient kurtosis_core_rpc_api_bindings.ApiContainerServiceClient

	// Provides connections to Kurtosis objectis in cluster
	connectionProvider *connection.GatewayConnectionProvider

	// ServiceMap and mutex to protect it
	mutex                               *sync.Mutex
	userServiceNameToLocalConnectionMap map[string]*runningLocalServiceConnection

	// User services port forwarding
	userServiceConnect kurtosis_core_rpc_api_bindings.Connect
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
		for _, runningLocalServiceConnection := range resultCoreGatewayServerService.userServiceNameToLocalConnectionMap {
			runningLocalServiceConnection.kurtosisConnection.Stop()
		}
	}

	return &ApiContainerGatewayServiceServer{
		remoteApiContainerClient:            remoteApiContainerClient,
		connectionProvider:                  connectionProvider,
		mutex:                               &sync.Mutex{},
		userServiceNameToLocalConnectionMap: userServiceToLocalConnectionMap,
		enclaveId:                           enclaveId,
		userServiceConnect:                  kurtosis_core_rpc_api_bindings.Connect_CONNECT,
	}, closeGatewayFunc
}

func (service *ApiContainerGatewayServiceServer) RunStarlarkScript(args *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, streamToWriteTo kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkScriptServer) error {
	logrus.Debug("Executing Starlark script")
	streamToReadFrom, err := service.remoteApiContainerClient.RunStarlarkScript(streamToWriteTo.Context(), args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the execution of Kurtosis code")
	}
	if err := common.ForwardKurtosisExecutionStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine](streamToReadFrom, streamToWriteTo); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from Kurtosis core back to the user")
	}
	return nil
}

func (service *ApiContainerGatewayServiceServer) ListFilesArtifactNamesAndUuids(ctx context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ListFilesArtifactNamesAndUuids(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) RunStarlarkPackage(args *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, streamToWriteTo kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkPackageServer) error {
	logrus.Debugf("Executing Starlark package '%s'", args.GetPackageId())
	streamToReadFrom, err := service.remoteApiContainerClient.RunStarlarkPackage(streamToWriteTo.Context(), args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the execution of Kurtosis code")
	}
	if err := common.ForwardKurtosisExecutionStream[kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine](streamToReadFrom, streamToWriteTo); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from Kurtosis core back to the user while executing package '%s'", args.GetPackageId())
	}
	return nil
}

func (service *ApiContainerGatewayServiceServer) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	if service.userServiceConnect == kurtosis_core_rpc_api_bindings.Connect_CONNECT {
		// Clean up the removed services when we have the full list of running services
		cleanupRemovedServices := len(args.ServiceIdentifiers) == 0

		if err := service.updateServicesLocalConnection(remoteApiContainerResponse.ServiceInfo, cleanupRemovedServices); err != nil {
			return nil, stacktrace.Propagate(err, "Error updating the services local connection")
		}
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) ConnectServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ConnectServicesArgs) (*kurtosis_core_rpc_api_bindings.ConnectServicesResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ConnectServices(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	service.userServiceConnect = args.Connect
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetExistingAndHistoricalServiceIdentifiers(ctx context.Context, args *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetExistingAndHistoricalServiceIdentifiers(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.ExecCommand(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.WaitForHttpGetEndpointAvailability(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.WaitForHttpPostEndpointAvailability(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) UploadFilesArtifact(server kurtosis_core_rpc_api_bindings.ApiContainerService_UploadFilesArtifactServer) error {
	client, err := service.remoteApiContainerClient.UploadFilesArtifact(server.Context())
	if err != nil {
		return stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	if err := forwardDataChunkStreamWithClose[*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse](server, client); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from UploadFilesArtifactV2 on gateway")
	}
	return nil
}

func (service *ApiContainerGatewayServiceServer) StoreWebFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StoreWebFilesArtifact(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}
func (service *ApiContainerGatewayServiceServer) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.StoreFilesArtifactFromService(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}

	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) InspectFilesArtifactContents(ctx context.Context, args *kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest) (*kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.InspectFilesArtifactContents(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) DownloadFilesArtifact(args *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs, server kurtosis_core_rpc_api_bindings.ApiContainerService_DownloadFilesArtifactServer) error {
	client, err := service.remoteApiContainerClient.DownloadFilesArtifact(server.Context(), args)
	if err != nil {
		return stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	if err := forwardDataChunkStream(client, server); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from DownloadFilesArtifactV2 on gateway")
	}
	return nil
}

func (service *ApiContainerGatewayServiceServer) UploadStarlarkPackage(server kurtosis_core_rpc_api_bindings.ApiContainerService_UploadStarlarkPackageServer) error {
	client, err := service.remoteApiContainerClient.UploadStarlarkPackage(server.Context())
	if err != nil {
		return stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	if err := forwardDataChunkStreamWithClose[*emptypb.Empty](server, client); err != nil {
		return stacktrace.Propagate(err, "Error forwarding stream from UploadStarlarkPackage on gateway")
	}
	return nil
}

func (service *ApiContainerGatewayServiceServer) GetStarlarkRun(ctx context.Context, args *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetStarlarkRun(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetStarlarkScriptPlanYaml(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StarlarkScriptPlanYamlArgs) (*kurtosis_core_rpc_api_bindings.PlanYaml, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetStarlarkScriptPlanYaml(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	return remoteApiContainerResponse, nil
}

func (service *ApiContainerGatewayServiceServer) GetStarlarkPackagePlanYaml(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StarlarkPackagePlanYamlArgs) (*kurtosis_core_rpc_api_bindings.PlanYaml, error) {
	remoteApiContainerResponse, err := service.remoteApiContainerClient.GetStarlarkPackagePlanYaml(ctx, args)
	if err != nil {
		return nil, stacktrace.Propagate(err, errorCallingRemoteApiContainerFromGateway)
	}
	return remoteApiContainerResponse, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================

// writeOverServiceInfoFieldsWithLocalConnectionInformationIfServiceRunning overwites the `MaybePublicPorts` and `MaybePublicIpAdrr` fields to connect to local ports forwarding requests to private ports in Kubernetes
// Only TCP Private Ports are forwarded
// Does nothing if the service is stopped (no pod running)
func (service *ApiContainerGatewayServiceServer) writeOverServiceInfoFieldsWithLocalConnectionInformationIfServiceRunning(serviceInfo *kurtosis_core_rpc_api_bindings.ServiceInfo) error {
	// If the service has no private ports, then don't overwrite any of the service info fields
	if len(serviceInfo.PrivatePorts) == 0 {
		return nil
	}

	serviceName := serviceInfo.GetName()
	var localConnErr error
	var runningLocalConnection *runningLocalServiceConnection
	cleanUpConnection := true
	runningLocalConnection, isFound := service.userServiceNameToLocalConnectionMap[serviceName]
	if !isFound {
		runningLocalConnection, localConnErr = service.startRunningConnectionForKurtosisServiceIfRunning(serviceName, serviceInfo.PrivatePorts)
		if localConnErr != nil {
			return stacktrace.Propagate(localConnErr, "Expected to be able to start a local connection to Kurtosis service '%v', instead a non-nil error was returned", serviceName)
		} else if runningLocalConnection == nil {
			return nil
		}
		defer func() {
			if cleanUpConnection {
				service.idempotentKillRunningConnectionForServiceName(serviceName)
			}
		}()
	}
	serviceInfo.MaybePublicPorts = runningLocalConnection.localPublicServicePorts
	serviceInfo.MaybePublicIpAddr = runningLocalConnection.localPublicIp

	cleanUpConnection = false
	return nil
}

// startRunningConnectionForKurtosisServiceIfRunning starts a port forwarding process from kernel assigned local ports to the remote service ports specified
// If privatePortsFromApi is empty, an error is thrown
func (service *ApiContainerGatewayServiceServer) startRunningConnectionForKurtosisServiceIfRunning(serviceName string, privatePortsFromApi map[string]*kurtosis_core_rpc_api_bindings.Port) (*runningLocalServiceConnection, error) {
	if len(privatePortsFromApi) == 0 {
		return nil, stacktrace.NewError("Expected Kurtosis service to have private ports specified for port forwarding, instead no ports were provided")
	}
	remotePrivatePortSpecs := map[string]*port_spec.PortSpec{}
	for portSpecId, coreApiPort := range privatePortsFromApi {
		if coreApiPort.GetTransportProtocol() != kurtosis_core_rpc_api_bindings.Port_TCP {
			logrus.Warnf(
				"Will not be able to forward service port with id '%v' for service with name '%v' in enclave '%v'. "+
					"The protocol of this port is '%v', but only '%v' is supported",
				portSpecId,
				serviceName,
				service.enclaveId,
				coreApiPort.GetTransportProtocol(),
				kurtosis_core_rpc_api_bindings.Port_TCP.String(),
			)
			continue
		}
		portNumberUint16 := uint16(coreApiPort.GetNumber())
		remotePortSpec, err := port_spec.NewPortSpec(portNumberUint16, port_spec.TransportProtocol_TCP, coreApiPort.GetMaybeApplicationProtocol(), nil)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to create port spec describing remote port '%v', instead a non-nil error was returned", portSpecId)
		}
		remotePrivatePortSpecs[portSpecId] = remotePortSpec
	}

	// Start listening
	serviceConnection, err := service.connectionProvider.ForUserServiceIfRunning(service.enclaveId, serviceName, remotePrivatePortSpecs)
	if err != nil {
		logrus.Errorf("Tried forwarding ports for user service '%v' in enclave '%v' but failed with error:\n%v", serviceName, service.enclaveId, err)
		return nil, nil
	} else if serviceConnection == nil {
		return nil, nil
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
			Number:                   uint32(localPortSpec.GetNumber()),
			TransportProtocol:        privateApiPort.GetTransportProtocol(),
			MaybeApplicationProtocol: privateApiPort.GetMaybeApplicationProtocol(),
			MaybeWaitTimeout:         privateApiPort.GetMaybeWaitTimeout(),
		}
	}

	runingLocalServiceConnection := &runningLocalServiceConnection{
		kurtosisConnection:      serviceConnection,
		localPublicServicePorts: localPublicApiPorts,
		localPublicIp:           localHostIpStr,
	}

	// Store information about our running gateway
	service.userServiceNameToLocalConnectionMap[serviceName] = runingLocalServiceConnection
	cleanUpMapEntry := true
	defer func() {
		if cleanUpMapEntry {
			delete(service.userServiceNameToLocalConnectionMap, serviceName)
		}
	}()

	cleanUpMapEntry = false
	cleanUpConnection = false
	return runingLocalServiceConnection, nil
}

func (service *ApiContainerGatewayServiceServer) idempotentKillRunningConnectionForServiceName(serviceName string) {
	runningLocalConnection, isRunning := service.userServiceNameToLocalConnectionMap[serviceName]
	// Nothing running, nothing to kill
	if !isRunning {
		return
	}

	// Close up the connection
	runningLocalConnection.kurtosisConnection.Stop()
	// delete the entry for the serve
	delete(service.userServiceNameToLocalConnectionMap, serviceName)
}

func (service *ApiContainerGatewayServiceServer) updateServicesLocalConnection(serviceInfos map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, cleanupRemovedServices bool) error {

	serviceNames := map[string]bool{}
	for _, serviceInfo := range serviceInfos {
		if err := service.writeOverServiceInfoFieldsWithLocalConnectionInformationIfServiceRunning(serviceInfo); err != nil {
			return stacktrace.Propagate(err, "Expected to be able to write over service info fields for service '%v', instead a non-nil error was returned", serviceInfo.Name)
		}
		serviceNames[serviceInfo.GetName()] = true
	}

	if cleanupRemovedServices {
		// Clean up connection for removed services
		for serviceName := range service.userServiceNameToLocalConnectionMap {
			if _, found := serviceNames[serviceName]; !found {
				service.idempotentKillRunningConnectionForServiceName(serviceName)
			}
		}
	}

	return nil
}

type dataChunkStreamReceiver interface {
	Recv() (*kurtosis_core_rpc_api_bindings.StreamedDataChunk, error)
}

type dataChunkStreamSenderCloserAndReceiver[T any] interface {
	dataChunkStreamSender
	CloseAndRecv() (T, error)
}

type dataChunkStreamSender interface {
	Send(*kurtosis_core_rpc_api_bindings.StreamedDataChunk) error
}

type dataChunkStreamReceiverSenderAndCloser[T any] interface {
	dataChunkStreamReceiver
	SendAndClose(T) error
}

func forwardDataChunkStream[T dataChunkStreamReceiver, U dataChunkStreamSender](streamToReadFrom T, streamToWriteTo U) error {
	for {
		dataChunk, readErr := streamToReadFrom.Recv()
		if readErr == io.EOF {
			logrus.Debug("Finished reading from the Kurtosis response line stream.")
			return nil
		}
		if readErr != nil {
			return stacktrace.Propagate(readErr, "Error reading Kurtosis execution lines from Kurtosis core stream")
		}
		if writeErr := streamToWriteTo.Send(dataChunk); writeErr != nil {
			return stacktrace.Propagate(readErr, "Received a Kurtosis execution line but failed forwarding it back to the user")
		}
	}
}

func forwardDataChunkStreamWithClose[T any, R dataChunkStreamReceiverSenderAndCloser[T], W dataChunkStreamSenderCloserAndReceiver[T]](streamToReadFrom R, streamToWriteTo W) error {
	err := forwardDataChunkStream(streamToReadFrom, streamToWriteTo)
	if err != nil {
		return err
	}
	uploadResponse, closeErr := streamToWriteTo.CloseAndRecv()
	if closeErr != nil {
		return stacktrace.Propagate(closeErr, "Error during Kurtosis closing upload client")
	}
	closeErr = streamToReadFrom.SendAndClose(uploadResponse)
	if closeErr != nil {
		return stacktrace.Propagate(closeErr, "Error during Kurtosis closing upload server")
	}
	return nil
}
