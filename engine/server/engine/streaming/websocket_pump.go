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
	websocket       *websocket.Conn
	inputChan       chan *T
	infoChan        chan *api_type.ResponseInfo
	ctx             context.Context
	cancelFunc      context.CancelFunc
	closed          bool
	connectionError *error
	onCloseCallback func()
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

	pump := &WebsocketPump[T]{
		websocket:       conn,
		inputChan:       make(chan *T),
		infoChan:        make(chan *api_type.ResponseInfo),
		ctx:             ctxWithCancel,
		cancelFunc:      cancelFunc,
		closed:          false,
		onCloseCallback: func() {},
	}

	go pump.startPumping()

	return pump, nil
}

func (pump *WebsocketPump[T]) readLoop() {
	for {
		_, _, err := pump.websocket.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (pump *WebsocketPump[T]) startPumping() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		pump.onCloseCallback()
		ticker.Stop()
		pump.websocket.Close()
		close(pump.inputChan)
		close(pump.infoChan)
		pump.closed = true
	}()

	logrus.WithFields(logrus.Fields{
		"pongWait":       pongWait,
		"pingPeriod":     pingPeriod,
		"maxMessageSize": maxMessageSize,
	}).Debug("Started keep alive process for websocket connection.")

	pump.websocket.SetReadLimit(maxMessageSize)
	if err := pump.websocket.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logrus.WithError(err).Error("Failed to set Pong wait time")
		pump.connectionError = &err
		return
	}
	// nolint:errcheck
	pump.websocket.SetPongHandler(func(string) error {
		logrus.Debug("Client is connected, got pong")
		return pump.websocket.SetReadDeadline(time.Now().Add(pongWait))
	})

	pump.websocket.SetCloseHandler(func(code int, text string) error {
		logrus.Infof("Websocket connection closed by the client - code: %d, msg: %s", code, text)
		pump.cancelFunc()
		return nil
	})

	// The read callbacks (handlers) are triggered from the ReadMessage calls, so
	// we also need a dummy reader loop.
	go pump.readLoop()

WRITE_LOOP:
	for {
		select {
		case <-ticker.C:
			if err := pump.websocket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logrus.Error("Websocket connection did not meet the write deadline")
				pump.connectionError = &err
				break WRITE_LOOP
			}
			if err := pump.websocket.WriteMessage(websocket.PingMessage, nil); err != nil {
				logrus.Error("Websocket connection is likely closed, exiting keep alive process")
				pump.connectionError = &err
				break WRITE_LOOP
			}
		case msg := <-pump.inputChan:
			if err := pump.websocket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logrus.Error("Websocket connection did not meet the write deadline")
				pump.connectionError = &err
				break WRITE_LOOP
			}
			if err := pump.websocket.WriteJSON(msg); err != nil {
				logrus.WithError(err).Warnf("Failed to send value of type `%T` via websocket", msg)
				pump.connectionError = &err
				break WRITE_LOOP
			}
		case msg := <-pump.infoChan:
			if err := pump.websocket.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logrus.Error("Websocket connection did not meet the write deadline")
				pump.connectionError = &err
				break WRITE_LOOP
			}
			if err := pump.websocket.WriteJSON(msg); err != nil {
				logrus.WithError(err).Warnf("Failed to send value of type `%T` via websocket", msg)
				pump.connectionError = &err
				break WRITE_LOOP
			}
		case <-pump.ctx.Done():
			logrus.Debug("Websocket pump has been asked to close, closing it.")
			break WRITE_LOOP
		}
	}
}

func (pump *WebsocketPump[T]) PumpResponseInfo(msg *api_type.ResponseInfo) error {
	if pump.closed {
		if pump.connectionError != nil {
			return stacktrace.Propagate(*pump.connectionError, "Websocket has been closed due connection error")
		}
		return nil
	}

	select {
	case _, ok := <-pump.infoChan:
		if !ok {
			logrus.Debug("Worker channel closed, cannot send message")
		}
		if pump.connectionError != nil {
			return stacktrace.Propagate(*pump.connectionError, "Websocket has been closed due connection error")
		}
		return stacktrace.NewError("Websocket has been closed due connection error")
	case pump.infoChan <- msg:
		return nil
	}
}

func (pump *WebsocketPump[T]) PumpMessage(msg *T) error {
	if pump.closed {
		if pump.connectionError != nil {
			return stacktrace.Propagate(*pump.connectionError, "Websocket has been closed due connection error")
		}
		return nil
	}

	select {
	case _, ok := <-pump.inputChan:
		if !ok {
			logrus.Debug("Worker channel closed, cannot send message")
		}
		if pump.connectionError != nil {
			return stacktrace.Propagate(*pump.connectionError, "Websocket has been closed due connection error")
		}
		return stacktrace.NewError("Websocket has been closed due connection error")
	case pump.inputChan <- msg:
		return nil
	}

}

func (pump *WebsocketPump[T]) Close() {
	pump.cancelFunc()
}

func (pump *WebsocketPump[T]) OnClose(callback func()) {
	pump.onCloseCallback = callback
}

func (pump *WebsocketPump[T]) IsClosed() (bool, *error) {
	return pump.closed, pump.connectionError
}
