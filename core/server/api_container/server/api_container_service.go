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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	grpc_file_transfer "github.com/kurtosis-tech/kurtosis/grpc-file-transfer"
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

	defaultParallelism = 4
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

	startosisModuleContentProvider startosis_packages.PackageContentProvider
}

func NewApiContainerService(
	filesArtifactStore *enclave_data_directory.FilesArtifactStore,
	serviceNetwork service_network.ServiceNetwork,
	startosisRunner *startosis_engine.StartosisRunner,
	startosisModuleContentProvider startosis_packages.PackageContentProvider,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		filesArtifactStore:             filesArtifactStore,
		serviceNetwork:                 serviceNetwork,
		startosisRunner:                startosisRunner,
		startosisModuleContentProvider: startosisModuleContentProvider,
	}

	return service, nil
}

func (apicService ApiContainerService) RunStarlarkScript(args *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkScriptServer) error {
	serializedStarlarkScript := args.GetSerializedScript()
	serializedParams := args.GetSerializedParams()
	parallelism := int(args.GetParallelism())
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)

	apicService.runStarlark(parallelism, dryRun, startosis_constants.PackageIdPlaceholderForStandaloneScript, serializedStarlarkScript, serializedParams, stream)
	return nil
}

func (apicService ApiContainerService) RunStarlarkPackage(args *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkPackageServer) error {
	packageId := args.GetPackageId()
	isRemote := args.GetRemote()
	moduleContentIfLocal := args.GetLocal()
	parallelism := int(args.GetParallelism())
	serializedParams := args.SerializedParams
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)

	scriptWithRunFunction, interpretationError := apicService.runStarlarkPackageSetup(packageId, isRemote, moduleContentIfLocal)
	if interpretationError != nil {
		if err := stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError.ToAPIType())); err != nil {
			return stacktrace.Propagate(err, "Error preparing for package execution and this error could not be sent through the output stream: '%s'", packageId)
		}
		return nil
	}
	apicService.runStarlark(parallelism, dryRun, packageId, scriptWithRunFunction, serializedParams, stream)
	return nil
}

func (apicService ApiContainerService) StartServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServicesArgs) (*kurtosis_core_rpc_api_bindings.StartServicesResponse, error) {
	failedServicesPool := map[kurtosis_backend_service.ServiceName]error{}
	serviceNamesToAPIConfigs := map[kurtosis_backend_service.ServiceName]*kurtosis_core_rpc_api_bindings.ServiceConfig{}

	for serviceNameStr, apiServiceConfig := range args.ServiceNamesToConfigs {
		logrus.Debugf("Received request to start service with the following args: %+v", apiServiceConfig)
		serviceNamesToAPIConfigs[kurtosis_backend_service.ServiceName(serviceNameStr)] = apiServiceConfig
	}

	successfulServices, failedServices, err := apicService.serviceNetwork.StartServices(ctx, serviceNamesToAPIConfigs, defaultParallelism)
	if err != nil {
		return nil, stacktrace.Propagate(err, "None of the services '%v' mentioned in the request were able to start due to an unexpected error", serviceNamesToAPIConfigs)
	}
	// TODO We SHOULD defer an undo to undo the service-starting resource that we did here, but we don't have a way to just undo
	// that and leave the registration intact (since we only have RemoveService as of 2022-08-11, but that also deletes the registration,
	// which would mean deleting a resource we don't own here)

	for serviceName, serviceErr := range failedServices {
		failedServicesPool[serviceName] = serviceErr
		logrus.Debugf("Failed to start service '%v'", serviceName)
	}

	serviceNamesToServiceInfo := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	for serviceName, startedService := range successfulServices {
		serviceRegistration := startedService.GetRegistration()
		serviceUuidStr := string(serviceRegistration.GetUUID())
		privateServiceIpStr := serviceRegistration.GetPrivateIP().String()
		privateServicePortSpecs := startedService.GetPrivatePorts()
		privateApiPorts, err := transformPortSpecMapToApiPortsMap(privateServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' private port specs to API ports", serviceName)
			continue
		}
		publicServicePortSpecs := startedService.GetMaybePublicPorts()
		publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePortSpecs)
		if err != nil {
			failedServicesPool[serviceName] = stacktrace.Propagate(err, "An error occurred transforming the service '%v' public port specs to API ports.", serviceName)
			continue
		}
		maybePublicIpAddr := startedService.GetMaybePublicIP()
		publicIpAddrStr := missingPublicIpAddrStr
		if maybePublicIpAddr != nil {
			publicIpAddrStr = maybePublicIpAddr.String()
		}

		serviceNamesToServiceInfo[string(serviceName)] = binding_constructors.NewServiceInfo(serviceUuidStr, string(serviceName), uuid_generator.ShortenedUUIDString(serviceUuidStr), privateServiceIpStr, privateApiPorts, publicIpAddrStr, publicApiPorts)
		serviceStartLoglineSuffix := ""
		if len(publicServicePortSpecs) > 0 {
			serviceStartLoglineSuffix = fmt.Sprintf(
				" with the following public ports: %+v",
				publicServicePortSpecs,
			)
		}
		logrus.Infof("Started service '%v'%v", serviceName, serviceStartLoglineSuffix)
	}

	failedServiceNamesToErrorStr := map[string]string{}
	for id, serviceErr := range failedServicesPool {
		failedServiceNamesToErrorStr[string(id)] = serviceErr.Error()
	}

	return binding_constructors.NewStartServicesResponse(serviceNamesToServiceInfo, failedServiceNamesToErrorStr), nil
}

func (apicService ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*kurtosis_core_rpc_api_bindings.RemoveServiceResponse, error) {
	serviceIdentifier := args.ServiceIdentifier

	serviceUuid, err := apicService.serviceNetwork.RemoveService(ctx, serviceIdentifier)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with identifier '%v'", serviceIdentifier)
	}
	return binding_constructors.NewRemoveServiceResponse(string(serviceUuid)), nil
}

func (apicService ApiContainerService) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]map[kurtosis_backend_service.ServiceName]bool{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := map[kurtosis_backend_service.ServiceName]bool{}
		for serviceNameStr := range servicesInPartition.ServiceNameSet {
			serviceId := kurtosis_backend_service.ServiceName(serviceNameStr)
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

			//TODO: We will be removing Repartition method completely from sdks (golang and typescript) so not needed in PartitionInfo
			partitionConnection := partition_topology.NewPartitionConnection(partition_topology.NewPacketLoss(connectionInfo.PacketLossPercentage), partition_topology.ConnectionWithNoPacketDelay)
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	//TODO: We will be removing Repartition method completely from sdks (golang and typescript) so not needed in PartitionInfo
	defaultConnection := partition_topology.NewPartitionConnection(partition_topology.NewPacketLoss(defaultConnectionInfo.PacketLossPercentage), partition_topology.ConnectionWithNoPacketDelay)

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
	serviceIdentifier := args.ServiceIdentifier
	err := service.serviceNetwork.PauseService(ctx, serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to pause service '%v'", serviceIdentifier)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) UnpauseService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnpauseServiceArgs) (*emptypb.Empty, error) {
	serviceIdentifier := args.ServiceIdentifier
	err := service.serviceNetwork.UnpauseService(ctx, serviceIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to unpause service '%v'", serviceIdentifier)
	}
	return &emptypb.Empty{}, nil
}

func (apicService ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdentifier := args.ServiceIdentifier
	command := args.CommandArgs
	exitCode, logOutput, err := apicService.serviceNetwork.ExecCommand(ctx, serviceIdentifier, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceIdentifier)
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

	serviceIdentifier := args.ServiceIdentifier

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdentifier,
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
	serviceIdentifier := args.ServiceIdentifier

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdentifier,
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
	filterServiceIdentifiers := args.ServiceIdentifiers

	// if there are any filters we fetch those services only
	if len(filterServiceIdentifiers) > 0 {
		for serviceIdentifier := range filterServiceIdentifiers {
			serviceInfo, err := apicService.getServiceInfo(ctx, serviceIdentifier)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to get service info for service '%v'", serviceIdentifier)
			}
			serviceInfos[serviceIdentifier] = serviceInfo
		}
		resp := binding_constructors.NewGetServicesResponse(serviceInfos)
		return resp, nil
	}

	// otherwise we fetch everything
	for serviceName := range apicService.serviceNetwork.GetServiceNames() {
		serviceNameStr := string(serviceName)
		serviceInfo, err := apicService.getServiceInfo(ctx, serviceNameStr)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get service info for service '%v'", serviceName)
		}
		serviceInfos[serviceNameStr] = serviceInfo
	}

	resp := binding_constructors.NewGetServicesResponse(serviceInfos)
	return resp, nil
}

func (apicService ApiContainerService) GetExistingAndHistoricalServiceIdentifiers(_ context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse, error) {
	allIdentifiers := apicService.serviceNetwork.GetExistingAndHistoricalServiceIdentifiers()
	return &kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse{AllIdentifiers: allIdentifiers}, nil
}

func (apicService ApiContainerService) UploadFilesArtifact(_ context.Context, args *kurtosis_core_rpc_api_bindings.UploadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {
	maybeArtifactName := args.GetName()
	if maybeArtifactName == "" {
		maybeArtifactName = apicService.filesArtifactStore.GenerateUniqueNameForFileArtifact()
	}

	filesArtifactUuid, err := apicService.serviceNetwork.UploadFilesArtifact(args.Data, maybeArtifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to upload the file")
	}

	response := &kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse{Uuid: string(filesArtifactUuid), Name: maybeArtifactName}
	return response, nil
}

func (apicService ApiContainerService) UploadFilesArtifactV2(serverStream kurtosis_core_rpc_api_bindings.ApiContainerService_UploadFilesArtifactV2Server) error {
	var maybeArtifactName string
	err := grpc_file_transfer.ReadBytesStream[kurtosis_core_rpc_api_bindings.FileArtifactChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse](
		serverStream,
		func(fileChunk *kurtosis_core_rpc_api_bindings.FileArtifactChunk) ([]byte, string, error) {
			if maybeArtifactName == "" {
				maybeArtifactName = fileChunk.GetName()
			} else if maybeArtifactName != fileChunk.GetName() {
				return nil, "", stacktrace.NewError("An unexpected error occurred receiving file artifacts chunk. artifact name was changed during the upload process")
			}
			return fileChunk.GetData(), fileChunk.GetPreviousChunkHash(), nil
		},
		func(fileContent []byte) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {
			if maybeArtifactName == "" {
				maybeArtifactName = apicService.filesArtifactStore.GenerateUniqueNameForFileArtifact()
			}

			// finished receiving all the chunks and assembling them into a single byte array
			filesArtifactUuid, err := apicService.serviceNetwork.UploadFilesArtifact(fileContent, maybeArtifactName)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred while trying to upload the file")
			}
			return &kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse{
				Uuid: string(filesArtifactUuid),
				Name: maybeArtifactName,
			}, nil
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "Error receiving file from CLI upload")
	}
	return nil
}

func (apicService ApiContainerService) DownloadFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse, error) {
	artifactIdentifier := args.Identifier
	if strings.TrimSpace(artifactIdentifier) == "" {
		return nil, stacktrace.NewError("Cannot download file with empty files artifact identifier")
	}

	filesArtifact, err := apicService.filesArtifactStore.GetFile(artifactIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact '%v'", artifactIdentifier)
	}

	fileBytes, err := ioutil.ReadFile(filesArtifact.GetAbsoluteFilepath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading files artifact file bytes")
	}

	resp := &kurtosis_core_rpc_api_bindings.DownloadFilesArtifactResponse{Data: fileBytes}
	return resp, nil
}

func (apicService ApiContainerService) StoreWebFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	url := args.Url
	artifactName := args.Name

	resp, err := http.Get(args.Url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the files artifact bytes", url)
	}
	defer resp.Body.Close()
	body := bufio.NewReader(resp.Body)

	filesArtifactUuId, err := apicService.filesArtifactStore.StoreFile(body, artifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred storing the file from URL '%v' in the files artifact store", url)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse{Uuid: string(filesArtifactUuId)}
	return response, nil
}

func (apicService ApiContainerService) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
	serviceIdentifier := args.ServiceIdentifier
	srcPath := args.SourcePath
	name := args.Name

	filesArtifactId, err := apicService.serviceNetwork.CopyFilesFromService(ctx, serviceIdentifier, srcPath, name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred copying source '%v' from service with identifier '%v'", srcPath, serviceIdentifier)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse{Uuid: string(filesArtifactId)}
	return response, nil
}

func (apicService ApiContainerService) RenderTemplatesToFilesArtifact(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactResponse, error) {
	templatesAndDataByDestinationRelFilepath := args.TemplatesAndDataByDestinationRelFilepath
	filesArtifactUuid, err := apicService.serviceNetwork.RenderTemplates(templatesAndDataByDestinationRelFilepath, args.Name)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while rendering templates to files artifact")
	}
	response := binding_constructors.NewRenderTemplatesToFilesArtifactResponse(string(filesArtifactUuid))
	return response, nil
}

func (apicService ApiContainerService) ListFilesArtifactNamesAndUuids(_ context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse, error) {
	filesArtifactsNamesAndUuids := apicService.filesArtifactStore.GetFileNamesAndUuids()
	var filesArtifactNamesAndUuids []*kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid
	for _, nameAndUuid := range filesArtifactsNamesAndUuids {
		fileNameAndUuidGrpcType := &kurtosis_core_rpc_api_bindings.FilesArtifactNameAndUuid{
			FileName: nameAndUuid.GetName(),
			FileUuid: string(nameAndUuid.GetUuid()),
		}
		filesArtifactNamesAndUuids = append(filesArtifactNamesAndUuids, fileNameAndUuidGrpcType)
	}
	return &kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse{FileNamesAndUuids: filesArtifactNamesAndUuids}, nil
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
		serviceIdStr,
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

func (apicService ApiContainerService) getServiceInfo(ctx context.Context, serviceIdentifier string) (*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		serviceIdentifier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for service '%v'", serviceIdentifier)
	}
	privatePorts := serviceObj.GetPrivatePorts()
	privateIp := serviceObj.GetRegistration().GetPrivateIP()
	maybePublicIp := serviceObj.GetMaybePublicIP()
	maybePublicPorts := serviceObj.GetMaybePublicPorts()
	serviceUuidStr := string(serviceObj.GetRegistration().GetUUID())
	serviceNameStr := string(serviceObj.GetRegistration().GetName())

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
		serviceUuidStr,
		serviceNameStr,
		uuid_generator.ShortenedUUIDString(serviceUuidStr),
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
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while verifying that '%v' exists in the package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	mainScriptToExecute, err := os.ReadFile(pathToMainFile)
	if err != nil {
		return "", startosis_errors.WrapWithInterpretationError(err, "An error occurred while reading '%v' in the package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	return string(mainScriptToExecute), nil
}

func (apicService ApiContainerService) runStarlark(parallelism int, dryRun bool, packageId string, serializedStarlark string, serializedParams string, stream grpc.ServerStream) {
	responseLineStream := apicService.startosisRunner.Run(stream.Context(), dryRun, parallelism, packageId, serializedStarlark, serializedParams)
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
