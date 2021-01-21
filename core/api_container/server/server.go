/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/suite_registration_service"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type ApiContainerServer struct {
	coreFactory    ApiContainerServerCoreFactory
	listenProtocol string
	listenAddress  string
	// TODO some pluggable component based on the mode
}

func (server ApiContainerServer) Serve() exit_codes.ApiContainerExitCode {
	grpcServer := grpc.NewServer()

	isSuiteRegistered := NewConcurrentBool(false);

	// TODO get the correct SuiteAction
	suiteAction := bindings.SuiteAction_SERIALIZE_SUITE_METADATA

	suiteRegistrationSvc := suite_registration_service.NewSuiteRegistrationService(suiteAction, isSuiteRegistered)
	bindings.RegisterSuiteRegistrationServiceServer(suiteRegistrationSvc)

	shutdownChan := make(chan exit_codes.ApiContainerExitCode)
	core := server.coreFactory.Create(shutdownChan)

	core.RegisterServices(grpcServer)

	listener, err := net.Listen(server.listenProtocol, server.listenAddress)
	if err != nil {
		logrus.Errorf("An error occurred creating the listener on %v/%v:",
			server.listenProtocol,
			server.listenAddress)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		return exit_codes.StartupErrorExitCode
	}

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		// TODO handle error response here???
		grpcServer.Serve(listener)
	}()

	// TODO call post-startup hook

	// TODO modify wait conditions here
	select {
	case signal := <- signalChan:
		logrus.Infof("Received signal '%v'; server will shut down", signal)
		return exit_codes.ReceivedTermSignalExitCode
	}

	// TODO call shutdown hook??????

	// We use Stop rather than GracefulStop here because a stop condition means everything should shut down immediately
	grpcServer.Stop()
	return nil
}