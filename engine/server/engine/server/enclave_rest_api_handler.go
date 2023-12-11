package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_grpc"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_http"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/streaming"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/types"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/grpc-file-transfer/golang/grpc_file_streaming"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
	api "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/server/core_rest_api"
)

const (
	unknownStreamLength uint64 = 0 // set an unknown HTTP length header when using streaming/chunks
)

type enclaveRuntime struct {
	enclaveManager           enclave_manager.EnclaveManager
	remoteApiContainerClient map[string]rpc_api.ApiContainerServiceClient
	connectOnHostMachine     bool
	ctx                      context.Context
	lock                     sync.Mutex
	asyncStarlarkLogs        streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine]
}

func NewEnclaveRuntime(ctx context.Context, manager enclave_manager.EnclaveManager, asyncStarlarkLogs streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine], connectOnHostMachine bool) (*enclaveRuntime, error) {

	runtime := enclaveRuntime{
		enclaveManager:           manager,
		remoteApiContainerClient: map[string]rpc_api.ApiContainerServiceClient{},
		connectOnHostMachine:     connectOnHostMachine,
		ctx:                      ctx,
		asyncStarlarkLogs:        asyncStarlarkLogs,
		lock:                     sync.Mutex{},
	}

	err := runtime.refreshEnclaveConnections()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create enclave connections")
	}

	return &runtime, nil
}

// ===============================================================================================================
// ============================= Implementing  StrictServerInterface =============================================
// ===============================================================================================================

// (GET /enclaves/{enclave_identifier}/artifacts)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifacts(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}

	artifacts, err := (*apiContainerClient).ListFilesArtifactNamesAndUuids(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "fail to list fails file artifacts")
	}

	results := utils.MapList(
		artifacts.FileNamesAndUuids,
		func(x *rpc_api.FilesArtifactNameAndUuid) api_type.FileArtifactReference {
			return api_type.FileArtifactReference{
				Name: x.FileName,
				Uuid: x.FileUuid,
			}
		})

	return api.GetEnclavesEnclaveIdentifierArtifacts200JSONResponse(results), nil
}

// (POST /enclaves/{enclave_identifier}/artifacts/local-file)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierArtifactsLocalFile(ctx context.Context, request api.PostEnclavesEnclaveIdentifierArtifactsLocalFileRequestObject) (api.PostEnclavesEnclaveIdentifierArtifactsLocalFileResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsLocalFiledefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Uploading file artifact to enclave %s", enclaveIdentifier)

	uploadedArtifacts := map[string]api_type.FileArtifactUploadResult{}
	for {
		// Get next part (file) from the the multipart POST request
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		filename := part.FileName()

		client, err := (*apiContainerClient).UploadFilesArtifact(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Can't start file upload gRPC call with enclave %s", enclaveIdentifier)
		}
		clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, rpc_api.UploadFilesArtifactResponse](client)

		var result api_type.FileArtifactUploadResult
		uploadResult, err := clientStream.SendData(
			filename,
			part,
			unknownStreamLength,
			func(previousChunkHash string, contentChunk []byte) (*rpc_api.StreamedDataChunk, error) {
				return &rpc_api.StreamedDataChunk{
					Data:              contentChunk,
					PreviousChunkHash: previousChunkHash,
					Metadata: &rpc_api.DataChunkMetadata{
						Name: filename,
					},
				}, nil
			},
		)
		if err != nil {
			logrus.Errorf("Failed to upload file %s with error: %v", filename, err)
			response := api_type.ResponseInfo{
				Code:    http.StatusInternalServerError,
				Type:    api_type.ERROR,
				Message: fmt.Sprintf("Failed to upload file: %s", filename),
			}
			if err := result.FileArtifactUploadResult.FromResponseInfo(response); err != nil {
				return nil, stacktrace.Propagate(err, "Failed to write to enum http type with %T", response)
			}
			uploadedArtifacts[filename] = result
			continue
		}

		if uploadResult == nil {
			logrus.Warnf("File %s has been already uploaded", filename)
			response := api_type.ResponseInfo{
				Code:    http.StatusConflict,
				Type:    api_type.WARNING,
				Message: fmt.Sprintf("File %s has been already uploaded", filename),
			}
			if err := result.FileArtifactUploadResult.FromResponseInfo(response); err != nil {
				return nil, stacktrace.Propagate(err, "Failed to write to enum http type with %T", response)
			}
			uploadedArtifacts[filename] = result
			continue
		}

		artifactResponse := api_type.FileArtifactReference{
			Name: uploadResult.Name,
			Uuid: uploadResult.Uuid,
		}
		if err := result.FileArtifactUploadResult.FromFileArtifactReference(artifactResponse); err != nil {
			return nil, stacktrace.Propagate(err, "Failed to write to enum http type with %T", artifactResponse)
		}
		uploadedArtifacts[filename] = result
		uploadedArtifacts[filename] = result
	}

	return api.PostEnclavesEnclaveIdentifierArtifactsLocalFile200JSONResponse(uploadedArtifacts), nil
}

// (POST /enclaves/{enclave_identifier}/artifacts/remote-file)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierArtifactsRemoteFile(ctx context.Context, request api.PostEnclavesEnclaveIdentifierArtifactsRemoteFileRequestObject) (api.PostEnclavesEnclaveIdentifierArtifactsRemoteFileResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsRemoteFiledefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Uploading file artifact to enclave %s", enclaveIdentifier)

	storeWebFilesArtifactArgs := rpc_api.StoreWebFilesArtifactArgs{
		Url:  request.Body.Url,
		Name: request.Body.Name,
	}
	storedArtifact, err := (*apiContainerClient).StoreWebFilesArtifact(ctx, &storeWebFilesArtifactArgs)
	if err != nil {
		logrus.Errorf("Can't start file upload gRPC call with enclave %s, error: %s", enclaveIdentifier, err)
		return nil, stacktrace.NewError("Can't start file upload gRPC call with enclave %s", enclaveIdentifier)
	}

	artifactResponse := api_type.FileArtifactReference{
		Uuid: storedArtifact.Uuid,
		Name: request.Body.Name,
	}
	return api.PostEnclavesEnclaveIdentifierArtifactsRemoteFile200JSONResponse(artifactResponse), nil
}

// (POST /enclaves/{enclave_identifier}/artifacts/services/{service_identifier})
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifier(ctx context.Context, request api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierRequestObject) (api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	serviceIdentifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Storing file artifact from service %s on enclave %s", serviceIdentifier, enclaveIdentifier)

	storeWebFilesArtifactArgs := rpc_api.StoreFilesArtifactFromServiceArgs{
		ServiceIdentifier: serviceIdentifier,
		SourcePath:        request.Body.SourcePath,
		Name:              request.Body.Name,
	}
	storedArtifact, err := (*apiContainerClient).StoreFilesArtifactFromService(ctx, &storeWebFilesArtifactArgs)
	if err != nil {
		return nil, stacktrace.NewError("Can't start file upload gRPC call with enclave %s", enclaveIdentifier)
	}

	artifactResponse := api_type.FileArtifactReference{
		Uuid: storedArtifact.Uuid,
		Name: request.Body.Name,
	}
	return api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifier200JSONResponse(artifactResponse), nil
}

// (GET /enclaves/{enclave_identifier}/artifacts/{artifact_identifier})
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	artifactIdentifier := request.ArtifactIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Inspecting file artifact %s on enclave %s", artifactIdentifier, enclaveIdentifier)

	inspectFilesArtifactContentsRequest := rpc_api.InspectFilesArtifactContentsRequest{
		FileNamesAndUuid: &rpc_api.FilesArtifactNameAndUuid{
			FileName: artifactIdentifier,
			FileUuid: artifactIdentifier,
		},
	}
	storedArtifact, err := (*apiContainerClient).InspectFilesArtifactContents(ctx, &inspectFilesArtifactContentsRequest)
	if err != nil {
		return nil, stacktrace.NewError("Can't inspect artifact using gRPC call with enclave %s", enclaveIdentifier)
	}

	artifactContentList := utils.MapList(
		storedArtifact.FileDescriptions,
		func(x *rpc_api.FileArtifactContentsFileDescription) api_type.FileArtifactDescription {
			size := int64(x.Size)
			return api_type.FileArtifactDescription{
				Path:        x.Path,
				Size:        size,
				TextPreview: x.TextPreview,
			}
		})

	return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifier200JSONResponse(artifactContentList), nil
}

// (GET /enclaves/{enclave_identifier}/artifacts/{artifact_identifier}/download)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownload(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloadRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloadResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	artifactIdentifier := request.ArtifactIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloaddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Downloading file artifact %s from enclave %s", artifactIdentifier, enclaveIdentifier)

	downloadFilesArtifactArgs := rpc_api.DownloadFilesArtifactArgs{
		Identifier: artifactIdentifier,
	}
	client, err := (*apiContainerClient).DownloadFilesArtifact(ctx, &downloadFilesArtifactArgs)
	if err != nil {
		return nil, stacktrace.NewError("Can't start file download gRPC call with enclave %s", enclaveIdentifier)
	}

	clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, []byte](client)
	pipeReader := clientStream.PipeReader(
		artifactIdentifier,
		func(dataChunk *rpc_api.StreamedDataChunk) ([]byte, string, error) {
			return dataChunk.Data, dataChunk.PreviousChunkHash, nil
		},
	)

	response := api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownload200ApplicationoctetStreamResponse{
		Body:          pipeReader,
		ContentLength: int64(unknownStreamLength),
	}

	return response, nil
}

// (GET /enclaves/{enclave_identifier}/services)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServices(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesRequestObject) (api.GetEnclavesEnclaveIdentifierServicesResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about services enclave %s", enclaveIdentifier)

	serviceIds := utils.DerefWith(request.Params.Services, []string{})
	getServicesArgs := rpc_api.GetServicesArgs{
		ServiceIdentifiers: utils.NewMapFromList(serviceIds, func(x string) bool { return true }),
	}
	services, err := (*apiContainerClient).GetServices(ctx, &getServicesArgs)
	if err != nil {
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclaveIdentifier)
	}

	mappedServices := utils.MapMapValues(services.ServiceInfo, to_http.ToHttpServiceInfo)
	return api.GetEnclavesEnclaveIdentifierServices200JSONResponse(mappedServices), nil
}

// (GET /enclaves/{enclave_identifier}/services/history)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesHistory(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesHistoryRequestObject) (api.GetEnclavesEnclaveIdentifierServicesHistoryResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesHistorydefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Listing services from enclave %s", enclaveIdentifier)

	services, err := (*apiContainerClient).GetExistingAndHistoricalServiceIdentifiers(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclaveIdentifier)
	}

	response := utils.MapList(services.AllIdentifiers, func(service *rpc_api.ServiceIdentifiers) api_type.ServiceIdentifiers {
		return api_type.ServiceIdentifiers{
			ServiceUuid:   service.ServiceUuid,
			ShortenedUuid: service.ShortenedUuid,
			Name:          service.Name,
		}
	})

	return api.GetEnclavesEnclaveIdentifierServicesHistory200JSONResponse(response), nil
}

// (POST /enclaves/{enclave_identifier}/services/connection)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierServicesConnection(ctx context.Context, request api.PostEnclavesEnclaveIdentifierServicesConnectionRequestObject) (api.PostEnclavesEnclaveIdentifierServicesConnectionResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierServicesConnectiondefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Listing services from enclave %s", enclaveIdentifier)

	connectServicesArgs := rpc_api.ConnectServicesArgs{
		Connect: to_grpc.ToGrpcConnect(*request.Body),
	}
	_, err := (*apiContainerClient).ConnectServices(ctx, &connectServicesArgs)
	if err != nil {
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclaveIdentifier)
	}

	return api.PostEnclavesEnclaveIdentifierServicesConnection200Response{}, nil
}

// (GET /enclaves/{enclave_identifier}/services/{service_identifier})
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesServiceIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	serviceIdentifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about service %s from enclave %s", serviceIdentifier, enclaveIdentifier)

	getServicesArgs := rpc_api.GetServicesArgs{
		ServiceIdentifiers: map[string]bool{serviceIdentifier: true},
	}
	services, err := (*apiContainerClient).GetServices(ctx, &getServicesArgs)
	if err != nil {
		logrus.Errorf("Can't list services using gRPC call with enclave %s, error: %s", enclaveIdentifier, err)
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclaveIdentifier)
	}

	mappedServices := utils.MapMapValues(services.ServiceInfo, to_http.ToHttpServiceInfo)
	selectedService, found := mappedServices[serviceIdentifier]
	if !found {
		notFound := api_type.ResponseInfo{
			Code:    http.StatusNotFound,
			Type:    api_type.INFO,
			Message: fmt.Sprintf("service '%s' not found", serviceIdentifier),
		}
		return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierdefaultJSONResponse{
			Body:       notFound,
			StatusCode: int(notFound.Code),
		}, nil
	}
	return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifier200JSONResponse(selectedService), nil
}

// (POST /enclaves/{enclave_identifier}/services/{service_identifier}/command)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommand(ctx context.Context, request api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandRequestObject) (api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	serviceIdentifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommanddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about service %s from enclave %s", serviceIdentifier, enclaveIdentifier)

	execCommandArgs := rpc_api.ExecCommandArgs{
		ServiceIdentifier: serviceIdentifier,
		CommandArgs:       request.Body.CommandArgs,
	}
	execResult, err := (*apiContainerClient).ExecCommand(ctx, &execCommandArgs)
	if err != nil {
		return nil, stacktrace.NewError("Can't execute commands using gRPC call with enclave %s", enclaveIdentifier)
	}

	response := api_type.ExecCommandResult{
		ExitCode:  execResult.ExitCode,
		LogOutput: execResult.LogOutput,
	}
	return api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommand200JSONResponse(response), nil
}

// (GET /enclaves/{enclave_identifier}/services/{service_identifier}/endpoints/{port_number}/availability)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailability(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityRequestObject) (api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	serviceIdentifier := request.ServiceIdentifier
	portNumber := request.PortNumber
	endpointMethod := utils.DerefWith(request.Params.HttpMethod, api_type.GET)
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilitydefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about service %s from enclave %s", serviceIdentifier, enclaveIdentifier)

	castToUInt32 := func(v int32) uint32 { return uint32(v) }

	var err error
	switch endpointMethod {
	case api_type.GET:
		waitForHttpGetEndpointAvailabilityArgs := rpc_api.WaitForHttpGetEndpointAvailabilityArgs{
			ServiceIdentifier:        serviceIdentifier,
			Port:                     uint32(portNumber),
			Path:                     request.Params.Path,
			InitialDelayMilliseconds: utils.MapPointer(request.Params.InitialDelayMilliseconds, castToUInt32),
			Retries:                  utils.MapPointer(request.Params.Retries, castToUInt32),
			RetriesDelayMilliseconds: utils.MapPointer(request.Params.RetriesDelayMilliseconds, castToUInt32),
			BodyText:                 request.Params.ExpectedResponse,
		}
		_, err = (*apiContainerClient).WaitForHttpGetEndpointAvailability(ctx, &waitForHttpGetEndpointAvailabilityArgs)
	case api_type.POST:
		waitForHttpPostEndpointAvailabilityArgs := rpc_api.WaitForHttpPostEndpointAvailabilityArgs{
			ServiceIdentifier:        serviceIdentifier,
			Port:                     uint32(portNumber),
			Path:                     request.Params.Path,
			InitialDelayMilliseconds: utils.MapPointer(request.Params.InitialDelayMilliseconds, castToUInt32),
			Retries:                  utils.MapPointer(request.Params.Retries, castToUInt32),
			RetriesDelayMilliseconds: utils.MapPointer(request.Params.RetriesDelayMilliseconds, castToUInt32),
			BodyText:                 request.Params.ExpectedResponse,
			RequestBody:              request.Params.RequestBody,
		}
		_, err = (*apiContainerClient).WaitForHttpPostEndpointAvailability(ctx, &waitForHttpPostEndpointAvailabilityArgs)
	default:
		return nil, stacktrace.NewError("Undefined method for availability endpoint: %s", endpointMethod)
	}

	if err != nil {
		logrus.Errorf("Can't execute commands using gRPC call with enclave %s, error: %s", enclaveIdentifier, err)
		return nil, stacktrace.NewError("Can't execute commands using gRPC call with enclave %s", enclaveIdentifier)
	}
	return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailability200Response{}, nil

}

// (GET /enclaves/{enclave_identifier}/starlark)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierStarlark(ctx context.Context, request api.GetEnclavesEnclaveIdentifierStarlarkRequestObject) (api.GetEnclavesEnclaveIdentifierStarlarkResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierStarlarkdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about last Starlark run on enclave %s", enclaveIdentifier)

	starlarkResult, err := (*apiContainerClient).GetStarlarkRun(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.NewError("Can't get Starlark info using gRPC call with enclave %s", enclaveIdentifier)
	}

	flags := utils.MapList(starlarkResult.ExperimentalFeatures, to_http.ToHttpFeatureFlag)
	policy := to_http.ToHttpRestartPolicy(starlarkResult.RestartPolicy)
	response := api_type.StarlarkDescription{
		ExperimentalFeatures:   flags,
		MainFunctionName:       starlarkResult.MainFunctionName,
		PackageId:              starlarkResult.PackageId,
		Parallelism:            starlarkResult.Parallelism,
		RelativePathToMainFile: starlarkResult.RelativePathToMainFile,
		RestartPolicy:          policy,
		SerializedParams:       starlarkResult.SerializedParams,
		SerializedScript:       starlarkResult.SerializedScript,
	}

	return api.GetEnclavesEnclaveIdentifierStarlark200JSONResponse(response), nil
}

// (POST /enclaves/{enclave_identifier}/starlark/packages)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkPackages(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkPackagesRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkPackagesResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Upload Starlark package on enclave %s", enclaveIdentifier)

	failedUploads := []string{}
	successfulUploads := []string{}
	for {
		// Get next part (file) from the the multipart POST request
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		filename := part.FileName()
		client, err := (*apiContainerClient).UploadStarlarkPackage(ctx)
		if err != nil {
			return nil, stacktrace.NewError("Can't upload Starlark package using gRPC call with enclave %s", enclaveIdentifier)
		}
		clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, emptypb.Empty](client)

		_, err = clientStream.SendData(
			filename,
			part,
			unknownStreamLength,
			func(previousChunkHash string, contentChunk []byte) (*rpc_api.StreamedDataChunk, error) {
				return &rpc_api.StreamedDataChunk{
					Data:              contentChunk,
					PreviousChunkHash: previousChunkHash,
					Metadata: &rpc_api.DataChunkMetadata{
						Name: filename,
					},
				}, nil
			},
		)
		if err != nil {
			logrus.Warnf("Failed to upload file %s with error: %v", filename, err)
			failedUploads = append(failedUploads, filename)
		} else {
			successfulUploads = append(successfulUploads, filename)
		}
	}

	if len(successfulUploads) > 0 && len(failedUploads) > 0 {
		response := api_type.ResponseInfo{
			Code:    http.StatusOK,
			Type:    api_type.WARNING,
			Message: fmt.Sprintf("Failed to upload some of the files: %+v", failedUploads),
		}
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesdefaultJSONResponse{Body: response, StatusCode: int(response.Code)}, nil
	}

	if len(successfulUploads) == 0 && len(failedUploads) > 0 {
		response := api_type.ResponseInfo{
			Code:    http.StatusInternalServerError,
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("Failed to upload the files: %+v", failedUploads),
		}
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesdefaultJSONResponse{Body: response, StatusCode: int(response.Code)}, nil
	}

	return api.PostEnclavesEnclaveIdentifierStarlarkPackages200Response{}, nil
}

// (POST /enclaves/{enclave_identifier}/starlark/packages/{package_id})
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkPackagesPackageId(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}

	packageId := request.PackageId
	flags := utils.MapList(utils.DerefWith(request.Body.ExperimentalFeatures, []api_type.KurtosisFeatureFlag{}), to_grpc.ToGrpcFeatureFlag)
	// The gRPC always expect a JSON object even though it's marked as optional, so we need to default to `{}``
	jsonParams := utils.DerefWith(request.Body.Params, map[string]interface{}{})
	jsonBlob, err := json.Marshal(jsonParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "I'm actually panicking here. Re-serializing a already deserialized JSON parameter should never fail.")
	}
	jsonString := string(jsonBlob)

	logrus.Infof("Executing Starlark package `%s`", packageId)
	runStarlarkPackageArgs := rpc_api.RunStarlarkPackageArgs{
		PackageId:              packageId,
		SerializedParams:       &jsonString,
		DryRun:                 request.Body.DryRun,
		Parallelism:            request.Body.Parallelism,
		ClonePackage:           request.Body.ClonePackage,
		RelativePathToMainFile: request.Body.RelativePathToMainFile,
		MainFunctionName:       request.Body.MainFunctionName,
		ExperimentalFeatures:   flags,
		CloudInstanceId:        request.Body.CloudInstanceId,
		CloudUserId:            request.Body.CloudUserId,
		ImageDownloadMode:      utils.MapPointer(request.Body.ImageDownloadMode, to_grpc.ToGrpcImageDownloadMode),
		// Deprecated: If the package is local, it should have been uploaded with UploadStarlarkPackage prior to calling
		// RunStarlarkPackage. If the package is remote and must be cloned within the APIC, use the standalone boolean flag
		// clone_package below
		StarlarkPackageContent: nil,
	}

	ctxWithCancel, cancelCtxFunc := context.WithCancel(context.Background())
	stream, err := (*apiContainerClient).RunStarlarkPackage(ctxWithCancel, &runStarlarkPackageArgs)
	if err != nil {
		cancelCtxFunc()
		return nil, stacktrace.Propagate(err, "Can't run Starlark package using gRPC call with enclave %s", enclaveIdentifier)
	}

	asyncLogs := streaming.NewAsyncStarlarkLogs(cancelCtxFunc)
	go asyncLogs.AttachStream(stream)

	isAsyncRetrieval := utils.DerefWith(request.Params.RetrieveLogsAsync, true)
	var syncLogs api_type.StarlarkRunResponse_StarlarkExecutionLogs

	if !isAsyncRetrieval {
		allLogs, err := asyncLogs.WaitAndConsumeAll()
		if err != nil {
			cancelCtxFunc()
			return nil, stacktrace.Propagate(err, "Failed to consume all logs with enclave %s", enclaveIdentifier)
		}
		httpLogs, err := utils.MapListWithRefStopOnError(allLogs, to_http.ToHttpStarlarkRunResponseLine)
		if err != nil {
			cancelCtxFunc()
			return nil, stacktrace.Propagate(err, "Failed to convert values to http with enclave %s", enclaveIdentifier)
		}

		logs := utils.FilterListNils(httpLogs)
		if err := syncLogs.FromStarlarkRunLogs(logs); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", logs)
		}
	} else {
		asyncUuid := manager.asyncStarlarkLogs.Add(&asyncLogs)
		var asyncLogs api_type.AsyncStarlarkExecutionLogs
		asyncLogs.AsyncStarlarkExecutionLogs.StarlarkExecutionUuid = string(asyncUuid)
		if err := syncLogs.FromAsyncStarlarkExecutionLogs(asyncLogs); err != nil {
			return nil, stacktrace.Propagate(err, "failed to serialize %T", asyncLogs)
		}
	}

	response := api_type.StarlarkRunResponse{StarlarkExecutionLogs: &syncLogs}
	return api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageId200JSONResponse(response), nil
}

// (POST /enclaves/{enclave_identifier}/starlark/scripts)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkScripts(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkScriptsRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkScriptsResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.getApiClientOrResponseError(enclaveIdentifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkScriptsdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Run Starlark script on enclave %s", enclaveIdentifier)

	flags := utils.MapList(utils.DerefWith(request.Body.ExperimentalFeatures, []api_type.KurtosisFeatureFlag{}), to_grpc.ToGrpcFeatureFlag)
	jsonString, err := utils.MapPointerWithError(request.Body.Params, func(v map[string]interface{}) (string, error) {
		jsonBlob, err := json.Marshal(v)
		return string(jsonBlob), err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "I'm actually panicking here. Re-serializing a already deserialized JSON parameter should never fail.")
	}

	runStarlarkScriptArgs := rpc_api.RunStarlarkScriptArgs{
		SerializedScript:     request.Body.SerializedScript,
		SerializedParams:     jsonString,
		DryRun:               request.Body.DryRun,
		Parallelism:          request.Body.Parallelism,
		MainFunctionName:     request.Body.MainFunctionName,
		ExperimentalFeatures: flags,
		CloudInstanceId:      request.Body.CloudInstanceId,
		CloudUserId:          request.Body.CloudUserId,
		ImageDownloadMode:    utils.MapPointer(request.Body.ImageDownloadMode, to_grpc.ToGrpcImageDownloadMode),
	}

	ctxWithCancel, cancelCtxFunc := context.WithCancel(context.Background())
	stream, err := (*apiContainerClient).RunStarlarkScript(ctxWithCancel, &runStarlarkScriptArgs)
	if err != nil {
		cancelCtxFunc()
		logrus.Errorf("Can't run Starlark package using gRPC call with enclave %s, error: %s", enclaveIdentifier, err)
		return nil, stacktrace.Propagate(err, "Can't run Starlark script package using gRPC call with enclave %s", enclaveIdentifier)
	}

	asyncLogs := streaming.NewAsyncStarlarkLogs(cancelCtxFunc)
	go asyncLogs.AttachStream(stream)

	allLogs, err := asyncLogs.WaitAndConsumeAll()
	if err != nil {
		cancelCtxFunc()
		return nil, stacktrace.Propagate(err, "Failed to consume all logs with enclave %s", enclaveIdentifier)
	}
	httpLogs, err := utils.MapListWithRefStopOnError(allLogs, to_http.ToHttpStarlarkRunResponseLine)
	if err != nil {
		cancelCtxFunc()
		return nil, stacktrace.Propagate(err, "Failed to convert values to http with enclave %s", enclaveIdentifier)
	}

	logs := utils.FilterListNils(httpLogs)
	var syncLogs api_type.StarlarkRunResponse_StarlarkExecutionLogs
	if err := syncLogs.FromStarlarkRunLogs(logs); err != nil {
		return nil, stacktrace.Propagate(err, "failed to serialize %T", logs)
	}

	response := api_type.StarlarkRunResponse{StarlarkExecutionLogs: &syncLogs}
	return api.PostEnclavesEnclaveIdentifierStarlarkScripts200JSONResponse(response), nil
}

// ===============================================================================================================
// ===================================== Internal Functions =====================================================
// ===============================================================================================================

// GetGrpcClientConn returns a client conn dialed in to the local port
// It is the caller's responsibility to call resultClientConn.close()
func getGrpcClientConn(enclaveInfo types.EnclaveInfo, connectOnHostMachine bool) (resultClientConn *grpc.ClientConn, resultErr error) {
	enclaveAPIContainerInfo := enclaveInfo.ApiContainerInfo
	if enclaveAPIContainerInfo == nil {
		logrus.Infof("No API container info is available for enclave %s", enclaveInfo.EnclaveUuid)
		return nil, nil
	}
	apiContainerGrpcPort := enclaveAPIContainerInfo.GrpcPortInsideEnclave
	apiContainerIP := enclaveAPIContainerInfo.BridgeIpAddress

	enclaveAPIContainerHostMachineInfo := enclaveInfo.ApiContainerHostMachineInfo
	if connectOnHostMachine && enclaveAPIContainerHostMachineInfo == nil {
		logrus.Infof("No API container info is available for enclave %s", enclaveInfo.EnclaveUuid)
		return nil, nil
	}
	if connectOnHostMachine && (enclaveAPIContainerHostMachineInfo != nil) {
		apiContainerGrpcPort = enclaveAPIContainerHostMachineInfo.GrpcPortOnHostMachine
		apiContainerIP = enclaveAPIContainerHostMachineInfo.IpOnHostMachine
	}

	grpcServerAddress := fmt.Sprintf("%v:%v", apiContainerIP, apiContainerGrpcPort)
	grpcConnection, err := grpc.Dial(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to create a GRPC client connection on address '%v', but a non-nil error was returned", grpcServerAddress)
	}
	logrus.Debugf("Creating gRPC connection with enclave manager service on enclave %s on %s", enclaveInfo.EnclaveUuid, grpcServerAddress)
	return grpcConnection, nil
}

func (manager *enclaveRuntime) getApiClientOrResponseError(enclaveUuid string) (*rpc_api.ApiContainerServiceClient, *api_type.ResponseInfo) {
	client, err := manager.getGrpcClientForEnclaveUUID(enclaveUuid)
	if err != nil {
		return nil, &api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: "Couldn't retrieve connection with enclave",
			Code:    http.StatusInternalServerError,
		}
	}
	if client == nil {
		return nil, &api_type.ResponseInfo{
			Type:    api_type.INFO,
			Message: fmt.Sprintf("enclave '%s' not found", enclaveUuid),
			Code:    http.StatusNotFound,
		}
	}
	return client, nil
}

func (manager *enclaveRuntime) getGrpcClientForEnclaveUUID(enclaveUuid string) (*rpc_api.ApiContainerServiceClient, error) {
	err := manager.refreshEnclaveConnections()
	if err != nil {
		return nil, err
	}

	client, found := manager.remoteApiContainerClient[enclaveUuid]
	if !found {
		return nil, nil
	}

	return &client, nil
}

func (runtime *enclaveRuntime) refreshEnclaveConnections() error {
	runtime.lock.Lock()
	defer runtime.lock.Unlock()

	enclaves, err := runtime.enclaveManager.GetEnclaves(runtime.ctx)
	if err != nil {
		return stacktrace.Propagate(err, "failed to retrieve the list of enclaves")
	}

	// Clean up removed enclaves (or enclaves not in the list of enclaves anymore)
	for uuid := range runtime.remoteApiContainerClient {
		_, found := enclaves[uuid]
		if !found {
			delete(runtime.remoteApiContainerClient, uuid)
		}
	}

	// Add new enclaves - assuming enclaves properties (API container connection) are immutable
	for uuid, info := range enclaves {
		_, found := runtime.remoteApiContainerClient[uuid]
		if !found && info != nil {
			conn, err := getGrpcClientConn(*info, runtime.connectOnHostMachine)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to establish gRPC connection with enclave manager service on enclave %s", uuid)
			}
			if conn == nil {
				logrus.Warnf("Unavailable gRPC connection to enclave '%s', skipping it!", uuid)
				continue
			}
			apiContainerClient := rpc_api.NewApiContainerServiceClient(conn)
			runtime.remoteApiContainerClient[uuid] = apiContainerClient
		}
	}

	return nil
}
