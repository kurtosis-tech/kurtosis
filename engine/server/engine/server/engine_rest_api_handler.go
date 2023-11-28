package server

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/types"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"

	api "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rest_api_bindings"
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
	removedEnclaveUuidsAndNames, err := engine.EnclaveManager.Clean(ctx, *request.Params.RemoveAll)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while cleaning enclaves")
	}
	if *request.Params.RemoveAll {
		if err = engine.LogFileManager.RemoveAllLogs(); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred removing all logs.")
		}
	}
	removedApiResponse := utils.MapList(removedEnclaveUuidsAndNames, toHttpApiEnclaveNameAndUuid)
	return api.DeleteEnclaves200JSONResponse(api.DeletionSummary{RemovedEnclaveNameAndUuids: &removedApiResponse}), nil
}

// Get Enclaves
// (GET /enclaves)
func (engine EngineRuntime) GetEnclaves(ctx context.Context, request api.GetEnclavesRequestObject) (api.GetEnclavesResponseObject, error) {
	infoForEnclaves, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	response := utils.MapMapValues(infoForEnclaves, func(enclave *types.EnclaveInfo) api.EnclaveInfo { return toHttpApiEnclaveInfo(*enclave) })
	return api.GetEnclaves200JSONResponse(response), nil
}

// Create Enclave
// (POST /enclaves)
func (engine EngineRuntime) PostEnclaves(ctx context.Context, request api.PostEnclavesRequestObject) (api.PostEnclavesResponseObject, error) {
	if err := engine.MetricsClient.TrackCreateEnclave(*request.Body.EnclaveName, subnetworkDisableBecauseItIsDeprecated); err != nil {
		logrus.Warn("An error occurred while logging the create enclave event")
	}

	logrus.Debugf("request: %+v", request)
	apiContainerLogLevel, err := logrus.ParseLevel(*request.Body.ApiContainerLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", request.Body.ApiContainerLogLevel)
	}

	isProduction := false
	if *request.Body.Mode == api.PRODUCTION {
		isProduction = true
	}

	enclaveInfo, err := engine.EnclaveManager.CreateEnclave(
		ctx,
		engine.ImageVersionTag,
		*request.Body.ApiContainerVersionTag,
		apiContainerLogLevel,
		*request.Body.EnclaveName,
		isProduction,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with name '%v'", request.Body.EnclaveName)
	}

	response := toHttpApiEnclaveInfo(*enclaveInfo)
	return api.PostEnclaves200JSONResponse(response), nil
}

// Get Historical Enclaves
// (GET /enclaves/historical)
func (engine EngineRuntime) GetEnclavesHistorical(ctx context.Context, request api.GetEnclavesHistoricalRequestObject) (api.GetEnclavesHistoricalResponseObject, error) {
	allIdentifiers, err := engine.EnclaveManager.GetExistingAndHistoricalEnclaveIdentifiers()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching enclave identifiers")
	}
	identifiers_map_api := utils.MapList(allIdentifiers, toHttpApiEnclaveIdentifiers)
	return api.GetEnclavesHistorical200JSONResponse(identifiers_map_api), nil
}

// Destroy Enclave
// (DELETE /enclaves/{enclave_identifier})
func (engine EngineRuntime) DeleteEnclavesEnclaveIdentifier(ctx context.Context, request api.DeleteEnclavesEnclaveIdentifierRequestObject) (api.DeleteEnclavesEnclaveIdentifierResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier

	if err := engine.MetricsClient.TrackDestroyEnclave(enclaveIdentifier); err != nil {
		logrus.Warnf("An error occurred while logging the destroy enclave event for enclave '%v'", enclaveIdentifier)
	}

	if err := engine.EnclaveManager.DestroyEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying enclave with identifier '%v':", enclaveIdentifier)
	}
	return api.DeleteEnclavesEnclaveIdentifier200Response{}, nil
}

// Get Enclave Info
// (GET /enclaves/{enclave_identifier})
func (engine EngineRuntime) GetEnclavesEnclaveIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierResponseObject, error) {
	infoForEnclaves, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	info, found := infoForEnclaves[request.EnclaveIdentifier]
	if !found {
		notFoundErr := stacktrace.NewError("Enclave '%s' not found.", request.EnclaveIdentifier)
		return nil, notFoundErr
	}
	response := toHttpApiEnclaveInfo(*info)
	return api.GetEnclavesEnclaveIdentifier200JSONResponse(response), nil
}

// Get enclave status
// (GET /enclaves/{enclave_identifier}/status)
func (engine EngineRuntime) GetEnclavesEnclaveIdentifierStatus(ctx context.Context, request api.GetEnclavesEnclaveIdentifierStatusRequestObject) (api.GetEnclavesEnclaveIdentifierStatusResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	enclaveList, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
	}
	info, found := enclaveList[enclaveIdentifier]
	if !found {
		err := stacktrace.NewError("Enclave not found: '%s'", enclaveIdentifier)
		return nil, err
	}

	return api.GetEnclavesEnclaveIdentifierStatus200JSONResponse(toHttpEnclaveContainersStatus(info.EnclaveContainersStatus)), nil
}

// Set enclave status
// (POST /enclaves/{enclave_identifier}/status)
func (engine EngineRuntime) PostEnclavesEnclaveIdentifierStatus(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStatusRequestObject) (api.PostEnclavesEnclaveIdentifierStatusResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	targetState := request.Body

	switch *targetState {
	case api.STOP:
		if err := engine.MetricsClient.TrackStopEnclave(enclaveIdentifier); err != nil {
			logrus.Warnf("An error occurred while logging the stop enclave event for enclave '%v'", enclaveIdentifier)
		}

		if err := engine.EnclaveManager.StopEnclave(ctx, enclaveIdentifier); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
		}
		return api.PostEnclavesEnclaveIdentifierStatus200Response{}, nil
	default:
		err := stacktrace.NewError("Unsupported target state: '%s'", string(*targetState))
		return nil, err
	}

}

// Get Engine Info
// (GET /engine/info)
func (engine EngineRuntime) GetEngineInfo(ctx context.Context, request api.GetEngineInfoRequestObject) (api.GetEngineInfoResponseObject, error) {
	result := api.EngineInfo{EngineVersion: engine.ImageVersionTag}
	return api.GetEngineInfo200JSONResponse(result), nil
}

// =============================================================================================================================================
// ============================================== Helper Functions =============================================================================
// =============================================================================================================================================

func toHttpApiEnclaveContainersStatus(status types.EnclaveContainersStatus) api.EnclaveContainersStatus {
	switch status {
	case types.EnclaveContainersStatus_EMPTY:
		return api.EnclaveContainersStatusEMPTY
	case types.EnclaveContainersStatus_STOPPED:
		return api.EnclaveContainersStatusSTOPPED
	case types.EnclaveContainersStatus_RUNNING:
		return api.EnclaveContainersStatusRUNNING
	default:
		panic(fmt.Sprintf("Undefined mapping of value: %s", status))
	}
}

func toHttpApiContainerStatus(status types.ContainerStatus) api.ApiContainerStatus {
	switch status {
	case types.ContainerStatus_NONEXISTENT:
		return api.ApiContainerStatusNONEXISTENT
	case types.ContainerStatus_STOPPED:
		return api.ApiContainerStatusSTOPPED
	case types.ContainerStatus_RUNNING:
		return api.ApiContainerStatusRUNNING
	default:
		panic(fmt.Sprintf("Undefined mapping of value: %s", status))
	}
}

func toHttpApiEnclaveAPIContainerInfo(info types.EnclaveAPIContainerInfo) api.EnclaveAPIContainerInfo {
	port := int(info.GrpcPortInsideEnclave)
	return api.EnclaveAPIContainerInfo{
		ContainerId:           info.ContainerId,
		IpInsideEnclave:       info.IpInsideEnclave,
		GrpcPortInsideEnclave: port,
		BridgeIpAddress:       info.BridgeIpAddress,
	}
}

func toHttpApiApiContainerHostMachineInfo(info types.EnclaveAPIContainerHostMachineInfo) api.EnclaveAPIContainerHostMachineInfo {
	port := int(info.GrpcPortOnHostMachine)
	return api.EnclaveAPIContainerHostMachineInfo{
		IpOnHostMachine:       info.IpOnHostMachine,
		GrpcPortOnHostMachine: port,
	}
}

func toHttpApiEnclaveMode(mode types.EnclaveMode) api.EnclaveMode {
	switch mode {
	case types.EnclaveMode_PRODUCTION:
		return api.PRODUCTION
	case types.EnclaveMode_TEST:
		return api.TEST
	default:
		panic(fmt.Sprintf("Undefined mapping of value: %s", mode))
	}
}

func toHttpApiEnclaveInfo(info types.EnclaveInfo) api.EnclaveInfo {
	return api.EnclaveInfo{
		EnclaveUuid:                 info.EnclaveUuid,
		ShortenedUuid:               info.ShortenedUuid,
		Name:                        info.Name,
		ContainersStatus:            toHttpApiEnclaveContainersStatus(info.EnclaveContainersStatus),
		ApiContainerStatus:          toHttpApiContainerStatus(info.ApiContainerStatus),
		ApiContainerInfo:            utils.MapPointer(info.ApiContainerInfo, toHttpApiEnclaveAPIContainerInfo),
		ApiContainerHostMachineInfo: utils.MapPointer(info.ApiContainerHostMachineInfo, toHttpApiApiContainerHostMachineInfo),
		CreationTime:                info.CreationTime,
		Mode:                        toHttpApiEnclaveMode(info.Mode),
	}
}

func toHttpApiEnclaveIdentifiers(identifier *types.EnclaveIdentifiers) api.EnclaveIdentifiers {
	return api.EnclaveIdentifiers{
		EnclaveUuid:   identifier.EnclaveUuid,
		Name:          identifier.Name,
		ShortenedUuid: identifier.ShortenedUuid,
	}
}

func toHttpApiEnclaveNameAndUuid(identifier *types.EnclaveNameAndUuid) api.EnclaveNameAndUuid {
	return api.EnclaveNameAndUuid{
		Uuid: identifier.Uuid,
		Name: identifier.Name,
	}
}

func toHttpEnclaveContainersStatus(status types.EnclaveContainersStatus) api.EnclaveContainersStatus {
	switch status {
	case types.EnclaveContainersStatus_STOPPED:
		return api.EnclaveContainersStatusSTOPPED
	case types.EnclaveContainersStatus_RUNNING:
		return api.EnclaveContainersStatusRUNNING
	case types.EnclaveContainersStatus_EMPTY:
		return api.EnclaveContainersStatusEMPTY
	default:
		panic(fmt.Sprintf("Undefined mapping of value: %s", status))
	}
}
