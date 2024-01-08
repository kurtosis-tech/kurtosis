package connect_server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	ConnectHTTPServerLogPrefix = "[Connect-HTTP-ERROR]"
)

type ConnectServer struct {
	listenPort      uint16
	path            string
	handler         http.Handler
	stopGracePeriod time.Duration // How long we'll give the server to stop after asking nicely before we kill it
}

func NewConnectServer(listenPort uint16, stopGracePeriod time.Duration, handler http.Handler, path string) *ConnectServer {
	return &ConnectServer{
		listenPort:      listenPort,
		stopGracePeriod: stopGracePeriod,
		handler:         handler,
		path:            path,
	}
}
func (server *ConnectServer) RunServerUntilInterrupted() error {
	return server.RunServerUntilInterruptedWithCors(cors.Default())
}

func (server *ConnectServer) RunServerUntilInterruptedWithCors(cors *cors.Cors) error {
	// Signals are used to interrupt the server, so we catch them here
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverStopChan := make(chan struct{}, 1)
	go func() {
		<-termSignalChan
		interruptSignal := struct{}{}
		serverStopChan <- interruptSignal
	}()
	if err := server.RunServerUntilStopped(serverStopChan, cors); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the server using the interrupt channel for stopping")
	}
	return nil
}

func (server *ConnectServer) RunServerUntilStopped(
	stopper <-chan struct{},
	cors *cors.Cors,
) error {
	mux := http.NewServeMux()

	mux.Handle(server.path, server.handler)

	// nolint:exhaustruct
	httpServer := http.Server{
		Addr:     fmt.Sprintf(":%v", server.listenPort),
		Handler:  cors.Handler(h2c.NewHandler(mux, &http2.Server{})),
		ErrorLog: log.New(logrus.StandardLogger().Out, ConnectHTTPServerLogPrefix, log.Ldate|log.Ltime|log.Lshortfile),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Infof("Error occurred while starting the server, error: %+v", err)
		}
	}()

	<-stopper
	serverStoppedChan := make(chan interface{})
	go func() {
		if err := httpServer.Shutdown(context.Background()); err != nil {
			logrus.WithError(err).Error("Failed to shutdown the HTTP server")
		}
		serverStoppedChan <- nil
	}()
	select {
	case <-serverStoppedChan:
		logrus.Debug("API server has exited gracefully")
	case <-time.After(server.stopGracePeriod):
		if err := httpServer.Close(); err != nil {
			logrus.Infof("Error occurred while forcefully closing the server, error: %+v", err)
		}
	}

	return nil
}
