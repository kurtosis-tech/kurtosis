/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/kurtosis-tech/kurtosis/core/server/commons/yaml_parser"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_constants"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang/grpc_file_streaming"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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

	emptyFileArtifactIdentifier = ""
	unlimitedLineCount          = math.MaxInt
	allFilePermissionsForOwner  = 0700

	defaultCloudUserId       = ""
	defaultCloudInstanceId   = ""
	defaultImageDownloadMode = kurtosis_core_rpc_api_bindings.ImageDownloadMode_missing
	isScript                 = true
	isNotScript              = false
	isNotRemote              = false
	defaultParallelism       = 4
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

	restartPolicy kurtosis_core_rpc_api_bindings.RestartPolicy

	starlarkRun *kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse

	metricsClient metrics_client.MetricsClient
}

func NewApiContainerService(
	filesArtifactStore *enclave_data_directory.FilesArtifactStore,
	serviceNetwork service_network.ServiceNetwork,
	startosisRunner *startosis_engine.StartosisRunner,
	startosisModuleContentProvider startosis_packages.PackageContentProvider,
	restartPolicy kurtosis_core_rpc_api_bindings.RestartPolicy,
	metricsClient metrics_client.MetricsClient,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		filesArtifactStore:             filesArtifactStore,
		serviceNetwork:                 serviceNetwork,
		startosisRunner:                startosisRunner,
		startosisModuleContentProvider: startosisModuleContentProvider,
		restartPolicy:                  restartPolicy,
		starlarkRun: &kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse{
			PackageId:              startosis_constants.PackageIdPlaceholderForStandaloneScript,
			SerializedScript:       "",
			SerializedParams:       "",
			Parallelism:            defaultParallelism,
			RelativePathToMainFile: startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
			MainFunctionName:       "",
			ExperimentalFeatures:   []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{},
			RestartPolicy:          kurtosis_core_rpc_api_bindings.RestartPolicy_NEVER,
		},
		metricsClient: metricsClient,
	}

	return service, nil
}

func (apicService *ApiContainerService) RunStarlarkScript(args *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkScriptServer) error {
	serializedStarlarkScript := args.GetSerializedScript()
	serializedParams := args.GetSerializedParams()
	parallelism := int(args.GetParallelism())
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)
	mainFuncName := args.GetMainFunctionName()
	experimentalFeatures := args.GetExperimentalFeatures()
	cloudUserId := shared_utils.GetOrDefaultString(args.CloudUserId, defaultCloudUserId)
	cloudInstanceID := shared_utils.GetOrDefaultString(args.CloudInstanceId, defaultCloudInstanceId)
	ApiDownloadMode := shared_utils.GetOrDefault(args.ImageDownloadMode, defaultImageDownloadMode)

	downloadMode := convertFromImageDownloadModeAPI(ApiDownloadMode)

	metricsErr := apicService.metricsClient.TrackKurtosisRun(startosis_constants.PackageIdPlaceholderForStandaloneScript, isNotRemote, dryRun, isScript, cloudInstanceID, cloudUserId)
	if metricsErr != nil {
		logrus.Warn("An error occurred tracking kurtosis run event")
	}
	noPackageReplaceOptions := map[string]string{}

	apicService.runStarlark(parallelism, dryRun, startosis_constants.PackageIdPlaceholderForStandaloneScript, noPackageReplaceOptions, mainFuncName, startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript, serializedStarlarkScript, serializedParams, downloadMode, args.GetExperimentalFeatures(), stream)

	apicService.starlarkRun = &kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse{
		PackageId:              startosis_constants.PackageIdPlaceholderForStandaloneScript,
		SerializedScript:       serializedStarlarkScript,
		SerializedParams:       serializedParams,
		Parallelism:            int32(parallelism),
		RelativePathToMainFile: startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		MainFunctionName:       mainFuncName,
		ExperimentalFeatures:   experimentalFeatures,
		RestartPolicy:          apicService.restartPolicy,
	}

	return nil
}

func (apicService *ApiContainerService) UploadStarlarkPackage(server kurtosis_core_rpc_api_bindings.ApiContainerService_UploadStarlarkPackageServer) error {
	var packageId string
	serverStream := grpc_file_streaming.NewServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, emptypb.Empty](server)

	err := serverStream.ReceiveData(
		packageId,
		func(dataChunk *kurtosis_core_rpc_api_bindings.StreamedDataChunk) ([]byte, string, error) {
			if packageId == "" {
				packageId = dataChunk.GetMetadata().GetName()
			} else if packageId != dataChunk.GetMetadata().GetName() {
				return nil, "", stacktrace.NewError("An unexpected error occurred receiving Starlark package chunk. Package name was changed during the upload process")
			}
			return dataChunk.GetData(), dataChunk.GetPreviousChunkHash(), nil
		},
		func(assembledContent io.Reader) (*emptypb.Empty, error) {
			if packageId == "" {
				return nil, stacktrace.NewError("The package being uploaded did not have any name attached. This is a Kurtosis bug")
			}

			// finished receiving all the chunks and assembling them into a single byte array
			_, interpretationError := apicService.startosisModuleContentProvider.StorePackageContents(packageId, assembledContent, doOverwriteExistingModule)
			if interpretationError != nil {
				return nil, stacktrace.Propagate(interpretationError, "Error storing package in APIC once received")
			}
			return &emptypb.Empty{}, nil
		},
	)

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred receiving the Starlark package payload")
	}
	return nil
}

func (apicService *ApiContainerService) InspectFilesArtifactContents(_ context.Context, args *kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsRequest) (*kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse, error) {
	artifactIdentifier := ""
	if args.GetFileNamesAndUuid().GetFileUuid() != emptyFileArtifactIdentifier {
		artifactIdentifier = args.GetFileNamesAndUuid().GetFileUuid()
	}
	if args.GetFileNamesAndUuid().GetFileName() != emptyFileArtifactIdentifier {
		artifactIdentifier = args.GetFileNamesAndUuid().GetFileName()
	}
	if artifactIdentifier == emptyFileArtifactIdentifier {
		return nil, stacktrace.NewError("An error occurred because files artifact identifier is empty '%v'", artifactIdentifier)
	}

	_, filesArtifact, _, found, err := apicService.filesArtifactStore.GetFile(artifactIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting files artifact '%v'", artifactIdentifier)
	}
	if !found {
		return nil, stacktrace.NewError("An error occurred getting files artifact '%v', it doesn't exist in this enclave", artifactIdentifier)
	}

	fileDescriptions, err := getFileDescriptionsFromArtifact(filesArtifact.GetAbsoluteFilepath())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting file descriptions from '%v'", artifactIdentifier)
	}

	return &kurtosis_core_rpc_api_bindings.InspectFilesArtifactContentsResponse{
		FileDescriptions: fileDescriptions,
	}, nil
}

func (apicService *ApiContainerService) RunStarlarkPackage(args *kurtosis_core_rpc_api_bindings.RunStarlarkPackageArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkPackageServer) error {
	packageId := args.GetPackageId()
	parallelism := int(args.GetParallelism())
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)
	serializedParams := args.GetSerializedParams()
	relativePathToMainFile := args.GetRelativePathToMainFile()
	mainFuncName := args.GetMainFunctionName()
	cloudUserId := shared_utils.GetOrDefaultString(args.CloudUserId, defaultCloudUserId)
	cloudInstanceID := shared_utils.GetOrDefaultString(args.CloudInstanceId, defaultCloudInstanceId)
	ApiDownloadMode := shared_utils.GetOrDefault(args.ImageDownloadMode, defaultImageDownloadMode)

	downloadMode := convertFromImageDownloadModeAPI(ApiDownloadMode)

	if relativePathToMainFile == "" {
		relativePathToMainFile = startosis_constants.MainFileName
	}

	// TODO: remove this fork once everything uses the UploadStarlarkPackage endpoint prior to calling this
	//  right now the TS SDK still uses the old deprecated behavior
	var scriptWithRunFunction string
	var interpretationError *startosis_errors.InterpretationError
	var isRemote bool
	var kurtosisYml *yaml_parser.KurtosisYaml
	if args.ClonePackage != nil {
		scriptWithRunFunction, kurtosisYml, interpretationError = apicService.runStarlarkPackageSetup(packageId, args.GetClonePackage(), nil, relativePathToMainFile)
		isRemote = args.GetClonePackage()
	} else {
		// old deprecated syntax in use
		moduleContentIfLocal := args.GetLocal()
		isRemote = args.GetRemote()
		scriptWithRunFunction, kurtosisYml, interpretationError = apicService.runStarlarkPackageSetup(packageId, args.GetRemote(), moduleContentIfLocal, relativePathToMainFile)
	}
	if interpretationError != nil {
		if err := stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError.ToAPIType())); err != nil {
			return stacktrace.Propagate(err, "Error preparing for package execution and this error could not be sent through the output stream: '%s'", packageId)
		}
		return nil
	}
	packageName := kurtosisYml.GetPackageName()
	packageReplaceOptions := kurtosisYml.GetPackageReplaceOptions()
	logrus.Debugf("package replace options received '%+v'", packageReplaceOptions)

	metricsErr := apicService.metricsClient.TrackKurtosisRun(packageName, isRemote, dryRun, isNotScript, cloudInstanceID, cloudUserId)
	if metricsErr != nil {
		logrus.Warn("An error occurred tracking kurtosis run event")
	}
	apicService.runStarlark(parallelism, dryRun, packageName, packageReplaceOptions, mainFuncName, relativePathToMainFile, scriptWithRunFunction, serializedParams, downloadMode, args.ExperimentalFeatures, stream)

	apicService.starlarkRun = &kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse{
		PackageId:              packageId,
		SerializedScript:       scriptWithRunFunction,
		SerializedParams:       serializedParams,
		Parallelism:            int32(parallelism),
		RelativePathToMainFile: relativePathToMainFile,
		MainFunctionName:       mainFuncName,
		ExperimentalFeatures:   args.ExperimentalFeatures,
		RestartPolicy:          apicService.restartPolicy,
	}
	return nil
}

func (apicService *ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdentifier := args.ServiceIdentifier
	command := args.CommandArgs
	execResult, err := apicService.serviceNetwork.RunExec(ctx, serviceIdentifier, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceIdentifier)
	}
	numLogOutputBytes := len(execResult.GetOutput())
	if numLogOutputBytes > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError(
			"Log output from docker exec command '%+v' was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			numLogOutputBytes,
			maxLogOutputSizeBytes,
		)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode:  execResult.GetExitCode(),
		LogOutput: execResult.GetOutput(),
	}
	return resp, nil
}

func (apicService *ApiContainerService) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {

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

func (apicService *ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {
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

func (apicService *ApiContainerService) GetServices(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServicesArgs) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {
	serviceInfos := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	filterServiceIdentifiers := args.ServiceIdentifiers

	// if there are any filters we fetch those services only - this goes one by one
	// TODO (maybe) - explore perf differences between individual fetches vs filtering on the APIC side
	// Note as of 2023-08-23 I(gyani) has only seen instances of fetch everything & fetch one, so we don't need to optimize just yet
	if len(filterServiceIdentifiers) > 0 {
		for serviceIdentifier := range filterServiceIdentifiers {
			serviceInfo, err := apicService.getServiceInfoForIdentifier(ctx, serviceIdentifier)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Failed to get service info for service '%v'", serviceIdentifier)
			}
			serviceInfos[serviceIdentifier] = serviceInfo
		}
		resp := binding_constructors.NewGetServicesResponse(serviceInfos)
		return resp, nil
	}

	allServices, err := apicService.serviceNetwork.GetServices(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while fetching all services from the backend")
	}
	serviceInfos, err = getServiceInfosFromServiceObjs(allServices)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while converting the service obj into service info")
	}

	resp := binding_constructors.NewGetServicesResponse(serviceInfos)
	return resp, nil
}

func (apicService *ApiContainerService) ConnectServices(_ context.Context, _ *kurtosis_core_rpc_api_bindings.ConnectServicesArgs) (*kurtosis_core_rpc_api_bindings.ConnectServicesResponse, error) {
	resp := binding_constructors.NewConnectServicesResponse()
	return resp, nil
}

func (apicService *ApiContainerService) GetExistingAndHistoricalServiceIdentifiers(_ context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse, error) {
	allIdentifiers, err := apicService.serviceNetwork.GetExistingAndHistoricalServiceIdentifiers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting existing and historical service identifiers")
	}
	serviceIdentifiersGrpc := []*kurtosis_core_rpc_api_bindings.ServiceIdentifiers{}

	for _, serviceIdentifier := range allIdentifiers {
		serviceIdentifierGrpc := &kurtosis_core_rpc_api_bindings.ServiceIdentifiers{
			ServiceUuid:   string(serviceIdentifier.GetUuid()),
			ShortenedUuid: serviceIdentifier.GetShortenedUUIDStr(),
			Name:          string(serviceIdentifier.GetName()),
		}
		serviceIdentifiersGrpc = append(serviceIdentifiersGrpc, serviceIdentifierGrpc)
	}

	existingAndHistoricalServiceIdentifiersResponse := &kurtosis_core_rpc_api_bindings.GetExistingAndHistoricalServiceIdentifiersResponse{
		AllIdentifiers: serviceIdentifiersGrpc,
	}
	return existingAndHistoricalServiceIdentifiersResponse, nil
}

func (apicService *ApiContainerService) UploadFilesArtifact(server kurtosis_core_rpc_api_bindings.ApiContainerService_UploadFilesArtifactServer) error {
	var maybeArtifactName string
	serverStream := grpc_file_streaming.NewServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse](server)

	err := serverStream.ReceiveData(
		maybeArtifactName,
		func(dataChunk *kurtosis_core_rpc_api_bindings.StreamedDataChunk) ([]byte, string, error) {
			if maybeArtifactName == "" {
				maybeArtifactName = dataChunk.GetMetadata().GetName()
			} else if maybeArtifactName != dataChunk.GetMetadata().GetName() {
				return nil, "", stacktrace.NewError("An unexpected error occurred receiving file artifacts chunk. Artifact name was changed during the upload process")
			}
			return dataChunk.GetData(), dataChunk.GetPreviousChunkHash(), nil
		},
		func(assembledContent io.Reader) (*kurtosis_core_rpc_api_bindings.UploadFilesArtifactResponse, error) {
			if maybeArtifactName == "" {
				maybeArtifactName = apicService.filesArtifactStore.GenerateUniqueNameForFileArtifact()
			}

			// finished receiving all the chunks and assembling them into a single byte array
			// TODO: pass in the md5 from the CLI (which currently drops it because APIC API doesn't accept it)
			//  for now it's fine, it's just that file hash comparison for this file will always return false
			filesArtifactUuid, err := apicService.serviceNetwork.UploadFilesArtifact(assembledContent, []byte{}, maybeArtifactName)
			if err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred while trying to upload the file")
			}
			return binding_constructors.NewUploadFilesArtifactResponse(string(filesArtifactUuid), maybeArtifactName), nil
		},
	)

	if err != nil {
		return stacktrace.Propagate(err, "An error occurred receiving the file payload")
	}
	return nil
}

func (apicService *ApiContainerService) DownloadFilesArtifact(args *kurtosis_core_rpc_api_bindings.DownloadFilesArtifactArgs, server kurtosis_core_rpc_api_bindings.ApiContainerService_DownloadFilesArtifactServer) error {
	artifactIdentifier := args.Identifier
	if strings.TrimSpace(artifactIdentifier) == "" {
		return stacktrace.NewError("Cannot download file with empty files artifact identifier")
	}

	_, filesArtifact, _, found, err := apicService.filesArtifactStore.GetFile(artifactIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting files artifact '%v'", artifactIdentifier)
	}
	if !found {
		return stacktrace.NewError("An error occurred getting files artifact '%v', it doesn't exist in this enclave", artifactIdentifier)
	}

	file, err := os.OpenFile(filesArtifact.GetAbsoluteFilepath(), os.O_RDONLY, allFilePermissionsForOwner)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading files artifact file at '%s'",
			filesArtifact.GetAbsoluteFilepath())
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting files artifact file at '%s'",
			filesArtifact.GetAbsoluteFilepath())
	}
	var fileSize uint64
	if fileInfo.Size() >= 0 {
		fileSize = uint64(fileInfo.Size())
	} else {
		return stacktrace.Propagate(err, "An error occurred inspecting files artifact file at '%s'. The size of "+
			"the file is a negative integer",
			filesArtifact.GetAbsoluteFilepath())
	}

	serverStream := grpc_file_streaming.NewServerStream[kurtosis_core_rpc_api_bindings.StreamedDataChunk, []byte](server)
	err = serverStream.SendData(
		args.Identifier,
		file,
		fileSize,
		func(previousChunkHash string, contentChunk []byte) (*kurtosis_core_rpc_api_bindings.StreamedDataChunk, error) {
			return &kurtosis_core_rpc_api_bindings.StreamedDataChunk{
				Data:              contentChunk,
				PreviousChunkHash: previousChunkHash,
				Metadata: &kurtosis_core_rpc_api_bindings.DataChunkMetadata{
					Name: args.Identifier,
				},
			}, nil
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred receiving the file payload")
	}
	return nil
}

func (apicService *ApiContainerService) StoreWebFilesArtifact(_ context.Context, args *kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactArgs) (*kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse, error) {
	url := args.Url
	artifactName := args.Name

	resp, err := http.Get(args.Url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred making the request to URL '%v' to get the files artifact bytes", url)
	}
	defer resp.Body.Close()
	body := bufio.NewReader(resp.Body)

	// TODO: we should probably wrap the web file into a file artifact here, not sure how files look in the APIC since
	//  it might not even be a TGZ.
	filesArtifactUuId, err := apicService.filesArtifactStore.StoreFile(body, []byte{}, artifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred storing the file from URL '%v' in the files artifact store", url)
	}

	response := &kurtosis_core_rpc_api_bindings.StoreWebFilesArtifactResponse{Uuid: string(filesArtifactUuId)}
	return response, nil
}

func (apicService *ApiContainerService) StoreFilesArtifactFromService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceArgs) (*kurtosis_core_rpc_api_bindings.StoreFilesArtifactFromServiceResponse, error) {
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

func (apicService *ApiContainerService) ListFilesArtifactNamesAndUuids(_ context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.ListFilesArtifactNamesAndUuidsResponse, error) {
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

func (apicService *ApiContainerService) GetStarlarkRun(_ context.Context, _ *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse, error) {
	return apicService.starlarkRun, nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func transformPortSpecToApiPort(port *port_spec.PortSpec) (*kurtosis_core_rpc_api_bindings.Port, error) {
	portNumUint16 := port.GetNumber()
	portSpecProto := port.GetTransportProtocol()

	// Yes, this isn't the most efficient way to do this, but the map is tiny, so it doesn't matter
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

	maybeWaitTimeout := ""
	if port.GetWait() != nil {
		maybeWaitTimeout = port.GetWait().GetTimeout().String()
	}

	result := binding_constructors.NewPort(uint32(portNumUint16), apiProto, maybeApplicationProtocol, maybeWaitTimeout)
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

func (apicService *ApiContainerService) waitForEndpointAvailability(
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
	if serviceObj.GetContainer().GetStatus() != container.ContainerStatus_Running {
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

		bodyBytes, err := io.ReadAll(body)

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

func (apicService *ApiContainerService) getServiceInfoForIdentifier(ctx context.Context, serviceIdentifier string) (*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	serviceObj, err := apicService.serviceNetwork.GetService(
		ctx,
		serviceIdentifier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for service '%v'", serviceIdentifier)
	}
	serviceInfo, err := getServiceInfoFromServiceObj(serviceObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "an error occurred while converting service obj for service with id '%v' to service info", serviceIdentifier)
	}
	return serviceInfo, nil
}

func (apicService *ApiContainerService) runStarlarkPackageSetup(
	packageId string,
	clonePackage bool,
	moduleContentIfLocal []byte,
	relativePathToMainFile string,
) (string, *yaml_parser.KurtosisYaml, *startosis_errors.InterpretationError) {
	var packageRootPathOnDisk string
	var interpretationError *startosis_errors.InterpretationError

	if clonePackage {
		packageRootPathOnDisk, interpretationError = apicService.startosisModuleContentProvider.ClonePackage(packageId)
	} else if moduleContentIfLocal != nil {
		// TODO: remove this once UploadStarlarkPackage is called prior to calling RunStarlarkPackage by all consumers
		//  of this API
		packageRootPathOnDisk, interpretationError = apicService.startosisModuleContentProvider.StorePackageContents(packageId, bytes.NewReader(moduleContentIfLocal), doOverwriteExistingModule)
	} else {
		// We just need to retrieve the absolute path, the content is will not be stored here since it has been uploaded
		// prior to this call
		packageRootPathOnDisk, interpretationError = apicService.startosisModuleContentProvider.GetOnDiskAbsolutePackagePath(packageId)

	}
	if interpretationError != nil {
		return "", nil, interpretationError
	}

	kurtosisYml, interpretationError := apicService.startosisModuleContentProvider.GetKurtosisYaml(packageRootPathOnDisk)
	if interpretationError != nil {
		return "", nil, interpretationError
	}

	pathToMainFile := path.Join(packageRootPathOnDisk, relativePathToMainFile)

	if _, err := os.Stat(pathToMainFile); err != nil {
		return "", nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while verifying that '%v' exists in the package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	mainScriptToExecute, err := os.ReadFile(pathToMainFile)
	if err != nil {
		return "", nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while reading '%v' in the package '%v' at '%v'", startosis_constants.MainFileName, packageId, pathToMainFile)
	}

	return string(mainScriptToExecute), kurtosisYml, nil
}

func (apicService *ApiContainerService) runStarlark(
	parallelism int,
	dryRun bool,
	packageId string,
	packageReplaceOptions map[string]string,
	mainFunctionName string,
	relativePathToMainFile string,
	serializedStarlark string,
	serializedParams string,
	imageDownloadMode image_download_mode.ImageDownloadMode,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	stream grpc.ServerStream,
) {
	responseLineStream := apicService.startosisRunner.Run(stream.Context(), dryRun, parallelism, packageId, packageReplaceOptions, mainFunctionName, relativePathToMainFile, serializedStarlark, serializedParams, imageDownloadMode, experimentalFeatures)
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

func getServiceInfosFromServiceObjs(services map[service.ServiceUUID]*service.Service) (map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	serviceInfos := map[string]*kurtosis_core_rpc_api_bindings.ServiceInfo{}
	for uuid, serviceObj := range services {
		serviceInfo, err := getServiceInfoFromServiceObj(serviceObj)
		if err != nil {
			return nil, stacktrace.Propagate(err, "there was an error converting the service obj for service with uuid '%v' and name '%v' to service info", uuid, serviceObj.GetRegistration().GetName())
		}
		serviceInfos[serviceInfo.Name] = serviceInfo
	}
	return serviceInfos, nil
}

func getServiceInfoFromServiceObj(serviceObj *service.Service) (*kurtosis_core_rpc_api_bindings.ServiceInfo, error) {
	privatePorts := serviceObj.GetPrivatePorts()
	privateIp := serviceObj.GetRegistration().GetPrivateIP()
	maybePublicIp := serviceObj.GetMaybePublicIP()
	maybePublicPorts := serviceObj.GetMaybePublicPorts()
	serviceUuidStr := string(serviceObj.GetRegistration().GetUUID())
	serviceNameStr := string(serviceObj.GetRegistration().GetName())
	serviceStatus, err := convertServiceStatusToServiceInfoStatus(serviceObj.GetRegistration().GetStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the service status to a service info status")
	}
	serviceContainer := serviceObj.GetContainer()
	serviceInfoContainerStatus, err := convertContainerStatusToServiceInfoContainerStatus(serviceContainer.GetStatus())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred converting the service container status to a service info container status")
	}

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
	serviceInfoContainer := &kurtosis_core_rpc_api_bindings.Container{
		Status:         serviceInfoContainerStatus,
		ImageName:      serviceContainer.GetImageName(),
		EntrypointArgs: serviceContainer.GetEntrypointArgs(),
		CmdArgs:        serviceContainer.GetCmdArgs(),
		EnvVars:        serviceContainer.GetEnvVars(),
	}

	serviceInfoResponse := binding_constructors.NewServiceInfo(
		serviceUuidStr,
		serviceNameStr,
		uuid_generator.ShortenedUUIDString(serviceUuidStr),
		privateIp.String(),
		privateApiPorts,
		publicIpAddrStr,
		publicApiPorts,
		serviceStatus,
		serviceInfoContainer,
	)
	return serviceInfoResponse, nil
}

func getFileDescriptionsFromArtifact(artifactPath string) ([]*kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription, error) {
	file, err := os.Open(artifactPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get file descriptions for artifact path '%v'", artifactPath)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get create gzip reader for artifact path '%v'", artifactPath)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	fileDescriptions := []*kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, stacktrace.Propagate(err, "Failed to get header from artifact path '%v'", artifactPath)
		}

		filePath := header.Name
		fileSize := header.Size
		textPreview, err := getTextRepresentation(tarReader, unlimitedLineCount)
		if err != nil {
			// TODO(vcolombo): Return this as part of the request?
			logrus.Debugf("Failed to get text preview for file '%v' with error '%v'", filePath, err)
		}
		fileDescriptions = append(fileDescriptions, &kurtosis_core_rpc_api_bindings.FileArtifactContentsFileDescription{
			Path:        filePath,
			Size:        uint64(fileSize),
			TextPreview: textPreview,
		})
	}
	return fileDescriptions, nil
}

func getTextRepresentation(reader io.Reader, lineCount int) (*string, error) {
	scanner := bufio.NewScanner(reader)
	textRepresentation := strings.Builder{}
	for i := 0; i < lineCount && scanner.Scan(); i += 1 {
		line := scanner.Text()
		for _, char := range line {
			if !unicode.IsPrint(char) {
				return nil, stacktrace.NewError("File has no text representation because '%v' is not printable", char)
			}
		}
		textRepresentation.WriteString(fmt.Sprintf("%s\n", line))
	}

	if err := scanner.Err(); err != nil {
		return nil, stacktrace.Propagate(err, "Scanning file failed")
	}

	text := textRepresentation.String()
	return &text, nil
}

func convertServiceStatusToServiceInfoStatus(serviceStatus service.ServiceStatus) (kurtosis_core_rpc_api_bindings.ServiceStatus, error) {
	switch serviceStatus {
	case service.ServiceStatus_Started:
		return kurtosis_core_rpc_api_bindings.ServiceStatus_RUNNING, nil
	case service.ServiceStatus_Stopped:
		return kurtosis_core_rpc_api_bindings.ServiceStatus_STOPPED, nil
	default:
		return kurtosis_core_rpc_api_bindings.ServiceStatus_UNKNOWN, stacktrace.NewError("Failed to convert service status %v", serviceStatus)
	}
}

func convertContainerStatusToServiceInfoContainerStatus(containerStatus container.ContainerStatus) (kurtosis_core_rpc_api_bindings.Container_Status, error) {
	switch containerStatus {
	case container.ContainerStatus_Running:
		return kurtosis_core_rpc_api_bindings.Container_RUNNING, nil
	case container.ContainerStatus_Stopped:
		return kurtosis_core_rpc_api_bindings.Container_STOPPED, nil
	default:
		return kurtosis_core_rpc_api_bindings.Container_UNKNOWN, stacktrace.NewError("Failed to convert container status %v", containerStatus)
	}
}

func convertFromImageDownloadModeAPI(api_mode kurtosis_core_rpc_api_bindings.ImageDownloadMode) image_download_mode.ImageDownloadMode {
	switch api_mode {
	case kurtosis_core_rpc_api_bindings.ImageDownloadMode_always:
		return image_download_mode.ImageDownloadMode_Always
	case kurtosis_core_rpc_api_bindings.ImageDownloadMode_missing:
		return image_download_mode.ImageDownloadMode_Missing
	default:
		panic(stacktrace.NewError("Failed to convert image download mode %v", api_mode))
	}
}
