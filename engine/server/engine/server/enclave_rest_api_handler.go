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
	api "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/core_rest_api"
)

type enclaveRuntime struct {
	enclaveManager           enclave_manager.EnclaveManager
	remoteApiContainerClient map[string]rpc_api.ApiContainerServiceClient
	connectOnHostMachine     bool
	ctx                      context.Context
	lock                     sync.Mutex
	asyncStarlarkLogs        streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine]
}

func (runtime enclaveRuntime) refreshEnclaveConnections() error {
	runtime.lock.Lock()
	defer runtime.lock.Unlock()

	enclaves, err := runtime.enclaveManager.GetEnclaves(runtime.ctx)
	if err != nil {
		return err
	}

	// Clean up removed enclaves
	for uuid := range runtime.remoteApiContainerClient {
		_, found := enclaves[uuid]
		if !found {
			delete(runtime.remoteApiContainerClient, uuid)
		}
	}

	// Add new enclaves - assuming enclaves properties (API container connection) are immutable
	for uuid, info := range enclaves {
		_, found := runtime.remoteApiContainerClient[uuid]
		if !found && (info != nil) {
			conn, err := getGrpcClientConn(*info, runtime.connectOnHostMachine)
			if err != nil {
				logrus.Errorf("Failed to establish gRPC connection with enclave manager service on enclave %s", uuid)
				return err
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

func NewEnclaveRuntime(ctx context.Context, manager enclave_manager.EnclaveManager, asyncStarlarkLogs streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine], connectOnHostMachine bool) (*enclaveRuntime, error) {

	runtime := enclaveRuntime{
		enclaveManager:           manager,
		remoteApiContainerClient: map[string]rpc_api.ApiContainerServiceClient{},
		connectOnHostMachine:     connectOnHostMachine,
		ctx:                      ctx,
		asyncStarlarkLogs:        asyncStarlarkLogs,
	}

	err := runtime.refreshEnclaveConnections()
	if err != nil {
		return nil, err
	}

	return &runtime, nil
}

// ===============================================================================================================
// ============================= Implementing  StrictServerInterface =============================================
// ===============================================================================================================

// (GET /enclaves/{enclave_identifier}/artifacts)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifacts(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}

	artifacts, err := (*apiContainerClient).ListFilesArtifactNamesAndUuids(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
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
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsLocalFiledefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Uploading file artifact to enclave %s", enclave_identifier)

	uploaded_artifacts := map[string]api_type.FileArtifactReference{}
	for {
		// Get next part (file) from the the multipart POST request
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		filename := part.FileName()

		client, err := (*apiContainerClient).UploadFilesArtifact(ctx)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Can't start file upload gRPC call with enclave %s", enclave_identifier)
		}
		clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, rpc_api.UploadFilesArtifactResponse](client)

		response, err := clientStream.SendData(
			filename,
			part,
			0, // Length unknown head of time
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

		// The response is nil when a file artifact with the same has already been uploaded
		// TODO (edgar) Is this the expected behavior? If so, we should be explicit about it.
		if response != nil {
			artifact_response := api_type.FileArtifactReference{
				Name: response.Name,
				Uuid: response.Uuid,
			}
			uploaded_artifacts[filename] = artifact_response
		}
	}

	return api.PostEnclavesEnclaveIdentifierArtifactsLocalFile200JSONResponse(uploaded_artifacts), nil
}

// (POST /enclaves/{enclave_identifier}/artifacts/remote-file)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierArtifactsRemoteFile(ctx context.Context, request api.PostEnclavesEnclaveIdentifierArtifactsRemoteFileRequestObject) (api.PostEnclavesEnclaveIdentifierArtifactsRemoteFileResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsRemoteFiledefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Uploading file artifact to enclave %s", enclave_identifier)

	storeWebFilesArtifactArgs := rpc_api.StoreWebFilesArtifactArgs{
		Url:  request.Body.Url,
		Name: request.Body.Name,
	}
	stored_artifact, err := (*apiContainerClient).StoreWebFilesArtifact(ctx, &storeWebFilesArtifactArgs)
	if err != nil {
		logrus.Errorf("Can't start file upload gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't start file upload gRPC call with enclave %s", enclave_identifier)
	}

	artifact_response := api_type.FileArtifactReference{
		Uuid: stored_artifact.Uuid,
		Name: request.Body.Name,
	}
	return api.PostEnclavesEnclaveIdentifierArtifactsRemoteFile200JSONResponse(artifact_response), nil
}

// (POST /enclaves/{enclave_identifier}/artifacts/services/{service_identifier})
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifier(ctx context.Context, request api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierRequestObject) (api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	service_identifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Storing file artifact from service %s on enclave %s", service_identifier, enclave_identifier)

	storeWebFilesArtifactArgs := rpc_api.StoreFilesArtifactFromServiceArgs{
		ServiceIdentifier: service_identifier,
		SourcePath:        request.Body.SourcePath,
		Name:              request.Body.Name,
	}
	stored_artifact, err := (*apiContainerClient).StoreFilesArtifactFromService(ctx, &storeWebFilesArtifactArgs)
	if err != nil {
		logrus.Errorf("Can't start file upload gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't start file upload gRPC call with enclave %s", enclave_identifier)
	}

	artifact_response := api_type.FileArtifactReference{
		Uuid: stored_artifact.Uuid,
		Name: request.Body.Name,
	}
	return api.PostEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifier200JSONResponse(artifact_response), nil
}

// (GET /enclaves/{enclave_identifier}/artifacts/{artifact_identifier})
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	artifact_identifier := request.ArtifactIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Inspecting file artifact %s on enclave %s", artifact_identifier, enclave_identifier)

	inspectFilesArtifactContentsRequest := rpc_api.InspectFilesArtifactContentsRequest{
		FileNamesAndUuid: &rpc_api.FilesArtifactNameAndUuid{
			FileName: artifact_identifier,
			FileUuid: artifact_identifier,
		},
	}
	stored_artifact, err := (*apiContainerClient).InspectFilesArtifactContents(ctx, &inspectFilesArtifactContentsRequest)
	if err != nil {
		logrus.Errorf("Can't inspect artifact using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't inspect artifact using gRPC call with enclave %s", enclave_identifier)
	}

	artifact_content_list := utils.MapList(
		stored_artifact.FileDescriptions,
		func(x *rpc_api.FileArtifactContentsFileDescription) api_type.FileArtifactDescription {
			size := int64(x.Size)
			return api_type.FileArtifactDescription{
				Path:        x.Path,
				Size:        size,
				TextPreview: x.TextPreview,
			}
		})

	return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifier200JSONResponse(artifact_content_list), nil
}

// (GET /enclaves/{enclave_identifier}/artifacts/{artifact_identifier}/download)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownload(ctx context.Context, request api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloadRequestObject) (api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloadResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	artifact_identifier := request.ArtifactIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownloaddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Downloading file artifact %s from enclave %s", artifact_identifier, enclave_identifier)

	downloadFilesArtifactArgs := rpc_api.DownloadFilesArtifactArgs{
		Identifier: artifact_identifier,
	}
	client, err := (*apiContainerClient).DownloadFilesArtifact(ctx, &downloadFilesArtifactArgs)
	if err != nil {
		logrus.Errorf("Can't start file download gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't start file download gRPC call with enclave %s", enclave_identifier)
	}

	clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, []byte](client)
	pipeReader := clientStream.PipeReader(
		artifact_identifier,
		func(dataChunk *rpc_api.StreamedDataChunk) ([]byte, string, error) {
			return dataChunk.Data, dataChunk.PreviousChunkHash, nil
		},
	)

	response := api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownload200ApplicationoctetStreamResponse{
		Body:          pipeReader,
		ContentLength: 0, // No file size is provided since we are streaming it directly
	}

	return api.GetEnclavesEnclaveIdentifierArtifactsArtifactIdentifierDownload200ApplicationoctetStreamResponse(response), nil
}

// (GET /enclaves/{enclave_identifier}/services)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServices(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesRequestObject) (api.GetEnclavesEnclaveIdentifierServicesResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about services enclave %s", enclave_identifier)

	service_ids := utils.DerefWith(request.Params.Services, []string{})
	getServicesArgs := rpc_api.GetServicesArgs{
		ServiceIdentifiers: utils.NewMapFromList(service_ids, func(x string) bool { return true }),
	}
	services, err := (*apiContainerClient).GetServices(ctx, &getServicesArgs)
	if err != nil {
		logrus.Errorf("Can't list services using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclave_identifier)
	}

	mapped_services := utils.MapMapValues(services.ServiceInfo, to_http.ToHttpServiceInfo)
	return api.GetEnclavesEnclaveIdentifierServices200JSONResponse(mapped_services), nil
}

// (GET /enclaves/{enclave_identifier}/services/history)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesHistory(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesHistoryRequestObject) (api.GetEnclavesEnclaveIdentifierServicesHistoryResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesHistorydefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Listing services from enclave %s", enclave_identifier)

	services, err := (*apiContainerClient).GetExistingAndHistoricalServiceIdentifiers(ctx, &emptypb.Empty{})
	if err != nil {
		logrus.Errorf("Can't list services using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclave_identifier)
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
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierServicesConnectiondefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Listing services from enclave %s", enclave_identifier)

	connectServicesArgs := rpc_api.ConnectServicesArgs{
		Connect: to_grpc.ToGrpcConnect(*request.Body),
	}
	_, err := (*apiContainerClient).ConnectServices(ctx, &connectServicesArgs)
	if err != nil {
		logrus.Errorf("Can't list services using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclave_identifier)
	}

	return api.PostEnclavesEnclaveIdentifierServicesConnection200Response{}, nil
}

// (GET /enclaves/{enclave_identifier}/services/{service_identifier})
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesServiceIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	service_identifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about service %s from enclave %s", service_identifier, enclave_identifier)

	getServicesArgs := rpc_api.GetServicesArgs{
		ServiceIdentifiers: map[string]bool{service_identifier: true},
	}
	services, err := (*apiContainerClient).GetServices(ctx, &getServicesArgs)
	if err != nil {
		logrus.Errorf("Can't list services using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't  list services using gRPC call with enclave %s", enclave_identifier)
	}

	mapped_services := utils.MapMapValues(services.ServiceInfo, to_http.ToHttpServiceInfo)
	selected_service, found := mapped_services[service_identifier]
	if !found {
		// TODO(edgar) add 404 return
		return nil, stacktrace.NewError("Service %s not found", service_identifier)
	}
	return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifier200JSONResponse(selected_service), nil
}

// (POST /enclaves/{enclave_identifier}/services/{service_identifier}/command)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommand(ctx context.Context, request api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandRequestObject) (api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	service_identifier := request.ServiceIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommanddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about service %s from enclave %s", service_identifier, enclave_identifier)

	execCommandArgs := rpc_api.ExecCommandArgs{
		ServiceIdentifier: service_identifier,
		CommandArgs:       request.Body.CommandArgs,
	}
	exec_result, err := (*apiContainerClient).ExecCommand(ctx, &execCommandArgs)
	if err != nil {
		logrus.Errorf("Can't execute commands using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't execute commands using gRPC call with enclave %s", enclave_identifier)
	}

	response := api_type.ExecCommandResult{
		ExitCode:  exec_result.ExitCode,
		LogOutput: exec_result.LogOutput,
	}
	return api.PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommand200JSONResponse(response), nil
}

// (GET /enclaves/{enclave_identifier}/services/{service_identifier}/endpoints/{port_number}/availability)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailability(ctx context.Context, request api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityRequestObject) (api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	service_identifier := request.ServiceIdentifier
	port_number := request.PortNumber
	endpoint_method := utils.DerefWith(request.Params.HttpMethod, api_type.GET)
	apiContainerClient, errConn := manager.GetGrpcClientForEnclaveUUID(enclave_identifier)
	if errConn != nil {
		return nil, errConn
	}
	logrus.Infof("Getting info about service %s from enclave %s", service_identifier, enclave_identifier)

	castToUInt32 := func(v int32) uint32 { return uint32(v) }

	var err error
	switch endpoint_method {
	case api_type.GET:
		waitForHttpGetEndpointAvailabilityArgs := rpc_api.WaitForHttpGetEndpointAvailabilityArgs{
			ServiceIdentifier:        service_identifier,
			Port:                     uint32(port_number),
			Path:                     request.Params.Path,
			InitialDelayMilliseconds: utils.MapPointer(request.Params.InitialDelayMilliseconds, castToUInt32),
			Retries:                  utils.MapPointer(request.Params.Retries, castToUInt32),
			RetriesDelayMilliseconds: utils.MapPointer(request.Params.RetriesDelayMilliseconds, castToUInt32),
			BodyText:                 request.Params.ExpectedResponse,
		}
		_, err = (*apiContainerClient).WaitForHttpGetEndpointAvailability(ctx, &waitForHttpGetEndpointAvailabilityArgs)
	case api_type.POST:
		waitForHttpPostEndpointAvailabilityArgs := rpc_api.WaitForHttpPostEndpointAvailabilityArgs{
			ServiceIdentifier:        service_identifier,
			Port:                     uint32(port_number),
			Path:                     request.Params.Path,
			InitialDelayMilliseconds: utils.MapPointer(request.Params.InitialDelayMilliseconds, castToUInt32),
			Retries:                  utils.MapPointer(request.Params.Retries, castToUInt32),
			RetriesDelayMilliseconds: utils.MapPointer(request.Params.RetriesDelayMilliseconds, castToUInt32),
			BodyText:                 request.Params.ExpectedResponse,
			RequestBody:              request.Params.RequestBody,
		}
		_, err = (*apiContainerClient).WaitForHttpPostEndpointAvailability(ctx, &waitForHttpPostEndpointAvailabilityArgs)
	default:
		return nil, stacktrace.NewError("Undefined method for availability endpoint: %s", endpoint_method)
	}

	if err != nil {
		logrus.Errorf("Can't execute commands using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't execute commands using gRPC call with enclave %s", enclave_identifier)
	}
	return api.GetEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailability200Response{}, nil

}

// (GET /enclaves/{enclave_identifier}/starlark)
func (manager *enclaveRuntime) GetEnclavesEnclaveIdentifierStarlark(ctx context.Context, request api.GetEnclavesEnclaveIdentifierStarlarkRequestObject) (api.GetEnclavesEnclaveIdentifierStarlarkResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.GetEnclavesEnclaveIdentifierStarlarkdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Getting info about last Starlark run on enclave %s", enclave_identifier)

	starlark_result, err := (*apiContainerClient).GetStarlarkRun(ctx, &emptypb.Empty{})
	if err != nil {
		logrus.Errorf("Can't get Starlark info using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.NewError("Can't get Starlark info using gRPC call with enclave %s", enclave_identifier)
	}

	flags := utils.MapList(starlark_result.ExperimentalFeatures, to_http.ToHttpFeatureFlag)
	policy := to_http.ToHttpRestartPolicy(starlark_result.RestartPolicy)
	response := api_type.StarlarkDescription{
		ExperimentalFeatures:   flags,
		MainFunctionName:       starlark_result.MainFunctionName,
		PackageId:              starlark_result.PackageId,
		Parallelism:            starlark_result.Parallelism,
		RelativePathToMainFile: starlark_result.RelativePathToMainFile,
		RestartPolicy:          policy,
		SerializedParams:       starlark_result.SerializedParams,
		SerializedScript:       starlark_result.SerializedScript,
	}

	return api.GetEnclavesEnclaveIdentifierStarlark200JSONResponse(response), nil
}

// (POST /enclaves/{enclave_identifier}/starlark/packages)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkPackages(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkPackagesRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkPackagesResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Upload Starlark package on enclave %s", enclave_identifier)

	for {
		// Get next part (file) from the the multipart POST request
		part, err := request.Body.NextPart()
		if err == io.EOF {
			break
		}
		filename := part.FileName()
		client, err := (*apiContainerClient).UploadStarlarkPackage(ctx)
		if err != nil {
			logrus.Errorf("Can't upload Starlark package using gRPC call with enclave %s, error: %s", enclave_identifier, err)
			return nil, stacktrace.NewError("Can't upload Starlark package using gRPC call with enclave %s", enclave_identifier)
		}
		clientStream := grpc_file_streaming.NewClientStream[rpc_api.StreamedDataChunk, emptypb.Empty](client)

		_, err = clientStream.SendData(
			filename,
			part,
			0, // Length unknown head of time
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
			// TODO(edgar) Should we stop on failure in case of multiple files? Should we return a list of succeed uploads?
			return nil, err
		}
	}

	return api.PostEnclavesEnclaveIdentifierStarlarkPackages200Response{}, nil
}

// (POST /enclaves/{enclave_identifier}/starlark/packages/{package_id})
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkPackagesPackageId(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIddefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Run Starlark package on enclave %s", enclave_identifier)

	package_id := request.PackageId
	flags := utils.MapList(utils.DerefWith(request.Body.ExperimentalFeatures, []api_type.KurtosisFeatureFlag{}), to_grpc.ToGrpcFeatureFlag)
	// The gRPC always expect a JSON object even though it's marked as optional, so we need to default to `{}``
	jsonParams := utils.DerefWith(request.Body.Params, map[string]interface{}{})
	jsonBlob, err := json.Marshal(jsonParams)
	if err != nil {
		panic("Failed to serialize parameters")
	}
	jsonString := string(jsonBlob)

	isRemote := utils.DerefWith(request.Body.Remote, false)
	logrus.Infof("Executing Starlark package `%s` with remote %t", package_id, isRemote)
	runStarlarkPackageArgs := rpc_api.RunStarlarkPackageArgs{
		PackageId:              package_id,
		StarlarkPackageContent: nil,
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
	}

	ctxWithCancel, cancelCtxFunc := context.WithCancel(context.Background())
	stream, err := (*apiContainerClient).RunStarlarkPackage(ctxWithCancel, &runStarlarkPackageArgs)
	if err != nil {
		cancelCtxFunc()
		logrus.Errorf("Can't run Starlark package using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.Propagate(err, "Can't run Starlark package using gRPC call with enclave %s", enclave_identifier)
	}

	asyncLogs := streaming.NewAsyncStarlarkLogs(cancelCtxFunc)
	go asyncLogs.AttachStream(stream)

	isAsyncRetrieval := utils.DerefWith(request.Params.RetrieveLogsAsync, true)
	var sync_logs api_type.StarlarkRunResponse_StarlarkExecutionLogs

	if !isAsyncRetrieval {
		logs := utils.MapList(asyncLogs.WaitAndConsumeAll(), to_http.ToHttpApiStarlarkRunResponseLine)
		sync_logs.FromStarlarkRunLogs(logs)
	} else {
		async_uuid := manager.asyncStarlarkLogs.Add(&asyncLogs)
		var async_logs api_type.AsyncStarlarkExecutionLogs
		async_logs.AsyncStarlarkExecutionLogs.StarlarkExecutionUuid = string(async_uuid)
		sync_logs.FromAsyncStarlarkExecutionLogs(async_logs)
	}

	response := api_type.StarlarkRunResponse{StarlarkExecutionLogs: &sync_logs}
	return api.PostEnclavesEnclaveIdentifierStarlarkPackagesPackageId200JSONResponse(response), nil
}

// (POST /enclaves/{enclave_identifier}/starlark/scripts)
func (manager *enclaveRuntime) PostEnclavesEnclaveIdentifierStarlarkScripts(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStarlarkScriptsRequestObject) (api.PostEnclavesEnclaveIdentifierStarlarkScriptsResponseObject, error) {
	enclave_identifier := request.EnclaveIdentifier
	apiContainerClient, responseErr := manager.GetApiClientOrResponseError(enclave_identifier)
	if responseErr != nil {
		return api.PostEnclavesEnclaveIdentifierStarlarkScriptsdefaultJSONResponse{Body: *responseErr, StatusCode: int(responseErr.Code)}, nil
	}
	logrus.Infof("Run Starlark script on enclave %s", enclave_identifier)

	flags := utils.MapList(utils.DerefWith(request.Body.ExperimentalFeatures, []api_type.KurtosisFeatureFlag{}), to_grpc.ToGrpcFeatureFlag)
	jsonString := utils.MapPointer(request.Body.Params, func(v map[string]interface{}) string {
		jsonBlob, err := json.Marshal(v)
		if err != nil {
			panic("Failed to serialize parsed JSON")
		}
		return string(jsonBlob)
	})

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
		logrus.Errorf("Can't run Starlark package using gRPC call with enclave %s, error: %s", enclave_identifier, err)
		return nil, stacktrace.Propagate(err, "Can't run Starlark script package using gRPC call with enclave %s", enclave_identifier)
	}

	asyncLogs := streaming.NewAsyncStarlarkLogs(cancelCtxFunc)
	go asyncLogs.AttachStream(stream)

	logs := utils.MapList(asyncLogs.WaitAndConsumeAll(), to_http.ToHttpApiStarlarkRunResponseLine)
	var sync_logs api_type.StarlarkRunResponse_StarlarkExecutionLogs
	sync_logs.FromStarlarkRunLogs(logs)

	response := api_type.StarlarkRunResponse{StarlarkExecutionLogs: &sync_logs}
	return api.PostEnclavesEnclaveIdentifierStarlarkScripts200JSONResponse(response), nil
}

// ===============================================================================================================
// ===============================================================================================================
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

func (manager enclaveRuntime) GetApiClientOrResponseError(enclave_uuid string) (*rpc_api.ApiContainerServiceClient, *api_type.ResponseInfo) {
	client, err := manager.GetGrpcClientForEnclaveUUID(enclave_uuid)
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
			Message: fmt.Sprintf("enclave '%s' not found", enclave_uuid),
			Code:    http.StatusNotFound,
		}
	}
	return client, nil
}

func (manager enclaveRuntime) GetGrpcClientForEnclaveUUID(enclave_uuid string) (*rpc_api.ApiContainerServiceClient, error) {
	err := manager.refreshEnclaveConnections()
	if err != nil {
		return nil, err
	}

	client, found := manager.remoteApiContainerClient[enclave_uuid]
	if !found {
		return nil, nil
	}

	return &client, nil
}