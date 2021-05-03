/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/rpc_api/rpc_api_consts"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	grpcServerStopGracePeriod = 10 * time.Second
)

type ApiContainerServer struct {
	core *ApiContainerService
}

func NewApiContainerServer(core *ApiContainerService) *ApiContainerServer {
	return &ApiContainerServer{core: core}
}

func (server ApiContainerServer) Run() error {
	grpcServer := grpc.NewServer()

	listenAddressStr := fmt.Sprintf(":%v", rpc_api_consts.ListenPort)
	listener, err := net.Listen(rpc_api_consts.ListenProtocol, listenAddressStr)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the listener on %v/%v",
			rpc_api_consts.ListenProtocol,
			listenAddressStr,
		)
	}

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	grpcServerResultChan := make(chan error)

	go func() {
		var resultErr error = nil
		if err := grpcServer.Serve(listener); err != nil {
			resultErr = stacktrace.Propagate(err, "The gRPC server exited with an error")
		}
		grpcServerResultChan <- resultErr
	}()

	// Wait until we get a shutdown signal
	<- termSignalChan

	serverStoppedChan := make(chan interface{})
	go func() {
		grpcServer.GracefulStop()
		serverStoppedChan <- nil
	}()
	select {
	case <- serverStoppedChan:
		logrus.Info("gRPC server has exited gracefully")
	case <- time.After(grpcServerStopGracePeriod):
		logrus.Warnf("gRPC server failed to stop gracefully after %v; hard-stopping now...", grpcServerStopGracePeriod)
		grpcServer.Stop()
		logrus.Info("gRPC server was forcefully stopped")
	}
	if err := <- grpcServerResultChan; err != nil {
		// Technically this doesn't need to be an error, but we make it so to fail loudly
		return stacktrace.Propagate(err, "gRPC server returned an error after it was done serving")
	}

	return nil
}

/*
func waitForExitCondition(
		suiteRegistrationChan chan interface{},
		termSignalChan chan os.Signal,
		shutdownChan chan int,
		mainService ApiContainerServerService) int {
	select {
	case <- suiteRegistrationChan:
		logrus.Debugf("Suite registered")
	// To guard against bugs in the testsuite container, we require a testsuite to register itself within
	//  a certain amount of time else the API container will kill itself with an error
	case <- time.After(suiteRegistrationTimeout):
		logrus.Errorf("No test suite registered itself after waiting for %v", suiteRegistrationTimeout)
		return api_container_exit_codes.NoTestSuiteRegistered
	// We don't technically have to catch this, but it'll help catch code bugs (it indicates that a service is sending
	//  a shutdown event before a testsuite is even registered)
	case exitCode := <- shutdownChan:
		logrus.Errorf(
			"Received shutdown event with exit code '%v' before testsuite is even registered; this is a code bug",
			exitCode)
		return api_container_exit_codes.ShutdownEventBeforeSuiteRegistration
	case termSignal := <-termSignalChan:
		logrus.Infof("Received term signal '%v' while waiting for suite registration", termSignal)
		return api_container_exit_codes.ReceivedTermSignal
	}

	if err := mainService.HandleSuiteRegistrationEvent(); err != nil {
		logrus.Errorf("Encountered an error sending the testsuite registration event to the main service:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		return api_container_exit_codes.StartupError
	}

	// NOTE: We intentionally don't set a timeout here, so the API container could run forever
	//  If this becomes problematic, we could add a very long timeout here
	select {
	case exitCode := <- shutdownChan:
		logrus.Infof("Received signal to shutdown with exit code '%v'", exitCode)
		return exitCode
	case termSignal := <-termSignalChan:
		logrus.Infof("Received term signal '%v' while waiting for exit condition", termSignal)
		return api_container_exit_codes.ReceivedTermSignal
	}
}

 */