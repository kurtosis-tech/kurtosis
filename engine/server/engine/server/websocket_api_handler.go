package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	user_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/client_implementations/persistent_volume/log_file_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/mapping/to_http"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/streaming"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/utils"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	rpc_api "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/kurtosis-tech/kurtosis/api/golang/http_rest/api_types"
)

const (
	wsReadBufferSize  = 1024
	wsWriteBufferSize = 1024
	maxMessageSize    = 512
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	// nolint:gomnd
	pingPeriod = (pongWait * 9) / 10
)

// nolint: exhaustruct
var upgrader = websocket.Upgrader{
	ReadBufferSize:  wsReadBufferSize,
	WriteBufferSize: wsWriteBufferSize,
}

type WebSocketRuntime struct {
	// The version tag of the engine server image, so it can report its own version
	ImageVersionTag string

	EnclaveManager *enclave_manager.EnclaveManager

	// The protected user ID for metrics analytics purpose
	MetricsUserID string

	// User consent to send metrics
	DidUserAcceptSendingMetrics bool

	// The clients for consuming container logs from the logs' database server
	LogsDatabaseClient centralized_logs.LogsDatabaseClient

	LogFileManager *log_file_manager.LogFileManager

	MetricsClient metrics_client.MetricsClient

	// Pool of Starlark log streamers create by package/script runs
	AsyncStarlarkLogs streaming.StreamerPool[*rpc_api.StarlarkRunResponseLine]
}

func (engine WebSocketRuntime) GetEnclavesEnclaveIdentifierLogs(ctx echo.Context, enclaveIdentifier api_type.EnclaveIdentifier, params api_type.GetEnclavesEnclaveIdentifierLogsParams) error {
	streamer, err := streaming.NewServiceLogStreamer(
		ctx.Request().Context(),
		engine.EnclaveManager,
		enclaveIdentifier,
		engine.LogsDatabaseClient,
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
			"stacktrace":         fmt.Sprintf("%+v", err),
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
		engine.LogsDatabaseClient,
		serviceUuidStrSet,
		params.FollowLogs,
		params.ReturnAllLogs,
		utils.MapPointer(params.NumLogLines, func(x int) uint32 { return uint32(x) }),
		params.ConjunctiveFilters,
	)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"enclave_identifier": enclaveIdentifier,
			"parameters":         params,
			"stacktrace":         fmt.Sprintf("%+v", err),
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

// (GET /starlark/executions/{starlark_execution_uuid}/logs)
func (engine WebSocketRuntime) GetStarlarkExecutionsStarlarkExecutionUuidLogs(ctx echo.Context, starlarkExecutionUuid api_type.StarlarkExecutionUuid) error {
	asyncLogUuid := streaming.StreamerUUID(starlarkExecutionUuid)

	if ctx.IsWebSocket() {
		logrus.Infof("Starting log stream using Websocket for streamer UUUID: %s", starlarkExecutionUuid)
		streamStarlarkLogsWithWebsocket(ctx, engine.AsyncStarlarkLogs, asyncLogUuid)
	} else {
		logrus.Infof("Starting log stream using plain HTTP for streamer UUUID: %s", starlarkExecutionUuid)
		streamStarlarkLogsWithHTTP(ctx, engine.AsyncStarlarkLogs, asyncLogUuid)
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
	_ = enc.Encode(response)
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

	wsPump, err := NewWebsocketPump[api_type.StarlarkRunResponseLine](ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to start websocket connection")
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		enc := json.NewEncoder(ctx.Response())
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		err = enc.Encode(api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: "Failed to start websocket connection",
			Code:    http.StatusInternalServerError,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to encode error message")
		}
		return
	}
	defer wsPump.Close()
	go wsPump.startPumping()

	found, err := streamerPool.Consume(streaming.StreamerUUID(streamerUUID), func(logline *rpc_api.StarlarkRunResponseLine) error {
		response, err := to_http.ToHttpStarlarkRunResponseLine(logline)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to convert value of type `%T` to http", logline)
		}
		wsPump.pumpMessage(response)
		return nil
	})

	if !found {
		wsPump.pumpResponseInfo(&notFoundErr)
	}

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"streamerUUID": streamerUUID,
			"stacktrace":   fmt.Sprintf("%+v", err),
		}).Error("Failed to stream all data")
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("Log streaming '%s' failed while sending the data", streamerUUID),
			Code:    http.StatusInternalServerError,
		}
		wsPump.pumpResponseInfo(&streamingErr)
	}
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
		response, err := to_http.ToHttpStarlarkRunResponseLine(logline)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to convert value of type `%T` to http", logline)
		}
		if err := enc.Encode(response); err != nil {
			return stacktrace.Propagate(err, "Failed to send value of type `%T` via http streaming", logline)
		}
		ctx.Response().Flush()
		return nil
	})

	if !found {
		if err := enc.Encode(notFoundErr); err != nil {
			logrus.WithError(err).Errorf("Failed to send value via websocket")
		}
	}

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"streamerUUID": streamerUUID,
			"stacktrace":   fmt.Sprintf("%+v", err),
		}).Error("Failed to stream all data")
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: fmt.Sprintf("Log streaming '%s' failed while sending the data", streamerUUID),
			Code:    http.StatusInternalServerError,
		}
		if err := enc.Encode(streamingErr); err != nil {
			logrus.WithError(err).Errorf("Failed to send value via websocket")
		}
	}
}

func streamServiceLogsWithWebsocket(ctx echo.Context, streamer streaming.ServiceLogStreamer) {
	wsPump, err := NewWebsocketPump[api_type.ServiceLogs](ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to start websocket connection")
		ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		enc := json.NewEncoder(ctx.Response())
		ctx.Response().WriteHeader(http.StatusInternalServerError)
		err = enc.Encode(api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: "Failed to start websocket connection",
			Code:    http.StatusInternalServerError,
		})
		if err != nil {
			logrus.WithError(err).Error("Failed to encode error message")
		}
		return
	}
	defer wsPump.Close()
	go wsPump.startPumping()

	err = streamer.Consume(func(logline *api_type.ServiceLogs) error {
		wsPump.pumpMessage(logline)
		return nil
	})

	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"stacktrace": fmt.Sprintf("%+v", err),
			"services":   streamer.GetRequestedServiceUuids(),
		}).Error("Failed to stream all data")
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: "Log streaming failed while sending the data",
			Code:    http.StatusInternalServerError,
		}
		wsPump.pumpResponseInfo(&streamingErr)
	}
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
		logrus.WithError(err).WithFields(logrus.Fields{
			"stacktrace": fmt.Sprintf("%+v", err),
			"services":   streamer.GetRequestedServiceUuids(),
		}).Error("Failed to stream all data")
		streamingErr := api_type.ResponseInfo{
			Type:    api_type.ERROR,
			Message: "Log streaming failed while sending the data",
			Code:    http.StatusInternalServerError,
		}
		if err := enc.Encode(streamingErr); err != nil {
			logrus.WithError(err).Errorf("Failed to send value via websocket")
		}
	}
}

type WebsocketPump[T interface{}] struct {
	websocket  *websocket.Conn
	inputChan  chan *T
	infoChan   chan *api_type.ResponseInfo
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewWebsocketPump[T interface{}](ctx echo.Context) (*WebsocketPump[T], error) {
	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to upgrade http connection to websocket")
	}

	ctxWithCancel, cancelFunc := context.WithCancel(context.Background())

	return &WebsocketPump[T]{
		websocket:  conn,
		inputChan:  make(chan *T),
		ctx:        ctxWithCancel,
		cancelFunc: cancelFunc,
	}, nil
}

func (pump WebsocketPump[T]) startPumping() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		pump.websocket.Close()
	}()

	logrus.WithFields(logrus.Fields{
		"pongWait":       pongWait,
		"pingPeriod":     pingPeriod,
		"maxMessageSize": maxMessageSize,
	}).Debug("Started keep alive process for websocket connection.")

	pump.websocket.SetReadLimit(maxMessageSize)
	if err := pump.websocket.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logrus.WithError(err).Error("Failed to set Pong wait time")
		return
	}
	// nolint:errcheck
	pump.websocket.SetPongHandler(func(string) error { return pump.websocket.SetReadDeadline(time.Now().Add(pongWait)) })

	for {
		select {
		case <-ticker.C:
			if err := pump.websocket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logrus.Debug("Websocket connection is likely closed, exiting keep alive process")
				return
			}
			if err := pump.websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
				logrus.Debug("Websocket connection is likely closed, exiting keep alive process")
				return
			}
		case msg := <-pump.inputChan:
			if err := pump.websocket.WriteJSON(msg); err != nil {
				logrus.WithError(stacktrace.Propagate(err, "Failed to send value of type `%T` via websocket", msg)).Errorf("Failed to write message to websocket, closing it.")
				return
			}
		case msg := <-pump.infoChan:
			if err := pump.websocket.WriteJSON(msg); err != nil {
				logrus.WithError(stacktrace.Propagate(err, "Failed to send value of type `%T` via websocket", msg)).Errorf("Failed to write message to websocket, closing it.")
				return
			}
		case <-pump.ctx.Done():
			logrus.Debug("Websocket pumper has been asked to close, closing it.")
			return
		}
	}
}

func (pump *WebsocketPump[T]) pumpResponseInfo(msg *api_type.ResponseInfo) {
	pump.infoChan <- msg
}

func (pump *WebsocketPump[T]) pumpMessage(msg *T) {
	pump.inputChan <- msg
}

func (pump *WebsocketPump[T]) Close() {
	pump.cancelFunc()
}
