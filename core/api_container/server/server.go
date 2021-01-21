/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
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

type serverConfigForMode struct {
	// The action the testsuite should take when the API container is in the given mode
	suiteAction bindings.SuiteAction

	serviceFactory func(shutdownChan chan exit_codes.ApiContainerExitCode) ApiContainerServerService
}

type ApiContainerServer struct {
	core ApiContainerServerCore

	listenProtocol string
	listenAddress  string
}

func (server ApiContainerServer) Serve() exit_codes.ApiContainerExitCode {
	grpcServer := grpc.NewServer()

	suiteRegistrationChan := make(chan interface{}, 1)
	suiteAction := server.core.GetSuiteAction()
	suiteRegistrationSvc := newSuiteRegistrationService(suiteAction, suiteRegistrationChan)
	bindings.RegisterSuiteRegistrationServiceServer(grpcServer, suiteRegistrationSvc)

	shutdownChan := make(chan exit_codes.ApiContainerExitCode, 1)
	service := server.core.CreateAndRegisterService(shutdownChan, grpcServer)

	listener, err := net.Listen(server.listenProtocol, server.listenAddress)
	if err != nil {
		logrus.Errorf("An error occurred creating the listener on %v/%v:",
			server.listenProtocol,
			server.listenAddress)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		return exit_codes.StartupErrorExitCode
	}

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	termSignalChan := make(chan os.Signal, 1)
	signal.Notify(termSignalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logrus.Errorf("gRPC server returned an error after it was done serving:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
	}()

	exitCode := waitForExitCondition(suiteRegistrationChan, termSignalChan, shutdownChan, service)

	// TODO call shutdown hook??????

	// We use Stop rather than GracefulStop here because a stop condition means everything should shut down immediately
	grpcServer.Stop()
	return exitCode
}

func waitForExitCondition(
		suiteRegistrationChan chan interface{},
		termSignalChan chan os.Signal,
		shutdownChan chan exit_codes.ApiContainerExitCode,
		service ApiContainerServerService) exit_codes.ApiContainerExitCode {
	select {
	case <- suiteRegistrationChan:
		logrus.Debugf("Suite registered")
	// To guard against bugs in the testsuite container, we require a testsuite to register itself within
	//  a certain amount of time else the API container will kill itself with an error
	case <- time.After(suiteRegistrationTimeout):
		logrus.Errorf("No test suite registered itself after waiting for %v", suiteRegistrationTimeout)
		return exit_codes.NoTestSuiteRegisteredExitCode
	// We don't technically have to catch this, but it'll help catch code bugs (it indicates that a service is sending
	//  a shutdown event before a testsuite is even registered)
	case <- shutdownChan:
		logrus.Errorf("Received shutdown event with exit code '%v' before testsuite is even registered; this is a code bug")
		return exit_codes.ShutdownEventBeforeSuiteRegistration
	case termSignal := <-termSignalChan:
		logrus.Infof("Received term signal '%v' while waiting for suite registration", termSignal)
		return exit_codes.ReceivedTermSignalExitCode
	}

	service.ReceiveSuiteRegistrationEvent()

	// NOTE: We intentionally don't set a timeout here, so the API container could run forever
	//  If this becomes problematic, we could add a very long timeout here
	select {
	case exitCode := <- shutdownChan:
		logrus.Infof("Received signal to shutdown with exit code '%v'", exitCode)
		return exitCode
	case termSignal := <-termSignalChan:
		logrus.Infof("Received term signal '%v' while waiting for exit condition", termSignal)
		return exit_codes.ReceivedTermSignalExitCode
	}
}