/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024

	// The string returned by the API if a service's public IP address doesn't exist
	missingPublicIpAddrStr = ""

	defaultStartosisDryRun = false

	// Overwrite existing module with new module, this allows user to iterate on an enclave with a
	// given module
	doOverwriteExistingModule = true

	isScript    = true
	isNotScript = false
	isNotRemote = false
)

// Guaranteed (by a unit test) to be a 1:1 mapping between API port protos and port spec protos
var apiContainerPortProtoToPortSpecPortProto = map[kurtosis_core_rpc_api_bindings.Port_TransportProtocol]port_spec.TransportProtocol{
	kurtosis_core_rpc_api_bindings.Port_TCP:  port_spec.TransportProtocol_TCP,
	kurtosis_core_rpc_api_bindings.Port_SCTP: port_spec.TransportProtocol_SCTP,
	kurtosis_core_rpc_api_bindings.Port_UDP:  port_spec.TransportProtocol_UDP,
}

type ApiContainerService struct {
	filesArtifactStore *enclave_data_directory.FilesArtifactStore

	serviceNetwork service_network.ServiceNetwork

	startosisRunner *startosis_engine.StartosisRunner

	metricsClient client.MetricsClient

	startosisModuleContentProvider startosis_packages.PackageContentProvider
}

func NewApiContainerService(
	filesArtifactStore *enclave_data_directory.FilesArtifactStore,
	serviceNetwork service_network.ServiceNetwork,
	startosisRunner *startosis_engine.StartosisRunner,
	metricsClient client.MetricsClient,
	startosisModuleContentProvider startosis_packages.PackageContentProvider,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		filesArtifactStore:             filesArtifactStore,
		serviceNetwork:                 serviceNetwork,
		startosisRunner:                startosisRunner,
		metricsClient:                  metricsClient,
		startosisModuleContentProvider: startosisModuleContentProvider,
	}

	return service, nil
}

func (apicService ApiContainerService) RunStarlarkScript(args *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkScriptServer) error {
	serializedStarlarkScript := args.GetSerializedScript()
	serializedParams := args.GetSerializedParams()
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)

	if err := apicService.metricsClient.TrackKurtosisRun(startosis_constants.PackageIdPlaceholderForStandaloneScript, isNotRemote, dryRun, isScript); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking kurtosis run event\n%v", err)
	}

	apicService.runStarlark(dryRun, startosis_constants.PackageIdPlaceholderForStandaloneScript, serializedStarlarkScript, serializedParams, stream)
	return nil
}

func (apicService ApiContainerService) RunStarlarkPackage(args *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkPackageServer) error {
	packageId := args.GetPackageId()
	isRemote := args.GetRemote()
	moduleContentIfLocal := args.GetLocal()
	serializedParams := args.SerializedParams
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)

	if err := apicService.metricsClient.TrackKurtosisRun(packageId, isRemote, dryRun, isNotScript); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking kurtosis run event\n%v", err)
	}

	scriptWithRunFunction, interpretationError := apicService.runStarlarkPackageSetup(packageId, isRemote, moduleContentIfLocal)
	if interpretationError != nil {
		if err := stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError.ToAPIType())); err != nil {
			return stacktrace.Propagate(err, "Error preparing for package execution and this error could not be sent through the output stream: '%s'", packageId)
		}
		return nil
	}
	apicService.runStarlark(dryRun, packageId, scriptWithRunFunction, serializedParams, stream)
	return nil
}

func (apicService ApiContainerService) StartServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServicesArgs) (*kurtosis_core_rpc_api_bindings.StartServicesResponse, error) {
	failedServicesPool := map[kurtosis_backend_service.ServiceID]error{}
	serviceIDsToAPIConfigs := map[kurtosis_backend_service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig{}

	for serviceIDStr, apiServiceConfig := range args.ServiceIdsToConfigs {
		logrus.Debugf("Received request to start service with the following args: %+v", apiServiceConfig)
		serviceIDsToAPIConfigs[kurtosis_backend_service.ServiceID(serviceIDStr)] = apiServiceConfig
	}

	successfulServices, failedServices, err := apicService.serviceNetwork.StartServices(ctx, serviceIDsToAPIConfigs)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting services in the service network")
	}
	// TODO We SHOULD defer an undo to undo the service-starting resource that we did here, but we don't have a way to just undo
	// that and leave the registration intact (since we only have RemoveService as of 2022-08-11, but that also deletes the registration,
	// which would mean deleting a resource we don't own here)

	for serviceID, serviceErr := range failedServices {
		failedServicesPool[serviceID] = serviceErr
		logrus.Debugf("Failed to start service '%v'", serviceID)
	}

	serviceIDsToServiceInfo := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	for serviceID, startedService := range successfulServices {
		serviceRegistration := startedService.GetRegistration()
		serviceGuidStr := string(serviceRegistration.GetGUID())
		privateServiceIpStr := serviceRegistration.GetPrivateIP().String()
		privateServicePortSpecs := startedService.GetPrivatePorts()
		privateApiPorts, err := transformPortSpecMapToApiPortsMap(privateServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' private port specs to API ports", serviceID)
			continue
		}
		publicServicePortSpecs := startedService.GetMaybePublicPorts()
		publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceID] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' public port specs to API ports.", serviceID)
			continue
		}
		maybePublicIpAddr := startedService.GetMaybePublicIP()
		publicIpAddrStr := missingPublicIpAddrStr
		if maybePublicIpAddr != nil {
			publicIpAddrStr = maybePublicIpAddr.String()
		}

		serviceIDsToServiceInfo[string(serviceID)] = binding_constructors.NewServiceInfo(serviceGuidStr, privateServiceIpStr, privateApiPorts, publicIpAddrStr, publicApiPorts)
		serviceStartLoglineSuffix := ""
		if len(publicServicePortSpecs) > 0 {
			serviceStartLoglineSuffix = fmt.Sprintf(
				" with the following public ports: %+v",
				publicServicePortSpecs,
			)
		}
		logrus.Infof("Started service '%v'%v", serviceID, serviceStartLoglineSuffix)
	}

	failedServiceIDsToErrorStr := map[string]string{}
	for id, serviceErr := range failedServicesPool {
		failedServiceIDsToErrorStr[string(id)] = serviceErr.Error()
	}

	return binding_constructors.NewStartServicesResponse(serviceIDsToServiceInfo, failedServiceIDsToErrorStr), nil
}

func (apicService ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*kurtosis_core_rpc_api_bindings.RemoveServiceResponse, error) {
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)

	serviceGuid, err := apicService.serviceNetwork.RemoveService(ctx, serviceId)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return binding_constructors.NewRemoveServiceResponse(string(serviceGuid)), nil
}

func (apicService ApiContainerService) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]map[kurtosis_backend_service.ServiceID]bool{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := map[kurtosis_backend_service.ServiceID]bool{}
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
			serviceIdSet[serviceId] = true
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
			partitionConnection := partition_topology.NewPartitionConnection(connectionInfo.PacketLossPercentage)
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.NewPartitionConnection(defaultConnectionInfo.PacketLossPercentage)

	if err := apicService.serviceNetwork.Repartition(
		ctx,
		partitionServices,
		partitionConnections,
		defaultConnection); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) PauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.PauseServiceArgs) (*emptypb.Empty, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	err := service.serviceNetwork.PauseService(ctx, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to pause service '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) UnpauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnpauseServiceArgs) (*emptypb.Empty, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	err := service.serviceNetwork.UnpauseService(ctx, serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (apicService ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := apicService.serviceNetwork.ExecCommand(ctx, serviceId, command)
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
		ExitCode:  exitCode,
		LogOutput: logOutput,
	}
	return resp, nil
}

func (apicService ApiContainerService) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := apicService.waitForEndpointAvailability(
		ctx,
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

func (apicService ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := apicService.waitForEndpointAvailability(
		ctx,
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

func (apicService ApiContainerService) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	serviceInfos := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	filterServiceIds := args.ServiceIds

	for serviceID := range apicService.serviceNetwork.GetServiceIDs() {
		serviceIDStr := string(serviceID)
		if filterServiceIds != nil && len(filterServiceIds) > 0 {
			if _, found := filterServiceIds[serviceIDStr]; !found {
				continue
			}
		}
		serviceInfo, err := apicService.getServiceInfo(ctx, serviceID)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service info for service '%v'", serviceID)
		}
		serviceInfos[serviceIDStr] = serviceInfo
	}

	resp := binding_constructors.NewGetServicesResponse(serviceInfos)
	return resp, nil
}

func (apicService ApiContainerService) UploadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {

	filesArtifactUuid, err := apicService.serviceNetwork.UploadFilesArtifact(args.Data)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to upload the file")
	}

	response := &kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse{Uuid: string(filesArtifactUuid)}
	return response, nil
}

func (apicService ApiContainerService) DownloadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse, error) {
	filesArtifactIdStr := args.Id
	if strings.TrimSpace(filesArtifactIdStr) == "" {
		return nil, stacktrace.NewError("Cannot download file with empty files artifact UUID")
	}
	filesArtifactId := enclave_data_directory.FilesArtifactID(filesArtifactIdStr)

	artifactFile, err := apicService.filesArtifactStore.GetFile(filesArtifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact '%v'", filesArtifactId)
	}

	fileBytes, err := ioutil.ReadFile(artifactFile.GetAbsoluteFilepath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading files artifact file bytes")
	}

	resp := &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse{Data: fileBytes}
	return resp, nil
}

func (apicService ApiContainerService) StoreWebFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	url := args.Url

	resp, err := http.Get(args.Url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the files artifact bytes", url)
	}
	defer resp.Body.Close()
	body := bufio.NewReader(resp.Body)

	filesArtifactUuId, err := apicService.filesArtifactStore.StoreFile(body)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred storing the file from URL '%v' in the files artifact store", url)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse{Uuid: string(filesArtifactUuId)}
	return response, nil
}

func (apicService ApiContainerService) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	srcPath := args.SourcePath

	filesArtifactId, err := apicService.serviceNetwork.CopyFilesFromService(ctx, serviceId, srcPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying source '%v' from service with ID '%v'", srcPath, serviceId)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse{Uuid: string(filesArtifactId)}
	return response, nil
}

func (apicService ApiContainerService) RenderTemplatesToFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse, error) {
	templatesAndDataByDestinationRelFilepath := args.TemplatesAndDataByDestinationRelFilepath
	filesArtifactUuid, err := apicService.serviceNetwork.RenderTemplates(templatesAndDataByDestinationRelFilepath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while rendering templates to files artifact")
	}
	response := binding_constructors.NewRenderTemplatesToFilesArtifactResponse(string(filesArtifactUuid))
	return response, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func transformPortSpecToApiPort(port *port_spec.PortSpec) (*kurtosis_core_rpc_api_bindings.Port, error) {
	portNumUint16 := port.GetNumber()
	portSpecProto := port.GetTransportProtocol()

	// Yes, this isn't the most efficient way to do this, but the map is tiny so it doesn't matter
	var apiProto kurtosis_core_rpc_api_bindings.Port_TransportProtocol
	foundApiProto := false
	for mappedApiProto, mappedPortSpecProto := range apiContainerPortProtoToPortSpecPortProto {
		if portSpecProto == mappedPortSpecProto {
			apiProto = mappedApiProto
			foundApiProto = true
			break
		}
	}
	if !foundApiProto {
		return nil, stacktrace.NewError("Couldn't find an API port proto for port spec port proto '%v'; this should never happen, and is a bug in Kurtosis!", portSpecProto)
	}

	maybeApplicationProtocol := ""
	if port.GetMaybeApplicationProtocol() != nil {
		maybeApplicationProtocol = *port.GetMaybeApplicationProtocol()
	}

	result := binding_constructors.NewPort(uint32(portNumUint16), apiProto, maybeApplicationProtocol)
	return result, nil
}

func transformPortSpecMapToApiPortsMap(apiPorts map[string]*port_spec.PortSpec) (map[string]*kurtosis_core_rpc_api_bindings.Port, error) {
	result := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, portSpec := range apiPorts {
		publicApiPort, err := transformPortSpecToApiPort(portSpec)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming port spec for port '%v' into an API port", portId)
		}
		result[portId] = publicApiPort
	}
	return result, nil
}

func (apicService ApiContainerService) waitForEndpointAvailability(
	ctx context.Context,
	serviceIdStr string,
	httpMethod string,
	port uint32,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
	requestBody string,
	bodyText string) error {

	var (
		resp *http.Response
		err  error
	)

	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		kurtosis_backend_service.ServiceID(serviceIdStr),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service '%v'", serviceIdStr)
	}
	if serviceObj.GetStatus() != container_status.ContainerStatus_Running {
		return stacktrace.NewError("Service '%v' isn't running so can never become available", serviceIdStr)
	}
	privateIp := serviceObj.GetRegistration().GetPrivateIP()

	url := fmt.Sprintf("http://%v:%v/%v", privateIp.String(), port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		resp, err = makeHttpRequest(httpMethod, url, requestBody)
		if err == nil {
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

func makeHttpRequest(httpMethod string, url string, body string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	if httpMethod == http.MethodPost {
		var bodyByte = []byte(body)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	} else if httpMethod == http.MethodGet {
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

func (apicService ApiContainerService) getServiceInfo(ctx context.Context, serviceId kurtosis_backend_service.ServiceID) (*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		serviceId,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for service '%v'", serviceId)
	}
	privatePorts := serviceObj.GetPrivatePorts()
	privateIp := serviceObj.GetRegistration().GetPrivateIP()
	maybePublicIp := serviceObj.GetMaybePublicIP()
	maybePublicPorts := serviceObj.GetMaybePublicPorts()
	serviceGuidStr := string(serviceObj.GetRegistration().GetGUID())

	privateApiPorts, err := transformPortSpecMapToApiPortsMap(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's private port specs to API ports")
	}
	publicIpAddrStr := missingPublicIpAddrStr
	if maybePublicIp != nil {
		publicIpAddrStr = maybePublicIp.String()
	}
	publicApiPorts := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	if maybePublicPorts != nil {
		publicApiPorts, err = transformPortSpecMapToApiPortsMap(maybePublicPorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred transforming the service's public port spec ports to API ports")
		}
	}

	serviceInfoResponse := binding_constructors.NewServiceInfo(
		serviceGuidStr,
		privateIp.String(),
		privateApiPorts,
		publicIpAddrStr,
		publicApiPorts,
	)
	return serviceInfoResponse, nil
}

func (apicService ApiContainerService) runStarlarkPackageSetup(packageId string, isRemote bool, moduleContentIfLocal []byte) (string, *startosis_errors.InterpretationError) {
	var packageRootPathOnDisk string
	var interpretationError *startosis_errors.InterpretationError
	if isRemote {
		packageRootPathOnDisk, interpretationError = apicService.startosisModuleContentProvider.ClonePackage(packageId)
	} else {
		packageRootPathOnDisk, interpretationError = apicService.startosisModuleContentProvider.StorePackageContents(packageId, moduleContentIfLocal, doOverwriteExistingModule)
	}
	if interpretationError != nil {
		return "", interpretationError
	}

	pathToMainFile := path.Join(packageRootPathOnDisk, startosis_constants.MainFileName)
	if _, err := os.Stat(pathToMainFile); err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while verifying that '%v' exists on root of package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	mainScriptToExecute, err := os.ReadFile(pathToMainFile)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while reading '%v' at the root of package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	return string(mainScriptToExecute), nil
}

func (apicService ApiContainerService) runStarlark(dryRun bool, packageId string, serializedStarlark string, serializedParams string, stream grpc.ServerStream) {
	responseLineStream := apicService.startosisRunner.Run(stream.Context(), dryRun, packageId, serializedStarlark, serializedParams)
	for {
		select {
		case <-stream.Context().Done():
			// TODO: maybe add the ability to kill the execution
			logrus.Infof("Stream was closed by client. The script ouput won't be returned anymore but note that the execution won't be interrupted. There's currently no way to stop a Kurtosis script execution.")
			return
		case responseLine, isChanOpen := <-responseLineStream:
			if !isChanOpen {
				// Channel closed means that this function returned, so we won't receive any message through the stream anymore
				// We expect the stream to be closed soon and the above case to exit that function
				logrus.Info("Startosis script execution returned, no more output to stream.")
				return
			}
			// in addition to send the msg to the RPC stream, we also print the lines to the APIC logs at debug level
			logrus.Debugf("Received response line from Starlark runner: '%v'", responseLine)
			if err := stream.SendMsg(responseLine); err != nil {
				logrus.Errorf("Starlark response line sent through the channel but could not be forwarded to API Container client. Some log lines will not be returned to the user.\nResponse line was: \n%v. Error was: \n%v", responseLine, err.Error())
			}
		}
	}
}
