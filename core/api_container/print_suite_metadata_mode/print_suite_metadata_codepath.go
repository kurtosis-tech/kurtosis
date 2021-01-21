/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package print_suite_metadata_mode

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/lifecycle_service"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"net"
)

const (
	listenProtocol = "tcp"
)

type PrintSuiteMetadataCodepath struct {
	grpcServer *grpc.Server
	listenAddress string
	args PrintSuiteMetadataArgs
}

func NewPrintSuiteMetadataCodepath(args PrintSuiteMetadataArgs) *PrintSuiteMetadataCodepath {
	return &PrintSuiteMetadataCodepath{args: args}
}

func (codepath PrintSuiteMetadataCodepath) Execute() (int, error) {
	shutdownChan := make(chan interface{})
	lifecycleService := lifecycle_service.NewLifecycleService(shutdownChan)
	bindings.RegisterLifecycleServiceServer(codepath.grpcServer, lifecycleService)
	// TODO register printsuitemetadata service

	listener, err := net.Listen(listenProtocol, codepath.listenAddress)
	if err != nil {
		return exit_codes.StartupErrorExitCode, stacktrace.Propagate(
			err,
			"An error occurred creating the listener on %v/%v",
			listenProtocol,
			codepath.listenAddress)
	}
	if err := codepath.grpcServer.Serve(listener); err != nil {
		return exit_codes.StartupErrorExitCode, stacktrace.Propagate(
			err,
			"An error occurred starting the gRPC server",
			listenProtocol,
			codepath.listenAddress)
	}

	codepath.grpcServer


}

