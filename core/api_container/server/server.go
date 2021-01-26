/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/server/api_container_server_consts"
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
	// If no suite registers within this time, the API container will exit with an error
	suiteRegistrationTimeout = 10 * time.Second
)

type ApiContainerServer struct {
	core ApiContainerServerCore
}

func NewApiContainerServer(core ApiContainerServerCore) *ApiContainerServer {
	return &ApiContainerServer{core: core}
}

func (server ApiContainerServer) Run() int {
	grpcServer := grpc.NewServer()

	shutdownChan := make(chan int, 1)
	mainService := server.core.CreateAndRegisterService(shutdownChan, grpcServer)

	suiteRegistrationChan := make(chan interface{}, 1)
	suiteAction := server.core.GetSuiteAction()
	suiteRegistrationSvc := newSuiteRegistrationService(suiteAction, mainService, suiteRegistrationChan)
	bindings.RegisterSuiteRegistrationServiceServer(grpcServer, suiteRegistrationSvc)

	listenAddressStr := fmt.Sprintf(":%v", api_container_server_consts.ListenPort)
	listener, err := net.Listen(api_container_server_consts.ListenProtocol, listenAddressStr)
	if err != nil {
		logrus.Errorf("An error occurred creating the listener on %v/%v",
			api_container_server_consts.ListenProtocol,
			listenAddressStr)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		return api_container_exit_codes.StartupError
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

	exitCode := waitForExitCondition(suiteRegistrationChan, termSignalChan, shutdownChan)

	// NOTE: If we see weirdness with graceful stop, we could use the hard Stop though then we'd need to consider that
	//  RPC calls which send the shutdown signal might get killed before they can return a response to the client
	grpcServer.GracefulStop()

	if err := <- grpcServerResultChan; err != nil {
		logrus.Errorf("gRPC server returned an error after it was done serving:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		exitCode = api_container_exit_codes.ShutdownError
	}

	if err := mainService.HandlePostShutdownEvent(); err != nil {
		logrus.Errorf("Post-shutdown hook on service returned an error:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		exitCode = api_container_exit_codes.ShutdownError
	}

	return exitCode
}

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