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
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/docker_compose_transpiler"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_run"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages/git_package_content_provider"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
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

	startosisInterpreter *startosis_engine.StartosisInterpreter

	packageContentProvider startosis_packages.PackageContentProvider

	starlarkRunRepository *starlark_run.StarlarkRunRepository

	metricsClient metrics_client.MetricsClient

	githubAuthProvider *git_package_content_provider.GitHubPackageAuthProvider
}

func NewApiContainerService(
	filesArtifactStore *enclave_data_directory.FilesArtifactStore,
	serviceNetwork service_network.ServiceNetwork,
	startosisRunner *startosis_engine.StartosisRunner,
	startosisInterpreter *startosis_engine.StartosisInterpreter,
	startosisModuleContentProvider startosis_packages.PackageContentProvider,
	restartPolicy kurtosis_core_rpc_api_bindings.RestartPolicy,
	metricsClient metrics_client.MetricsClient,
	githubAuthProvider *git_package_content_provider.GitHubPackageAuthProvider,
	starlarkRunRepository *starlark_run.StarlarkRunRepository,
) (*ApiContainerService, error) {

	if err := initStarlarkRun(starlarkRunRepository, restartPolicy); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred initializing the starlark run")
	}

	service := &ApiContainerService{
		filesArtifactStore:     filesArtifactStore,
		serviceNetwork:         serviceNetwork,
		startosisRunner:        startosisRunner,
		startosisInterpreter:   startosisInterpreter,
		packageContentProvider: startosisModuleContentProvider,
		starlarkRunRepository:  starlarkRunRepository,
		metricsClient:          metricsClient,
		githubAuthProvider:     githubAuthProvider,
	}

	return service, nil
}

func (apicService *ApiContainerService) RunStarlarkScript(args *kurtosis_core_rpc_api_bindings.RunStarlarkScriptArgs, stream kurtosis_core_rpc_api_bindings.ApiContainerService_RunStarlarkScriptServer) error {
	serializedStarlarkScript := args.GetSerializedScript()
	serializedParams := args.GetSerializedParams()
	parallelism := args.GetParallelism()
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)
	mainFuncName := args.GetMainFunctionName()
	experimentalFeatures := args.GetExperimentalFeatures()
	ApiDownloadMode := shared_utils.GetOrDefault(args.ImageDownloadMode, defaultImageDownloadMode)
	nonBlockingMode := args.GetNonBlockingMode()
	downloadMode := convertFromImageDownloadModeAPI(ApiDownloadMode)

	metricsErr := apicService.metricsClient.TrackKurtosisRun(startosis_constants.PackageIdPlaceholderForStandaloneScript, isNotRemote, dryRun, isScript)
	if metricsErr != nil {
		logrus.Warn("An error occurred tracking kurtosis run event")
	}
	noPackageReplaceOptions := map[string]string{}

	apicService.runStarlark(
		int(parallelism),
		dryRun,
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		noPackageReplaceOptions,
		mainFuncName,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		serializedStarlarkScript,
		serializedParams,
		downloadMode,
		nonBlockingMode,
		experimentalFeatures,
		stream,
	)

	return nil
}

func (apicService *ApiContainerService) saveStarlarkRun(
	packageId string,
	serializedParams string,
	serializedStarlarkScript string,
	parallelism int,
	mainFuncName string,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
) error {

	var (
		initialSerializedParams string
		restartPolicy           = int32(kurtosis_core_rpc_api_bindings.RestartPolicy_NEVER.Number())
	)

	previousStarlarkRun, err := apicService.starlarkRunRepository.Get()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the starlark run object from the repository")
	}

	if previousStarlarkRun != nil {
		restartPolicy = previousStarlarkRun.GetRestartPolicy()
		if previousStarlarkRun.GetInitialSerializedParams() == "" {
			initialSerializedParams = serializedParams
		}
	}

	experimentalFeaturesSlice := []int32{}
	for _, feature := range experimentalFeatures {
		experimentalFeaturesSlice = append(experimentalFeaturesSlice, int32(feature.Number()))
	}

	currentStarlarkRun := starlark_run.NewStarlarkRun(
		packageId,
		serializedStarlarkScript,
		serializedParams,
		int32(parallelism),
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		mainFuncName,
		experimentalFeaturesSlice,
		restartPolicy,
		initialSerializedParams,
	)

	if err := apicService.starlarkRunRepository.Save(currentStarlarkRun); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving the current starlark run in the repository")
	}

	return nil
}

// Uploads a Starlark package for later execution
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
			_, interpretationError := apicService.packageContentProvider.StorePackageContents(packageId, assembledContent, doOverwriteExistingModule)
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

	var scriptWithRunFunction string
	var interpretationError *startosis_errors.InterpretationError
	var isRemote bool
	var detectedPackageId string
	var detectedPackageReplaceOptions map[string]string
	packageIdFromArgs := args.GetPackageId()
	parallelism := args.GetParallelism()
	if parallelism == 0 {
		parallelism = defaultParallelism
	}
	dryRun := shared_utils.GetOrDefaultBool(args.DryRun, defaultStartosisDryRun)
	serializedParams := args.GetSerializedParams()
	requestedRelativePathToMainFile := args.GetRelativePathToMainFile()
	mainFuncName := args.GetMainFunctionName()
	ApiDownloadMode := shared_utils.GetOrDefault(args.ImageDownloadMode, defaultImageDownloadMode)
	downloadMode := convertFromImageDownloadModeAPI(ApiDownloadMode)
	nonBlockingMode := args.GetNonBlockingMode()

	packageGitHubAuthToken := args.GetGithubAuthToken()
	if packageGitHubAuthToken != "" {
		err := apicService.githubAuthProvider.StoreGitHubTokenForPackage(packageIdFromArgs, args.GetGithubAuthToken())
		if err != nil {
			if err = stream.SendMsg(binding_constructors.NewStarlarkExecutionError(err.Error())); err != nil {
				return stacktrace.Propagate(err, "Error occurred providing github auth token for package: '%s'", packageIdFromArgs)
			}
		}
	}

	var actualRelativePathToMainFile string
	if args.ClonePackage != nil {
		scriptWithRunFunction, actualRelativePathToMainFile, detectedPackageId, detectedPackageReplaceOptions, interpretationError =
			apicService.runStarlarkPackageSetup(packageIdFromArgs, args.GetClonePackage(), nil, requestedRelativePathToMainFile)
		isRemote = args.GetClonePackage()
	} else {
		// OLD DEPRECATED SYNTAX
		// TODO: remove this fork once everything uses the UploadStarlarkPackage endpoint prior to calling this
		//  right now the TS SDK still uses the old deprecated behavior
		moduleContentIfLocal := args.GetLocal()
		isRemote = args.GetRemote()
		scriptWithRunFunction, actualRelativePathToMainFile, detectedPackageId, detectedPackageReplaceOptions, interpretationError =
			apicService.runStarlarkPackageSetup(packageIdFromArgs, args.GetRemote(), moduleContentIfLocal, requestedRelativePathToMainFile)
	}
	if interpretationError != nil {
		if err := stream.SendMsg(binding_constructors.NewStarlarkRunResponseLineFromInterpretationError(interpretationError.ToAPIType())); err != nil {
			return stacktrace.Propagate(err, "Error preparing for package execution and this error could not be sent through the output stream: '%s'", packageIdFromArgs)
		}
		return nil
	}
	logrus.Debugf("package replace options received '%+v'", detectedPackageReplaceOptions)

	metricsErr := apicService.metricsClient.TrackKurtosisRun(detectedPackageId, isRemote, dryRun, isNotScript)
	if metricsErr != nil {
		logrus.Warn("An error occurred tracking kurtosis run event")
	}

	logrus.Infof("package id: %v\n main func name: %v\n actual relative path to main file: %v\n script with run func: %v\n serialized params:%v\n",
		detectedPackageId,
		mainFuncName,
		actualRelativePathToMainFile,
		scriptWithRunFunction,
		serializedParams)
	apicService.runStarlark(int(parallelism), dryRun, detectedPackageId, detectedPackageReplaceOptions, mainFuncName, actualRelativePathToMainFile, scriptWithRunFunction, serializedParams, downloadMode, nonBlockingMode, args.ExperimentalFeatures, stream)

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

	serviceIdentifier := args.GetServiceIdentifier()

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdentifier,
		http.MethodGet,
		args.GetPort(),
		args.GetPath(),
		args.GetInitialDelayMilliseconds(),
		args.GetRetries(),
		args.GetRetriesDelayMilliseconds(),
		"",
		args.GetBodyText()); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.GetPath(),
		)
	}

	return &emptypb.Empty{}, nil
}

func (apicService *ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	serviceIdentifier := args.GetServiceIdentifier()

	if err := apicService.waitForEndpointAvailability(
		ctx,
		serviceIdentifier,
		http.MethodPost,
		args.GetPort(),
		args.GetPath(),
		args.GetInitialDelayMilliseconds(),
		args.GetRetries(),
		args.GetRetriesDelayMilliseconds(),
		args.GetRequestBody(),
		args.GetBodyText()); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.GetPath(),
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

	currentStarlarkRun, err := apicService.starlarkRunRepository.Get()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the starlark run from the repository")
	}

	kurtosisFeatureFlags := []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{}
	for _, experimentalFeature := range currentStarlarkRun.GetExperimentalFeatures() {
		kurtosisFeatureFlags = append(kurtosisFeatureFlags, kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag(experimentalFeature))
	}
	restartPolicy := kurtosis_core_rpc_api_bindings.RestartPolicy(currentStarlarkRun.GetRestartPolicy())
	initialSerializedParams := currentStarlarkRun.GetInitialSerializedParams()

	getStarlarkRunResponse := &kurtosis_core_rpc_api_bindings.GetStarlarkRunResponse{
		PackageId:               currentStarlarkRun.GetPackageId(),
		SerializedScript:        currentStarlarkRun.GetSerializedScript(),
		SerializedParams:        currentStarlarkRun.GetSerializedParams(),
		Parallelism:             currentStarlarkRun.GetParallelism(),
		RelativePathToMainFile:  currentStarlarkRun.GetRelativePathToMainFile(),
		MainFunctionName:        currentStarlarkRun.GetMainFunctionName(),
		ExperimentalFeatures:    kurtosisFeatureFlags,
		RestartPolicy:           restartPolicy,
		InitialSerializedParams: &initialSerializedParams,
	}

	return getStarlarkRunResponse, nil
}

func (apicService *ApiContainerService) GetStarlarkPackagePlanYaml(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StarlarkPackagePlanYamlArgs) (*kurtosis_core_rpc_api_bindings.PlanYaml, error) {
	packageIdFromArgs := args.GetPackageId()
	serializedParams := args.GetSerializedParams()
	requestedRelativePathToMainFile := args.GetRelativePathToMainFile()
	mainFuncName := args.GetMainFunctionName()

	var scriptWithRunFunction string
	var interpretationError *startosis_errors.InterpretationError
	var detectedPackageId string
	var detectedPackageReplaceOptions map[string]string
	var actualRelativePathToMainFile string
	scriptWithRunFunction, actualRelativePathToMainFile, detectedPackageId, detectedPackageReplaceOptions, interpretationError =
		apicService.runStarlarkPackageSetup(packageIdFromArgs, args.IsRemote, nil, requestedRelativePathToMainFile)
	if interpretationError != nil {
		return nil, stacktrace.Propagate(interpretationError, "An interpretation error occurred setting up the package for retrieving plan yaml for package: %v", packageIdFromArgs)
	}

	_, instructionsPlan, apiInterpretationError := apicService.startosisInterpreter.Interpret(
		ctx,
		detectedPackageId,
		mainFuncName,
		detectedPackageReplaceOptions,
		actualRelativePathToMainFile,
		scriptWithRunFunction,
		serializedParams,
		false,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0),
		image_download_mode.ImageDownloadMode_Always)
	if apiInterpretationError != nil {
		interpretationError = startosis_errors.NewInterpretationError(apiInterpretationError.GetErrorMessage())
		return nil, stacktrace.Propagate(interpretationError, "An interpretation error occurred interpreting package for retrieving plan yaml for package: %v", packageIdFromArgs)
	}
	planYamlStr, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(packageIdFromArgs))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating plan yaml for package: %v", packageIdFromArgs)
	}

	return &kurtosis_core_rpc_api_bindings.PlanYaml{PlanYaml: planYamlStr}, nil
}

// NOTE: GetStarlarkScriptPlanYaml endpoint is only meant to be called by the EM UI, Enclave Builder logic.
// Once the EM UI retrieves the plan yaml, the APIC is removed and not used again.
// It's not ideal that we have to even start an enclave/APIC to simply get the result of interpretation/plan yaml but that would require a larger refactor
// of the startosis_engine to enable the infra for interpretation to be executed as a standalone library, that could be setup by the engine, or even on the client.
func (apicService *ApiContainerService) GetStarlarkScriptPlanYaml(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StarlarkScriptPlanYamlArgs) (*kurtosis_core_rpc_api_bindings.PlanYaml, error) {
	serializedStarlarkScript := args.GetSerializedScript()
	serializedParams := args.GetSerializedParams()
	mainFuncName := args.GetMainFunctionName()
	noPackageReplaceOptions := map[string]string{}

	_, instructionsPlan, apiInterpretationError := apicService.startosisInterpreter.Interpret(
		ctx,
		startosis_constants.PackageIdPlaceholderForStandaloneScript,
		mainFuncName,
		noPackageReplaceOptions,
		startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
		serializedStarlarkScript,
		serializedParams,
		false,
		enclave_structure.NewEnclaveComponents(),
		resolver.NewInstructionsPlanMask(0),
		image_download_mode.ImageDownloadMode_Always)
	if apiInterpretationError != nil {
		return nil, startosis_errors.NewInterpretationError(apiInterpretationError.GetErrorMessage())
	}
	planYamlStr, err := instructionsPlan.GenerateYaml(plan_yaml.CreateEmptyPlan(startosis_constants.PackageIdPlaceholderForStandaloneScript))
	if err != nil {
		return nil, err
	}

	return &kurtosis_core_rpc_api_bindings.PlanYaml{PlanYaml: planYamlStr}, nil
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
	packageIdFromArgs string,
	clonePackage bool,
	moduleContentIfLocal []byte, // empty if clonePackage is set to true
	relativePathToMainFile string, // could be empty
) (
	string, // Entrypoint script to execute
	string, // Detected relative path (from package root) to main script
	string, // Detected Package ID detected from [clonePackage] or [moduleContentIfLocal]
	map[string]string, // Replace options detected from [clonePackage] or [moduleContentIfLocal]
	*startosis_errors.InterpretationError) {
	var packageRootPathOnDisk string
	var interpretationError *startosis_errors.InterpretationError

	if clonePackage {
		packageRootPathOnDisk, interpretationError = apicService.packageContentProvider.ClonePackage(packageIdFromArgs)
	} else if moduleContentIfLocal != nil {
		// TODO: remove this once UploadStarlarkPackage is called prior to calling RunStarlarkPackage by all consumers of this API
		packageRootPathOnDisk, interpretationError = apicService.packageContentProvider.StorePackageContents(packageIdFromArgs, bytes.NewReader(moduleContentIfLocal), doOverwriteExistingModule)
	} else {
		// We just need to retrieve the absolute path, the content will not be stored here since it has been uploaded prior to this call
		// This is used in the flow where we `replace` with a local call, in which case we need to store multiple packages on the APIC
		// before we actually do the run
		packageRootPathOnDisk, interpretationError = apicService.packageContentProvider.GetOnDiskAbsolutePackagePath(packageIdFromArgs)
	}
	if interpretationError != nil {
		return "", "", "", nil, interpretationError
	}

	// If kurtosis.yml exists in root, treat as kurtosis package
	candidateKurtosisYmlAbsFilepath := path.Join(packageRootPathOnDisk, startosis_constants.KurtosisYamlName)
	if _, err := os.Stat(candidateKurtosisYmlAbsFilepath); err == nil {
		kurtosisYml, interpretationError := apicService.packageContentProvider.GetKurtosisYaml(packageRootPathOnDisk)
		if interpretationError != nil {
			return "", "", "", nil, interpretationError
		}
		if relativePathToMainFile == "" {
			relativePathToMainFile = startosis_constants.MainFileName
		}
		pathToMainFile := path.Join(packageRootPathOnDisk, relativePathToMainFile)
		if _, err := os.Stat(pathToMainFile); err != nil {
			return "", "", "", nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while verifying that '%v' exists in the package '%v' at '%v'", startosis_constants.MainFileName, packageIdFromArgs, pathToMainFile)
		}
		mainScriptToExecuteBytes, err := os.ReadFile(pathToMainFile)
		if err != nil {
			return "", "", "", nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while reading '%v' in the package '%v' at '%v'", startosis_constants.MainFileName, packageIdFromArgs, pathToMainFile)
		}
		return string(mainScriptToExecuteBytes), relativePathToMainFile, kurtosisYml.PackageName, kurtosisYml.PackageReplaceOptions, nil
	}

	// If kurtosis.yml doesn't exist, assume a Compose package and transpile compose into starlark
	if relativePathToMainFile == "" {
		for _, defaultComposeFilename := range docker_compose_transpiler.DefaultComposeFilenames {
			candidateComposeAbsFilepath := path.Join(packageRootPathOnDisk, defaultComposeFilename)
			if _, err := os.Stat(candidateComposeAbsFilepath); err == nil {
				relativePathToMainFile = defaultComposeFilename
				break
			}
		}
		if relativePathToMainFile == "" {
			return "", "", "", nil, startosis_errors.NewInterpretationError(
				"No '%s' file was found in the package root so fell back to Docker Compose package, but no "+
					"default Compose files (%s) were found. Either add a '%s' file to the package root or add one of the "+
					"default Compose files.",
				startosis_constants.KurtosisYamlName,
				strings.Join(docker_compose_transpiler.DefaultComposeFilenames, ", "),
				startosis_constants.KurtosisYamlName,
			)
		}
	}
	mainScriptToExecute, transpilationErr := docker_compose_transpiler.TranspileDockerComposePackageToStarlark(packageRootPathOnDisk, relativePathToMainFile)
	if transpilationErr != nil {
		return "", "", "", nil, startosis_errors.WrapWithInterpretationError(transpilationErr, "An error occurred transpiling the Docker Compose package '%v' to Starlark", packageIdFromArgs)
	}

	replacesForComposePackage := map[string]string{}
	return mainScriptToExecute, relativePathToMainFile, packageIdFromArgs, replacesForComposePackage, nil
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
	nonBlockingMode bool,
	experimentalFeatures []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag,
	stream grpc.ServerStream,
) {
	responseLineStream := apicService.startosisRunner.Run(stream.Context(), dryRun, parallelism, packageId, packageReplaceOptions, mainFunctionName, relativePathToMainFile, serializedStarlark, serializedParams, imageDownloadMode, nonBlockingMode, experimentalFeatures)
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

			if runFinishedEvent := responseLine.GetRunFinishedEvent(); runFinishedEvent != nil {

				if err := apicService.saveStarlarkRun(
					packageId,
					serializedParams,
					serializedStarlark,
					parallelism,
					mainFunctionName,
					experimentalFeatures,
				); err != nil {
					logrus.Errorf("The Starlark code was successfully executed but something failed when trying to save the 'run' info in the enclave's database. Error was: \n%v", err.Error())
				}

				isSuccessful := runFinishedEvent.GetIsRunSuccessful()
				numberOfServicesAfterRunFinished := 0
				if serviceNames, err := apicService.serviceNetwork.GetServiceNames(); err != nil {
					logrus.Warn("Couldn't figure out the number of services after run finished, will be logging 0 in the metrics")
				} else {
					numberOfServicesAfterRunFinished = len(serviceNames)
				}
				if err := apicService.metricsClient.TrackKurtosisRunFinishedEvent(packageId, numberOfServicesAfterRunFinished, isSuccessful); err != nil {
					logrus.Warn("An error occurred tracking the run-finished event")
				}
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
	case service.ServiceStatus_Registered:
		// missing case flagged by the linter, returning default to keep the same behavior since there is match on gRPC api
		return kurtosis_core_rpc_api_bindings.ServiceStatus_UNKNOWN, stacktrace.NewError("Failed to convert service status %v", serviceStatus)
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

func initStarlarkRun(
	starlarkRunRepository *starlark_run.StarlarkRunRepository,
	restartPolicy kurtosis_core_rpc_api_bindings.RestartPolicy,
) error {

	currentStarlarkRun, err := starlarkRunRepository.Get()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the current starlark run")
	}

	if currentStarlarkRun == nil {
		initialStarlarkRun := starlark_run.NewStarlarkRun(
			startosis_constants.PackageIdPlaceholderForStandaloneScript,
			"",
			"",
			defaultParallelism,
			startosis_constants.PlaceHolderMainFileForPlaceStandAloneScript,
			"",
			[]int32{},
			int32(restartPolicy.Number()),
			"",
		)

		if err = starlarkRunRepository.Save(initialStarlarkRun); err != nil {
			return stacktrace.Propagate(err, "An error occurred saving the initial starlark run")
		}
	}

	return nil
}
