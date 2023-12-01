package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_http"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/streaming"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

type WebSocketRuntime struct {
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

	MetricsClient     metrics_client.MetricsClient
	AsyncStarlarkLogs streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine]
}

func (engine WebSocketRuntime) GetEnclavesEnclaveIdentifierLogs(ctx echo.Context, enclaveIdentifier api_type.EnclaveIdentifier, params api_type.GetEnclavesEnclaveIdentifierLogsParams) error {
	streamer, err := streaming.NewServiceLogStreamer(
		ctx.Request().Context(),
		engine.EnclaveManager,
		enclaveIdentifier,
		engine.PerWeekLogsDatabaseClient,
		engine.PerFileLogsDatabaseClient,
		utils.MapList(params.ServiceUuidSet, func(x string) user_service.ServiceUUID { return user_service.ServiceUUID(x) }),
		params.FollowLogs,
		params.ReturnAllLogs,
		utils.MapPointer(params.NumLogLines, func(x int) uint32 { return uint32(x) }),
		params.ConjunctiveFilters,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"enclave_identifier": enclaveIdentifier,
			"parameters":         params,
		}).Error("Failed to create log stream")
		errInfo := api_type.ResponseInfo{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create log stream",
			Type:    api_type.ERROR,
		}
		writeResponseInfo(ctx, errInfo)
		return nil
	}

	if ctx.IsWebSocket() {
		logrus.Infof("Starting log stream using Websocket")
		streamServiceLogsWithWebsocket(ctx, *streamer)
	} else {
		logrus.Infof("Starting log stream using plain HTTP")
		streamServiceLogsWithHTTP(ctx, *streamer)
	}
	return nil

}

func (engine WebSocketRuntime) GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogs(ctx echo.Context, enclaveIdentifier api_type.EnclaveIdentifier, serviceIdentifier api_type.ServiceIdentifier, params api_type.GetEnclavesEnclaveIdentifierServicesServiceIdentifierLogsParams) error {
	serviceUuidStrSet := []user_service.ServiceUUID{user_service.ServiceUUID(serviceIdentifier)}
	streamer, err := streaming.NewServiceLogStreamer(
		ctx.Request().Context(),
		engine.EnclaveManager,
		enclaveIdentifier,
		engine.PerWeekLogsDatabaseClient,
		engine.PerFileLogsDatabaseClient,
		serviceUuidStrSet,
		params.FollowLogs,
		params.ReturnAllLogs,
		utils.MapPointer(params.NumLogLines, func(x int) uint32 { return uint32(x) }),
		params.ConjunctiveFilters,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"enclave_identifier": enclaveIdentifier,
			"parameters":         params,
		}).Error("Failed to create log stream")
		errInfo := api_type.ResponseInfo{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create log stream",
			Type:    api_type.ERROR,
		}
		writeResponseInfo(ctx, errInfo)
		return nil
	}

	if ctx.IsWebSocket() {
		logrus.Infof("Starting log stream using Websocket")
		streamServiceLogsWithWebsocket(ctx, *streamer)
	} else {
		logrus.Infof("Starting log stream using plain HTTP")
		streamServiceLogsWithHTTP(ctx, *streamer)
	}

	return nil
}

// (GET /enclaves/{enclave_identifier}/starlark/executions/{starlark_execution_uuid}/logs)
func (engine WebSocketRuntime) GetEnclavesEnclaveIdentifierStarlarkExecutionsStarlarkExecutionUuidLogs(ctx echo.Context, enclaveIdentifier api_type.EnclaveIdentifier, starlarkExecutionUuid api_type.StarlarkExecutionUuid) error {
	async_log_uuid := streaming.StreamerUUID(starlarkExecutionUuid)

	if ctx.IsWebSocket() {
		logrus.Infof("Starting log stream using Websocket for streamer UUUID: %s", starlarkExecutionUuid)
		streamStarlarkLogsWithWebsocket(ctx, engine.AsyncStarlarkLogs, async_log_uuid)
	} else {
		logrus.Infof("Starting log stream using plain HTTP for streamer UUUID: %s", starlarkExecutionUuid)
		streamStarlarkLogsWithHTTP(ctx, engine.AsyncStarlarkLogs, async_log_uuid)
	}

	return nil
}

// =============================================================================================================================================
// ============================================== Helper Functions =============================================================================
// =============================================================================================================================================

func writeResponseInfo(ctx echo.Context, response api_type.ResponseInfo) {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	enc := json.NewEncoder(ctx.Response())
	ctx.Response().WriteHeader(int(response.Code))
	enc.Encode(response)
}

func streamStarlarkLogsWithWebsocket[T any](ctx echo.Context, streamerPool streaming.StreamerPool[T], streamerUUID streaming.StreamerUUID) {
	notFoundErr := api_type.ResponseInfo{
		Type:    api_type.INFO,
		Message: fmt.Sprintf("Log streaming '%s' not found. Either it has been consumed or has expired.", streamerUUID),
		Code:    http.StatusNotFound,
	}
	inPool := streamerPool.Contains(streamerUUID)
	if !inPool {
		writeResponseInfo(ctx, notFoundErr)
		return
	}

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		found, err := streamerPool.Consume(streaming.StreamerUUID(streamerUUID), func(logline *rpc_api.StarlarkRunResponseLine) error {
			err := websocket.JSON.Send(ws, utils.MapPointer(logline, to_http.ToHttpApiStarlarkRunResponseLine))
			if err != nil {
				return err
			}
			return nil
		})

		if !found {
			websocket.JSON.Send(ws, notFoundErr)
		}

		if err != nil {
			logrus.Errorf("Failed to stream all data %s", err)
			streamingErr := api_type.ResponseInfo{
				Type:    api_type.ERROR,
				Message: fmt.Sprintf("Log streaming '%s' failed while sending the data", streamerUUID),
				Code:    http.StatusInternalServerError,
			}
			websocket.JSON.Send(ws, streamingErr)
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())
}

func streamStarlarkLogsWithHTTP[T any](ctx echo.Context, streamerPool streaming.StreamerPool[T], streamerUUID streaming.StreamerUUID) {
	notFoundErr := api_type.ResponseInfo{
		Type:    api_type.INFO,
		Message: fmt.Sprintf("Log streaming '%s' not found. Either it has been consumed or has expired.", streamerUUID),
		Code:    http.StatusNotFound,
	}
	inPool := streamerPool.Contains(streamerUUID)
	if !inPool {
		writeResponseInfo(ctx, notFoundErr)
		return
	}

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	enc := json.NewEncoder(ctx.Response())
	ctx.Response().WriteHeader(http.StatusOK)
	found, err := streamerPool.Consume(streaming.StreamerUUID(streamerUUID), func(logline *rpc_api.StarlarkRunResponseLine) error {
		if err := enc.Encode(utils.MapPointer(logline, to_http.ToHttpApiStarlarkRunResponseLine)); err != nil {
			return err
		}
		ctx.Response().Flush()
		return nil
	})

	if !found {
		enc.Encode(notFoundErr)
	}

	if err != nil {
		logrus.Errorf("Failed to stream all data %s", err)
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("Log streaming '%s' failed while sending the data", streamerUUID),
			Code:    http.StatusInternalServerError,
		}
		enc.Encode(streamingErr)
	}
}

func streamServiceLogsWithWebsocket(ctx echo.Context, streamer streaming.ServiceLogStreamer) {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		err := streamer.Consume(func(logline *api_type.ServiceLogs) error {
			err := websocket.JSON.Send(ws, logline)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			logrus.Errorf("Failed to stream all data %s", err)
			streamingErr := api_type.ResponseInfo{
				Type:    api_type.ERROR,
				Message: fmt.Sprintf("Log streaming failed while sending the data"),
				Code:    http.StatusInternalServerError,
			}
			websocket.JSON.Send(ws, streamingErr)
		}
	}).ServeHTTP(ctx.Response(), ctx.Request())
}

func streamServiceLogsWithHTTP(ctx echo.Context, streamer streaming.ServiceLogStreamer) {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	enc := json.NewEncoder(ctx.Response())
	ctx.Response().WriteHeader(http.StatusOK)
	err := streamer.Consume(func(logline *api_type.ServiceLogs) error {
		if err := enc.Encode(logline); err != nil {
			return err
		}
		ctx.Response().Flush()
		return nil
	})

	if err != nil {
		logrus.Errorf("Failed to stream all data %s", err)
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("Log streaming failed while sending the data"),
			Code:    http.StatusInternalServerError,
		}
		enc.Encode(streamingErr)
	}
}
