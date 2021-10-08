/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/bulk_command_execution_engine"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/bulk_command_execution_engine/v0_bulk_command_execution"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/external_container_store"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/lambda_store"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/lambda_store/lambda_store_types"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_data_volume"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024
)

type ApiContainerService struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	enclaveDataVolume *enclave_data_volume.EnclaveDataVolume

	externalContainerStore *external_container_store.ExternalContainerStore

	serviceNetwork service_network.ServiceNetwork

	lambdaStore *lambda_store.LambdaStore

	bulkCmdExecEngine *bulk_command_execution_engine.BulkCommandExecutionEngine
}

func NewApiContainerService(
	enclaveDirectory *enclave_data_volume.EnclaveDataVolume,
	externalContainerStore *external_container_store.ExternalContainerStore,
	serviceNetwork service_network.ServiceNetwork,
	lambdaStore *lambda_store.LambdaStore,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		enclaveDataVolume: enclaveDirectory,
		externalContainerStore: externalContainerStore,
		serviceNetwork:    serviceNetwork,
		lambdaStore:       lambdaStore,
	}

	// NOTE: This creates a circular dependency between ApiContainerService <-> BulkCommandExecutionEngine, but out
	//  necessity: the API service must farm bulk commands out to the bulk command execution engine, which must call
	//  back to the API service to actually do work.
	v0BulkCmdProcessor, err := v0_bulk_command_execution.NewV0BulkCommandProcessor(serviceNetwork, service)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the v0 bulk command processor")
	}
	bulkCmdExecEngine := bulk_command_execution_engine.NewBulkCommandExecutionEngine(v0BulkCmdProcessor)
	service.bulkCmdExecEngine = bulkCmdExecEngine

	return service, nil
}

func (service ApiContainerService) StartExternalContainerRegistration(ctx context.Context, empty *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.StartExternalContainerRegistrationResponse, error) {
	registrationKey, ip, err := service.externalContainerStore.StartRegistration()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the registration with the external container store")
	}
	result := &kurtosis_core_rpc_api_bindings.StartExternalContainerRegistrationResponse{
		RegistrationKey: registrationKey,
		IpAddr:          ip.String(),
	}
	return result, nil
}

func (service ApiContainerService) FinishExternalContainerRegistration(ctx context.Context, args *kurtosis_core_rpc_api_bindings.FinishExternalContainerRegistrationArgs) (*emptypb.Empty, error) {
	registrationKey := args.RegistrationKey
	containerId := args.ContainerId
	if err := service.externalContainerStore.FinishRegistration(registrationKey, containerId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred finishing external container registration with registration key '%v' and container ID '%v'", registrationKey, containerId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) LoadLambda(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadLambdaArgs) (*emptypb.Empty, error) {
	lambdaId := lambda_store_types.LambdaID(args.LambdaId)
	image := args.ContainerImage
	serializedParams := args.SerializedParams
	if err := service.lambdaStore.LoadLambda(ctx, lambdaId, image, serializedParams); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading Lambda '%v' with container image '%v' and serialized params '%v'", lambdaId, image, serializedParams)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) UnloadLambda(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnloadLambdaArgs) (*emptypb.Empty, error) {
	lambdaId := lambda_store_types.LambdaID(args.LambdaId)
	if err := service.lambdaStore.UnloadLambda(ctx, lambdaId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unloading Lambda '%v' from the network", lambdaId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteLambda(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteLambdaArgs) (*kurtosis_core_rpc_api_bindings.ExecuteLambdaResponse, error) {
	lambdaId := lambda_store_types.LambdaID(args.LambdaId)
	serializedParams := args.SerializedParams
	serializedResult, err := service.lambdaStore.ExecuteLambda(ctx, lambdaId, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing Lambda '%v' with serialized params '%v'", lambdaId, serializedParams)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecuteLambdaResponse{SerializedResult: serializedResult}
	return resp, nil
}

func (service ApiContainerService) GetLambdaInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetLambdaInfoArgs) (*kurtosis_core_rpc_api_bindings.GetLambdaInfoResponse, error) {
	lambdaIdStr := args.LambdaId
	ipAddr, err := service.lambdaStore.GetLambdaIPAddrByID(lambda_store_types.LambdaID(lambdaIdStr))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IP address for Lambda '%v'", lambdaIdStr)
	}
	response := &kurtosis_core_rpc_api_bindings.GetLambdaInfoResponse{IpAddr: ipAddr.String()}
	return response, nil
}

func (service ApiContainerService) RegisterFilesArtifacts(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterFilesArtifactsArgs) (*emptypb.Empty, error) {
	filesArtifactCache, err := service.enclaveDataVolume.GetFilesArtifactCache()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the files artifact cache")
	}

	// TODO PERF: Do these in parallel
	logrus.Debug("Downloading files artifacts to the files artifact cache...")
	for artifactId, url := range args.FilesArtifactUrls {
		if err := filesArtifactCache.DownloadFilesArtifact(artifactId, url); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred downloading files artifact '%v' from URL '%v'", artifactId, url)
		}
	}
	logrus.Debug("Files artifacts downloaded successfully")

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	partitionId := service_network_types.PartitionID(args.PartitionId)

	ip, relativeServiceDirpath, err := service.serviceNetwork.RegisterService(serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return &kurtosis_core_rpc_api_bindings.RegisterServiceResponse{
		IpAddr:					ip.String(),
		RelativeServiceDirpath:	relativeServiceDirpath,
	}, nil
}

func (service ApiContainerService) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)

	usedPorts := map[nat.Port]bool{}
	portObjToPortSpecStr := map[nat.Port]string{}
	for portSpecStr := range args.UsedPorts {
		// NOTE: this function, frustratingly, doesn't return an error on failure - just emptystring
		protocol, portNumberStr := nat.SplitProtoPort(portSpecStr)
		if protocol == "" {
			return nil, stacktrace.NewError(
				"Could not split port specification string '%s' into protocol & number strings",
				portSpecStr)
		}
		portObj, err := nat.NewPort(protocol, portNumberStr)
		if err != nil {
			// TODO IP: Leaks internal information about the API container
			return nil, stacktrace.Propagate(
				err,
				"An error occurred constructing a port object out of protocol '%v' and port number string '%v'",
				protocol,
				portNumberStr)
		}
		usedPorts[portObj] = true
		portObjToPortSpecStr[portObj] = portSpecStr
	}

	serviceId := service_network_types.ServiceID(args.ServiceId)

	hostPortBindings, err := service.serviceNetwork.StartService(
		ctx,
		serviceId,
		args.DockerImage,
		usedPorts,
		args.EntrypointArgs,
		args.CmdArgs,
		args.DockerEnvVars,
		args.EnclaveDataVolMntDirpath,
		args.FilesArtifactMountDirpaths)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting the service in the service network")
	}

	// We strip out ports with nil host port bindings to make it easier to iterate over this map on the client side
	responseHostPortBindings := map[string]*kurtosis_core_rpc_api_bindings.PortBinding{}
	for portObj, hostPortBinding := range hostPortBindings {
		portSpecStr, found := portObjToPortSpecStr[portObj]
		if !found {
			return nil, stacktrace.NewError(
				"Found a port object, %+v, that doesn't correspond to a spec string as passed in via the args; this is very strange!",
				portObj,
			)
		}
		if hostPortBinding == nil {
			return nil, stacktrace.NewError(
				"Port spec string '%v' had a host port binding object returned by the Docker engine, but it was nil",
				portSpecStr,
			)
		}
		responseBinding := &kurtosis_core_rpc_api_bindings.PortBinding{
			InterfaceIp:   hostPortBinding.HostIP,
			InterfacePort: hostPortBinding.HostPort,
		}
		responseHostPortBindings[portSpecStr] = responseBinding
	}
	response := kurtosis_core_rpc_api_bindings.StartServiceResponse{
		UsedPortsHostPortBindings: responseHostPortBindings,
	}

	serviceStartLoglineSuffix := ""
	if len(responseHostPortBindings) > 0 {
		serviceStartLoglineSuffix = fmt.Sprintf(
			" with the following service-port-to-host-port bindings: %+v",
			responseHostPortBindings,
		)
	}
	logrus.Infof("Started service '%v'%v", serviceId, serviceStartLoglineSuffix)

	return &response, nil
}

func (service ApiContainerService) GetServiceInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServiceInfoArgs) (*kurtosis_core_rpc_api_bindings.GetServiceInfoResponse, error) {
	serviceIP, err := service.getServiceIPByServiceId(args.ServiceId)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred when trying to get the service IP address by service ID: '%v'",
			args.ServiceId)
	}

	serviceID := service_network_types.ServiceID(args.ServiceId)
	enclaveDataVolMntDirpath, err := service.serviceNetwork.GetServiceEnclaveDataVolMntDirpath(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred when trying to get service enclave data volume directory path by service ID: '%v'",
			serviceID)
	}
	relativeServiceDirpath, err := service.serviceNetwork.GetRelativeServiceDirpath(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred when trying to get relative service directory path by service ID: '%v'",
			serviceID)
	}

	serviceInfoResponse := &kurtosis_core_rpc_api_bindings.GetServiceInfoResponse{
		IpAddr:                        serviceIP.String(),
		EnclaveDataVolumeMountDirpath: enclaveDataVolMntDirpath,
		RelativeServiceDirpath: relativeServiceDirpath,
	}
	return serviceInfoResponse, nil
}

func (service ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)

	containerStopTimeoutSeconds := args.ContainerStopTimeoutSeconds
	containerStopTimeout := time.Duration(containerStopTimeoutSeconds) * time.Second

	if err := service.serviceNetwork.RemoveService(ctx, serviceId, containerStopTimeout); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := service_network_types.NewServiceIDSet()
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := service_network_types.ServiceID(serviceIdStr)
			serviceIdSet.AddElem(serviceId)
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := service_network_types.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection.ConnectionInfo {
			partitionBId := service_network_types.PartitionID(partitionBStr)
			partitionConnectionId := *service_network_types.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return nil, stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology.PartitionConnection{
				IsBlocked: connectionInfo.IsBlocked,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		IsBlocked: defaultConnectionInfo.IsBlocked,
	}

	if err := service.serviceNetwork.Repartition(
		ctx,
		partitionServices,
		partitionConnections,
		defaultConnection); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := service_network_types.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := service.serviceNetwork.ExecCommand(ctx, serviceId, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceId)
	}
	numLogOutputBytes := len(logOutput)
	if numLogOutputBytes > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError(
			"Log output from docker exec command '%+v' was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			numLogOutputBytes,
			maxLogOutputSizeBytes,
		)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode: exitCode,
		LogOutput: logOutput,
	}
	return resp, nil
}

func (service ApiContainerService) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := service.waitForEndpointAvailability(
			serviceIdStr,
			http.MethodGet,
			args.Port,
			args.Path,
			args.InitialDelayMilliseconds,
			args.Retries,
			args.RetriesDelayMilliseconds,
			"",
			args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := service.waitForEndpointAvailability(
		serviceIdStr,
		http.MethodPost,
		args.Port,
		args.Path,
		args.InitialDelayMilliseconds,
		args.Retries,
		args.RetriesDelayMilliseconds,
		args.RequestBody,
		args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteBulkCommands(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteBulkCommandsArgs) (*emptypb.Empty, error) {
	if err := service.bulkCmdExecEngine.Process(ctx, []byte(args.SerializedCommands)); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the bulk commands")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) GetServices(ctx context.Context, empty *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error){

	serviceIDs := make(map[string]bool, len(service.serviceNetwork.GetServiceIDs()))

	for serviceID, _ := range service.serviceNetwork.GetServiceIDs() {
		serviceIDStr := string(serviceID)
		if _, ok := serviceIDs[serviceIDStr]; !ok{
			serviceIDs[serviceIDStr] = true
		}
	}

	resp := &kurtosis_core_rpc_api_bindings.GetServicesResponse{
		ServiceIds: serviceIDs,
	}
	return resp, nil
}

func (service ApiContainerService) GetLambdas(ctx context.Context, empty *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetLambdasResponse, error){

	lambdaIDs := make(map[string]bool, len(service.lambdaStore.GetLambdas()))

	for lambdaID, _ := range service.lambdaStore.GetLambdas() {
		lambdaIDStr := string(lambdaID)
		if _, ok := lambdaIDs[lambdaIDStr]; !ok{
			lambdaIDs[lambdaIDStr] = true
		}
	}

	resp := &kurtosis_core_rpc_api_bindings.GetLambdasResponse{
		LambdaIds: lambdaIDs,
	}
	return resp, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func (service ApiContainerService) waitForEndpointAvailability(
		serviceIdStr string,
		httpMethod string,
		port uint32,
		path string,
		initialDelayMilliseconds uint32,
		retries uint32,
		retriesDelayMilliseconds uint32,
		requestBody string,
		bodyText string) error {

	var(
		resp *http.Response
		err error
	)

	serviceIP, err := service.getServiceIPByServiceId(serviceIdStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred when trying to get the IP address for service '%v'",
			serviceIdStr,
		)
	}

	url := fmt.Sprintf("http://%v:%v/%v", serviceIP, port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		resp, err = makeHttpRequest(httpMethod, url ,requestBody)
		if err == nil  {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	if bodyText != "" {
		body := resp.Body
		defer body.Close()

		bodyBytes, err := ioutil.ReadAll(body)

		if err != nil {
			return stacktrace.Propagate(err,
				"An error occurred reading the response body from endpoint '%v'", url)
		}

		bodyStr := string(bodyBytes)

		if bodyStr != bodyText {
			return stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", bodyText, url, bodyStr)
		}
	}

	return nil
}

func makeHttpRequest(httpMethod string, url string, body string) (*http.Response, error){
	var (
		resp *http.Response
		err error
	)

	if httpMethod == http.MethodPost {
		var bodyByte = []byte(body)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	} else if httpMethod == http.MethodGet{
		resp, err = http.Get(url)
	} else {
		return nil, stacktrace.NewError("HTTP method '%v' not allowed", httpMethod)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

func (service ApiContainerService) getServiceIPByServiceId(serviceId string) (net.IP, error){
	serviceID := service_network_types.ServiceID(serviceId)
	serviceIP, err := service.serviceNetwork.GetServiceIP(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"An error occurred when trying to get the service IP address by service ID: '%v'",
			serviceId)
	}
	return serviceIP, nil
}
