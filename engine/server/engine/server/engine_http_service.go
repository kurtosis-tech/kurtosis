package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/logline"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/types"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	api "github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_http_api_bindings"
)

type EngineRuntime struct {
	// The version tag of the engine server image, so it can report its own version
	ImageVersionTag string

	EnclaveManager *enclave_manager.EnclaveManager

	// The protected user ID for metrics analytics purpose
	MetricsUserID string

	// User consent to send metrics
	DidUserAcceptSendingMetrics bool

	// The clients for consuming container logs from the logs' database server

	// per week pulls logs from enclaves created post log retention feature
	PerWeekLogsDatabaseClient centralized_logs.LogsDatabaseClient

	// per file pulls logs from enclaves created pre log retention feature
	// TODO: remove once users are fully migrated to log retention/new log schema
	PerFileLogsDatabaseClient centralized_logs.LogsDatabaseClient

	LogFileManager *log_file_manager.LogFileManager

	MetricsClient metrics_client.MetricsClient
}

type Error struct {
}

func (error Error) Error() string {
	return "Not Implemented :("
}

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

func toHttpApiEnclaveAPIContainerInfo(info *types.EnclaveAPIContainerInfo) api.EnclaveAPIContainerInfo {
	if info == nil {
		return api.EnclaveAPIContainerInfo{}
	}
	port := int(info.GrpcPortInsideEnclave)
	return api.EnclaveAPIContainerInfo{
		ContainerId:           &info.ContainerId,
		IpInsideEnclave:       &info.IpInsideEnclave,
		GrpcPortInsideEnclave: &port,
		BridgeIpAddress:       &info.BridgeIpAddress,
	}
}

func toHttpApiApiContainerHostMachineInfo(info *types.EnclaveAPIContainerHostMachineInfo) api.EnclaveAPIContainerHostMachineInfo {
	if info == nil {
		return api.EnclaveAPIContainerHostMachineInfo{}
	}
	port := int(info.GrpcPortOnHostMachine)
	return api.EnclaveAPIContainerHostMachineInfo{
		IpOnHostMachine:       &info.IpOnHostMachine,
		GrpcPortOnHostMachine: &port,
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
	containerInfo := toHttpApiEnclaveAPIContainerInfo(info.ApiContainerInfo)
	apiHostMachine := toHttpApiApiContainerHostMachineInfo(info.ApiContainerHostMachineInfo)
	containersStatus := toHttpApiEnclaveContainersStatus(info.EnclaveContainersStatus)
	apiContainerStatus := toHttpApiContainerStatus(info.ApiContainerStatus)
	mode := toHttpApiEnclaveMode(info.Mode)
	return api.EnclaveInfo{
		EnclaveUuid:                 &info.EnclaveUuid,
		ShortenedUuid:               &info.ShortenedUuid,
		Name:                        &info.Name,
		ContainersStatus:            &containersStatus,
		ApiContainerStatus:          &apiContainerStatus,
		ApiContainerInfo:            &containerInfo,
		ApiContainerHostMachineInfo: &apiHostMachine,
		CreationTime:                &info.CreationTime,
		Mode:                        &mode,
	}
}

func toHttpApiEnclaveInfos(infos map[string]*types.EnclaveInfo) map[string]api.EnclaveInfo {
	info_map := make(map[string]api.EnclaveInfo)
	for key, info := range infos {
		if info != nil {
			grpc_info := toHttpApiEnclaveInfo(*info)
			info_map[key] = grpc_info
		}
	}
	return info_map
}

func toHttpApiEnclaveIdentifiers(identifier *types.EnclaveIdentifiers) api.EnclaveIdentifiers {
	return api.EnclaveIdentifiers{
		EnclaveUuid:   &identifier.EnclaveUuid,
		Name:          &identifier.Name,
		ShortenedUuid: &identifier.ShortenedUuid,
	}
}

func toHttpApiEnclaveNameAndUuid(identifier *types.EnclaveNameAndUuid) api.EnclaveNameAndUuid {
	return api.EnclaveNameAndUuid{
		Uuid: &identifier.Uuid,
		Name: &identifier.Name,
	}
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
	return api.DeleteEnclaves200JSONResponse(api.DeleteResponse{RemovedEnclaveNameAndUuids: &removedApiResponse}), nil
}

// Get Enclaves
// (GET /enclaves)
func (engine EngineRuntime) GetEnclaves(ctx context.Context, request api.GetEnclavesRequestObject) (api.GetEnclavesResponseObject, error) {
	infoForEnclaves, err := engine.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting info for enclaves")
	}
	info_map_http := toHttpApiEnclaveInfos(infoForEnclaves)
	response := api.GetEnclavesResponse{EnclaveInfo: &info_map_http}
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

	grpcEnclaveInfo := toHttpApiEnclaveInfo(*enclaveInfo)
	response := api.CreateEnclaveResponse{
		EnclaveInfo: &grpcEnclaveInfo,
	}

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
	response := api.GetExistingAndHistoricalEnclaveIdentifiersResponse{AllIdentifiers: &identifiers_map_api}
	return api.GetEnclavesHistorical200JSONResponse(response), nil
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
	return api.DeleteEnclavesEnclaveIdentifier200JSONResponse{}, nil
}

// Get Enclave Info
// (GET /enclaves/{enclave_identifier})
func (engine EngineRuntime) GetEnclavesEnclaveIdentifier(ctx context.Context, request api.GetEnclavesEnclaveIdentifierRequestObject) (api.GetEnclavesEnclaveIdentifierResponseObject, error) {
	return nil, Error{}
}

// Get Service Logs
// (POST /enclaves/{enclave_identifier}/logs)
func (engine EngineRuntime) PostEnclavesEnclaveIdentifierLogs(ctx context.Context, request api.PostEnclavesEnclaveIdentifierLogsRequestObject) (api.PostEnclavesEnclaveIdentifierLogsResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier
	enclaveUuid, err := engine.EnclaveManager.GetEnclaveUuidForEnclaveIdentifier(context.Background(), enclaveIdentifier)

	contextWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	if err != nil {
		logrus.Errorf("An error occurred while fetching uuid for enclave '%v'. This could happen if the enclave has been deleted. Treating it as UUID", enclaveIdentifier)
		enclaveUuid = enclave.EnclaveUUID(enclaveIdentifier)
	}
	serviceUuidStrSet := *request.Body.ServiceUuidSet
	requestedServiceUuids := make(map[user_service.ServiceUUID]bool, len(serviceUuidStrSet))
	shouldFollowLogs := *request.Body.FollowLogs
	shouldReturnAllLogs := *request.Body.ReturnAllLogs
	numLogLines := *request.Body.NumLogLines

	for _, serviceUuidStr := range serviceUuidStrSet {
		serviceUuid := user_service.ServiceUUID(serviceUuidStr)
		requestedServiceUuids[serviceUuid] = true
	}

	if engine.PerWeekLogsDatabaseClient == nil || engine.PerFileLogsDatabaseClient == nil {
		return nil, stacktrace.NewError("It's not possible to return service logs because there is no logs database client; this is bug in Kurtosis")
	}

	var (
		serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
		errChan                      chan error
		cancelCtxFunc                func()
	)

	notFoundServiceUuids, err := engine.reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(contextWithCancel, enclaveUuid, requestedServiceUuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reporting missing user service UUIDs for enclave '%v' and requested service UUIDs '%+v'", enclaveUuid, requestedServiceUuids)
	}

	conjunctiveLogLineFilters, err := newConjunctiveLogLineFiltersFromHttpLogLineFilters(*request.Body.ConjunctiveFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the conjunctive log line filters from the GRPC's conjunctive log line filters '%+v'", request.Body.ConjunctiveFilters)
	}

	// get enclave creation time to determine strategy to pull logs
	enclaveCreationTime, err := engine.getEnclaveCreationTime(ctx, enclaveUuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while trying to get the enclave creation time to determine how to pull logs.")
	}
	logsDatabaseClient := engine.getLogsDatabaseClient(enclaveCreationTime)

	serviceLogsByServiceUuidChan, errChan, cancelCtxFunc, err = logsDatabaseClient.StreamUserServiceLogs(
		contextWithCancel,
		enclaveUuid,
		requestedServiceUuids,
		conjunctiveLogLineFilters,
		shouldFollowLogs,
		shouldReturnAllLogs,
		uint32(numLogLines))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred streaming service logs for UUIDs '%+v' in enclave with ID '%v' using filters '%v+' "+
				"and with should follow logs value as '%v'",
			requestedServiceUuids,
			enclaveUuid,
			conjunctiveLogLineFilters,
			shouldFollowLogs,
		)
	}
	defer func() {
		if cancelCtxFunc != nil {
			cancelCtxFunc()
		}
	}()

	streamingResponse := StreamLogs{
		serviceLogsByServiceUuidChan: serviceLogsByServiceUuidChan,
		errChan:                      errChan,
		missingServicesUUID:          notFoundServiceUuids,
		cancelCtxFunc:                cancelCtxFunc,
		contextWithCancel:            contextWithCancel,
	}

	return streamingResponse, nil

}

// Stop Enclave
// (POST /enclaves/{enclave_identifier}/stop)
func (engine EngineRuntime) PostEnclavesEnclaveIdentifierStop(ctx context.Context, request api.PostEnclavesEnclaveIdentifierStopRequestObject) (api.PostEnclavesEnclaveIdentifierStopResponseObject, error) {
	enclaveIdentifier := request.EnclaveIdentifier

	if err := engine.MetricsClient.TrackStopEnclave(enclaveIdentifier); err != nil {
		logrus.Warnf("An error occurred while logging the stop enclave event for enclave '%v'", enclaveIdentifier)
	}

	if err := engine.EnclaveManager.StopEnclave(ctx, enclaveIdentifier); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
	}

	return api.PostEnclavesEnclaveIdentifierStop200JSONResponse{}, nil
}

// Get Engine Info
// (GET /engine/info)
func (engine EngineRuntime) GetEngineInfo(ctx context.Context, request api.GetEngineInfoRequestObject) (api.GetEngineInfoResponseObject, error) {
	result := api.GetEngineInfoResponse{EngineVersion: &engine.ImageVersionTag}
	return api.GetEngineInfo200JSONResponse(result), nil
}

// =============================================================================================================================================
// ============================================== Helper Functions =============================================================================
// =============================================================================================================================================

type StreamLogs struct {
	serviceLogsByServiceUuidChan chan map[user_service.ServiceUUID][]logline.LogLine
	errChan                      chan error
	requestedServicesUUID        []user_service.ServiceUUID
	missingServicesUUID          []string
	contextWithCancel            context.Context
	cancelCtxFunc                func()
}

func (response StreamLogs) VisitPostEnclavesEnclaveIdentifierLogsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	enc := json.NewEncoder(w)
	for {
		select {
		//stream case
		case serviceLogsByServiceUuid, isChanOpen := <-response.serviceLogsByServiceUuidChan:
			//If the channel is closed means that the logs database client won't continue sending streams
			if !isChanOpen {
				logrus.Debug("Exiting the stream loop after receiving a close signal from the service logs by service UUID channel")
				return nil
			}

			getServiceLogsResponse := newLogsResponseHttp(response.requestedServicesUUID, serviceLogsByServiceUuid, response.missingServicesUUID)
			enc.Encode(getServiceLogsResponse)
		//client cancel ctx case
		case <-response.contextWithCancel.Done():
			logrus.Debug("The user service logs stream has done")
			return nil
		//error from logs database case
		case _, isChanOpen := <-response.errChan:
			if isChanOpen {
				logrus.Debug("Exiting the stream because an error from the logs database client was received through the error chan.")
				return nil
			}
			logrus.Debug("Exiting the stream loop after receiving a close signal from the error chan")
			return nil
		}
	}
}

func (service *EngineRuntime) reportAnyMissingUuidsAndGetNotFoundUuidsListHttp(
	ctx context.Context,
	enclaveUuid enclave.EnclaveUUID,
	requestedServiceUuids map[user_service.ServiceUUID]bool,
) ([]string, error) {
	// doesn't matter which logs client is used here
	existingServiceUuids, err := service.PerWeekLogsDatabaseClient.FilterExistingServiceUuids(ctx, enclaveUuid, requestedServiceUuids)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred retrieving the exhaustive list of service UUIDs from the log client for enclave '%v' and for the requested UUIDs '%+v'", enclaveUuid, requestedServiceUuids)
	}

	notFoundServiceUuidsMap := getNotFoundServiceUuidsAndEmptyServiceLogsMap(requestedServiceUuids, existingServiceUuids)
	var notFoundServiceUuids []string
	for service, _ := range notFoundServiceUuidsMap {
		notFoundServiceUuids = append(notFoundServiceUuids, service)
	}
	return notFoundServiceUuids, nil
}

func newLogsResponseHttp(
	requestedServiceUuids []user_service.ServiceUUID,
	serviceLogsByServiceUuid map[user_service.ServiceUUID][]logline.LogLine,
	initialNotFoundServiceUuids []string,
) *api.GetServiceLogsResponse {
	serviceLogLinesByUuid := make(map[string]api.LogLine, len(serviceLogsByServiceUuid))
	notFoundServiceUuids := make([]string, len(initialNotFoundServiceUuids))
	for _, serviceUuid := range requestedServiceUuids {
		serviceUuidStr := string(serviceUuid)
		isInNotFoundUuidList := slices.Contains(initialNotFoundServiceUuids, serviceUuidStr)
		serviceLogLines, found := serviceLogsByServiceUuid[serviceUuid]
		// should continue in the not-found-UUID list
		if !found && isInNotFoundUuidList {
			notFoundServiceUuids = append(notFoundServiceUuids, serviceUuidStr)
		}

		// there is no new log lines but is a found UUID, so it has to be included in the service logs map
		if !found && !isInNotFoundUuidList {
			serviceLogLinesByUuid[serviceUuidStr] = api.LogLine{
				Line:      nil,
				Timestamp: nil,
			}
		}

		logLines := newHttpBindingsLogLineFromLogLines(serviceLogLines)
		serviceLogLinesByUuid[serviceUuidStr] = logLines
	}

	getServiceLogsResponse := &api.GetServiceLogsResponse{
		NotFoundServiceUuidSet:   &notFoundServiceUuids,
		ServiceLogsByServiceUuid: &serviceLogLinesByUuid,
	}
	return getServiceLogsResponse
}

func newHttpBindingsLogLineFromLogLines(logLines []logline.LogLine) api.LogLine {
	logLinesStr := make([]string, len(logLines))
	var logTimestamp time.Time

	for logLineIndex, logLine := range logLines {
		logLinesStr[logLineIndex] = logLine.GetContent()
		logTimestamp = logLine.GetTimestamp()
	}

	return api.LogLine{Line: &logLinesStr, Timestamp: &logTimestamp}

}

func newConjunctiveLogLineFiltersFromHttpLogLineFilters(
	logLineFilters []api.LogLineFilter,
) (logline.ConjunctiveLogLineFilters, error) {
	var conjunctiveLogLineFilters logline.ConjunctiveLogLineFilters

	for _, logLineFilter := range logLineFilters {
		var filter *logline.LogLineFilter
		operator := logLineFilter.Operator
		filterTextPattern := logLineFilter.TextPattern
		switch *operator {
		case api.DOESCONTAINTEXT:
			filter = logline.NewDoesContainTextLogLineFilter(*filterTextPattern)
		case api.DOESNOTCONTAINTEXT:
			filter = logline.NewDoesNotContainTextLogLineFilter(*filterTextPattern)
		case api.DOESCONTAINMATCHREGEX:
			filter = logline.NewDoesContainMatchRegexLogLineFilter(*filterTextPattern)
		case api.DOESNOTCONTAINMATCHREGEX:
			filter = logline.NewDoesNotContainMatchRegexLogLineFilter(*filterTextPattern)
		default:
			return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in GRPC filter '%v'; this is a bug in Kurtosis", operator, logLineFilter)
		}
		conjunctiveLogLineFilters = append(conjunctiveLogLineFilters, *filter)
	}

	return conjunctiveLogLineFilters, nil
}

// If the enclave was created prior to log retention, return the per file logs client
func (service *EngineRuntime) getLogsDatabaseClient(enclaveCreationTime time.Time) centralized_logs.LogsDatabaseClient {
	if enclaveCreationTime.After(logRetentionFeatureReleaseTime) {
		return service.PerWeekLogsDatabaseClient
	} else {
		return service.PerFileLogsDatabaseClient
	}
}

func (service *EngineRuntime) getEnclaveCreationTime(ctx context.Context, enclaveUuid enclave.EnclaveUUID) (time.Time, error) {
	enclaves, err := service.EnclaveManager.GetEnclaves(ctx)
	if err != nil {
		return time.Time{}, err
	}

	enclaveObj, found := enclaves[string(enclaveUuid)]
	if !found {
		return time.Time{}, stacktrace.NewError("Engine could not find enclave '%v'", enclaveUuid)
	}

	timestamp := enclaveObj.CreationTime
	return timestamp, nil
}
