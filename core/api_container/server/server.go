/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/lifecycle_service"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"net"
)

type ApiContainerServer struct {
	listenProtocol string
	listenAddress string
	// TODO some pluggable component based on the mode
}

func (server ApiContainerServer) Serve() error {
	grpcServer := grpc.NewServer()

	shutdownChan := make(chan interface{})
	lifecycleService := lifecycle_service.NewLifecycleService(shutdownChan)
	bindings.RegisterLifecycleServiceServer(grpcServer, lifecycleService)

	// TODO register more services here

	listener, err := net.Listen(server.listenProtocol, server.listenAddress)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred creating the listener on %v/%v",
			server.listenProtocol,
			server.listenAddress)
	}
	if err := grpcServer.Serve(listener); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred starting the gRPC server")
	}
}