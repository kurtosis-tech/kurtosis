package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_http"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/types"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/sirupsen/logrus"

	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
	api "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/server/engine_rest_api"
)

type EngineRuntime struct {
	// The version tag of the engine server image, so it can report its own version
	ImageVersionTag string

	EnclaveManager *enclave_manager.EnclaveManager

	LogFileManager *log_file_manager.LogFileManager

	MetricsClient metrics_client.MetricsClient
}

// Delete Enclaves
// (DELETE /enclaves)
func (engine EngineRuntime) DeleteEnclaves(ctx context.Context, request api.DeleteEnclavesRequestObject) (api.DeleteEnclavesResponseObject, error) {
	removeAll := utils.DerefWith(request.Params.RemoveAll, false)
	removedEnclaveUuidsAndNames, err := engine.EnclaveManager.Clean(ctx, removeAll)
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred while cleaning enclaves")
		return api.DeleteEnclavesdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}
	if removeAll {
		if err = engine.LogFileManager.RemoveAllLogs(); err != nil {
			response := internalErrorResponseInfof(err, "An error occurred removing all logs")
			return api.DeleteEnclavesdefaultJSONResponse{
				Body:       response,
				StatusCode: int(response.Code),
			}, nil
		}
	}
	removedApiResponse := utils.MapList(removedEnclaveUuidsAndNames, to_http.ToHttpEnclaveNameAndUuid)
	return api.DeleteEnclaves200JSONResponse(api_type.DeletionSummary{RemovedEnclaveNameAndUuids: &removedApiResponse}), nil
}

// Get Enclaves
// (GET /enclaves)
func (engine EngineRuntime) GetEnclaves(ctx context.Context, request api.GetEnclavesRequestObject) (api.GetEnclavesResponseObject, error) {
	infoForEnclaves, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred getting info for enclaves")
		return api.GetEnclavesdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}
	response := utils.MapMapValues(infoForEnclaves, func(enclave *types.EnclaveInfo) api_type.EnclaveInfo { return to_http.ToHttpEnclaveInfo(*enclave) })
	return api.GetEnclaves200JSONResponse(response), nil
}

// Create Enclave
// (POST /enclaves)
func (engine EngineRuntime) PostEnclaves(ctx context.Context, request api.PostEnclavesRequestObject) (api.PostEnclavesResponseObject, error) {
	enclaveMode := utils.DerefWith(request.Body.Mode, api_type.TEST)
	enclaveName := request.Body.EnclaveName
	apicVersionTag := request.Body.ApiContainerVersionTag
	shouldApicRunInDebugMode := utils.DerefWith(request.Body.ShouldApicRunInDebugMode, api_type.False)

	if err := engine.MetricsClient.TrackCreateEnclave(enclaveName, subnetworkDisableBecauseItIsDeprecated); err != nil {
		logrus.Warn("An error occurred while logging the create enclave event")
	}

	logrus.Debugf("request: %+v", request)
	apiContainerLogLevel, err := logrus.ParseLevel(utils.DerefWith(request.Body.ApiContainerLogLevel, "INFO"))
	if err != nil {
		logrus.Infof("An error occurred parsing the log level string '%v':", request.Body.ApiContainerLogLevel)
		response := api_type.ResponseInfo{
			Code:    http.StatusBadRequest,
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("An error occurred parsing the log level string '%v':", request.Body.ApiContainerLogLevel),
		}
		return api.PostEnclavesdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	isProduction := false
	if enclaveMode == api_type.PRODUCTION {
		isProduction = true
	}

	enclaveInfo, err := engine.EnclaveManager.CreateEnclave(
		ctx,
		engine.ImageVersionTag,
		apicVersionTag,
		apiContainerLogLevel,
		enclaveName,
		isProduction,
		bool(shouldApicRunInDebugMode),
	)
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred creating new enclave with name '%v'", request.Body.EnclaveName)
		return api.PostEnclavesdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	response := to_http.ToHttpEnclaveInfo(*enclaveInfo)
	return api.PostEnclaves200JSONResponse(response), nil
}

// Get History Enclaves
// (GET /enclaves/history)
func (engine EngineRuntime) GetEnclavesHistory(ctx context.Context, request api.GetEnclavesHistoryRequestObject) (api.GetEnclavesHistoryResponseObject, error) {
	allIdentifiers, err := engine.EnclaveManager.GetExistingAndHistoricalEnclaveIdentifiers()
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred while fetching enclave identifiers")
		return api.GetEnclavesHistorydefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	identifiersMapApi := utils.MapList(allIdentifiers, to_http.ToHttpEnclaveIdentifiers)
	return api.GetEnclavesHistory200JSONResponse(identifiersMapApi), nil
}

// Destroy Enclave
// (DELETE /enclaves/{enclave_identifier})
func (engine EngineRuntime) DeleteEnclavesEnclaveIdentifier(ctx context.Context, request api.DeleteEnclavesEnclaveIdentifierRequestObject) (api.DeleteEnclavesEnclaveIdentifierResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier

	if err := engine.MetricsClient.TrackDestroyEnclave(enclaveIdentifier); err != nil {
		logrus.Warnf("An error occurred while logging the destroy enclave event for enclave '%v'", enclaveIdentifier)
	}

	if err := engine.EnclaveManager.DestroyEnclave(ctx, enclaveIdentifier); err != nil {
		response := internalErrorResponseInfof(err, "An error occurred destroying enclave with identifier '%v':", enclaveIdentifier)
		return api.DeleteEnclavesEnclaveIdentifierdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	return api.DeleteEnclavesEnclaveIdentifier200Response{}, nil
}

// Get Enclave Info
// (GET /enclaves/{enclave_identifier})
func (engine EngineRuntime) GetEnclavesEnclaveIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierResponseObject, error) {
	infoForEnclaves, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred getting info for enclaves")
		return api.GetEnclavesEnclaveIdentifierdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	info, found := infoForEnclaves[request.EnclaveIdentifier]
	if !found {
		response := enclaveNotFoundResponseInfo(request.EnclaveIdentifier)
		return api.GetEnclavesEnclaveIdentifierdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	response := to_http.ToHttpEnclaveInfo(*info)
	return api.GetEnclavesEnclaveIdentifier200JSONResponse(response), nil
}

// Get enclave status
// (GET /enclaves/{enclave_identifier}/status)
func (engine EngineRuntime) GetEnclavesEnclaveIdentifierStatus(ctx context.Context, request api.GetEnclavesEnclaveIdentifierStatusRequestObject) (api.GetEnclavesEnclaveIdentifierStatusResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	enclaveList, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		response := internalErrorResponseInfof(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
		return api.GetEnclavesEnclaveIdentifierStatusdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	info, found := enclaveList[enclaveIdentifier]
	if !found {
		response := enclaveNotFoundResponseInfo(request.EnclaveIdentifier)
		return api.GetEnclavesEnclaveIdentifierStatusdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}

	return api.GetEnclavesEnclaveIdentifierStatus200JSONResponse(to_http.ToHttpEnclaveStatus(info.EnclaveStatus)), nil
}

// Set enclave status
// (POST /enclaves/{enclave_identifier}/status)
func (engine EngineRuntime) PostEnclavesEnclaveIdentifierStatus(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStatusRequestObject) (api.PostEnclavesEnclaveIdentifierStatusResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	targetState := request.Body

	switch *targetState {
	case api_type.STOP:
		if err := engine.MetricsClient.TrackStopEnclave(enclaveIdentifier); err != nil {
			logrus.Warnf("An error occurred while logging the stop enclave event for enclave '%v'", enclaveIdentifier)
		}

		if err := engine.EnclaveManager.StopEnclave(ctx, enclaveIdentifier); err != nil {
			response := internalErrorResponseInfof(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
			return api.PostEnclavesEnclaveIdentifierStatusdefaultJSONResponse{
				Body:       response,
				StatusCode: int(response.Code),
			}, nil
		}

		return api.PostEnclavesEnclaveIdentifierStatus200Response{}, nil

	default:
		logrus.Infof("Unsupported target state: '%s'", string(*targetState))
		response := api_type.ResponseInfo{
			Code:    http.StatusBadRequest,
			Type:    api_type.WARNING,
			Message: fmt.Sprintf("Unsupported target state: '%s'", string(*targetState)),
		}
		return api.PostEnclavesEnclaveIdentifierStatusdefaultJSONResponse{
			Body:       response,
			StatusCode: int(response.Code),
		}, nil
	}
}

// Get Engine Info
// (GET /engine/info)
func (engine EngineRuntime) GetEngineInfo(ctx context.Context, request api.GetEngineInfoRequestObject) (api.GetEngineInfoResponseObject, error) {
	result := api_type.EngineInfo{EngineVersion: engine.ImageVersionTag}
	return api.GetEngineInfo200JSONResponse(result), nil
}

// ===============================================================================================================
// ===================================== Internal Functions =====================================================
// ===============================================================================================================

func enclaveNotFoundResponseInfo(enclaveIdentifier string) api_type.ResponseInfo {
	logrus.Infof("Enclave '%s' not found.", enclaveIdentifier)
	return api_type.ResponseInfo{
		Code:    http.StatusNotFound,
		Type:    api_type.INFO,
		Message: fmt.Sprintf("Enclave '%s' not found.", enclaveIdentifier),
	}
}

func internalErrorResponseInfof(err error, format string, args ...interface{}) api_type.ResponseInfo {
	logrus.WithField("stacktrace", fmt.Sprintf("%+v", err)).WithError(err).Errorf(format, args...)
	return api_type.ResponseInfo{
		Code:    http.StatusInternalServerError,
		Type:    api_type.ERROR,
		Message: fmt.Sprintf(format, args...),
	}
}
