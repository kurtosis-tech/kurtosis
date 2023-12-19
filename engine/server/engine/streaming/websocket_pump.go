package streaming

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/labstack/echo/v4"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

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

type WebsocketPump[T interface{}] struct {
	websocket  *websocket.Conn
	inputChan  chan *T
	infoChan   chan *api_type.ResponseInfo
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewWebsocketPump[T interface{}](ctx echo.Context, cors cors.Cors) (*WebsocketPump[T], error) {
	// nolint: exhaustruct
	upgrader := websocket.Upgrader{
		ReadBufferSize:  wsReadBufferSize,
		WriteBufferSize: wsWriteBufferSize,
		CheckOrigin:     cors.OriginAllowed,
	}

	conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to upgrade http connection to websocket")
	}

	ctxWithCancel, cancelFunc := context.WithCancel(context.Background())

	return &WebsocketPump[T]{
		websocket:  conn,
		inputChan:  make(chan *T),
		infoChan:   make(chan *api_type.ResponseInfo),
		ctx:        ctxWithCancel,
		cancelFunc: cancelFunc,
	}, nil
}

func (pump WebsocketPump[T]) StartPumping() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		pump.websocket.Close()
		close(pump.inputChan)
		close(pump.infoChan)
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

func (pump *WebsocketPump[T]) PumpResponseInfo(msg *api_type.ResponseInfo) {
	pump.infoChan <- msg
}

func (pump *WebsocketPump[T]) PumpMessage(msg *T) {
	pump.inputChan <- msg
}

func (pump *WebsocketPump[T]) Close() {
	pump.cancelFunc()
}
