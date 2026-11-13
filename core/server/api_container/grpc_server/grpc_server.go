package grpc_server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

const listenProtocol = "tcp"

type GrpcServer struct {
	listenPort               uint16
	stopGracePeriod          time.Duration
	serverOptions            []grpc.ServerOption
	serviceRegistrationFuncs []func(*grpc.Server)
}

func NewGrpcServer(listenPort uint16, stopGracePeriod time.Duration, serverOptions []grpc.ServerOption, serviceRegistrationFuncs []func(*grpc.Server)) *GrpcServer {
	return &GrpcServer{
		listenPort:               listenPort,
		stopGracePeriod:          stopGracePeriod,
		serverOptions:            serverOptions,
		serviceRegistrationFuncs: serviceRegistrationFuncs,
	}
}

func (server GrpcServer) RunUntilInterrupted() error {
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverStopChan := make(chan struct{}, 1)
	go func() {
		<-termSignalChan
		serverStopChan <- struct{}{}
	}()
	if err := server.RunUntilStopped(serverStopChan); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the server using the interrupt channel for stopping")
	}
	return nil
}

func (server GrpcServer) RunUntilStopped(stopper <-chan struct{}) error {
	loggingInterceptorFunc := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		grpcMethod := info.FullMethod
		logrus.Debugf("Received gRPC request to method '%v' with args:\n%+v", grpcMethod, req)
		resp, err := handler(ctx, req)
		if err != nil {
			logrus.Debugf("gRPC request to method '%v' failed with error:\n%v", grpcMethod, err)
		} else {
			logrus.Debugf("gRPC request to method '%v' succeeded with response:\n%+v", grpcMethod, resp)
		}
		return resp, err
	}
	serverOptions := append([]grpc.ServerOption{}, server.serverOptions...)
	serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(loggingInterceptorFunc))
	grpcServer := grpc.NewServer(serverOptions...)
	for _, registrationFunc := range server.serviceRegistrationFuncs {
		registrationFunc(grpcServer)
	}

	listener, err := net.Listen(listenProtocol, fmt.Sprintf(":%v", server.listenPort))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the listener on %v/%v", listenProtocol, server.listenPort)
	}

	maybeErrorResultChan := make(chan error)
	mux := cmux.New(listener)
	grpcWebListener := mux.Match(cmux.HTTP1Fast())
	grpcListener := mux.Match(cmux.Any())

	grpcWebServer := grpcweb.WrapServer(grpcServer, grpcweb.WithOriginFunc(func(origin string) bool { return true }))

	go func() {
		if resultErr := grpcServer.Serve(grpcListener); resultErr != nil {
			logrus.Debugf("error ocurred while creating grpc server: %v", resultErr)
			maybeErrorResultChan <- resultErr
		}
	}()

	go func() {
		// nolint: exhaustruct
		httpServer := &http.Server{
			Handler: http.Handler(grpcWebServer),
		}
		if resultErr := httpServer.Serve(grpcWebListener); resultErr != nil {
			logrus.Debugf("error ocurred while creating grpcweb server: %v", resultErr)
			maybeErrorResultChan <- resultErr
		}
	}()

	go func() {
		if resultErr := mux.Serve(); resultErr != nil {
			logrus.Debugf("error ocurred while creating mux server: %v", resultErr)
			maybeErrorResultChan <- resultErr
		}
	}()

	<-stopper

	serverStoppedChan := make(chan interface{})
	go func() {
		grpcServer.Stop()
		mux.Close()
		serverStoppedChan <- nil
	}()
	select {
	case <-serverStoppedChan:
		logrus.Debug("gRPC server has exited gracefully")
	case <-time.After(server.stopGracePeriod):
		logrus.Warnf("gRPC server failed to stop gracefully after %v; hard-stopping now...", server.stopGracePeriod)
		grpcServer.Stop()
		mux.Close()
		logrus.Debug("gRPC server was forcefully stopped")
	}

	if err := <-maybeErrorResultChan; err != nil && !isGracefulExit(err) {
		return stacktrace.Propagate(err, "gRPC server returned an error after it was done serving")
	}

	return nil
}

func isGracefulExit(err error) bool {
	switch err {
	case nil, http.ErrServerClosed, cmux.ErrListenerClosed, cmux.ErrServerClosed:
		return true
	default:
		return strings.Contains(err.Error(), "use of closed network connection")
	}
}
