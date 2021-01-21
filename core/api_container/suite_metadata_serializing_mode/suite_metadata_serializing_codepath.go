/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_metadata_serializing_mode

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

type SuiteMetadataSerializingCodepath struct {
	grpcServer    *grpc.Server
	listenAddress string
	args          SuiteMetadataSerializingArgs
}

func NewSuiteMetadataSerializingCodepath(args SuiteMetadataSerializingArgs) *SuiteMetadataSerializingCodepath {
	return &SuiteMetadataSerializingCodepath{args: args}
}

func (codepath SuiteMetadataSerializingCodepath) Execute() (int, error) {
	shutdownChan := make(chan interface{})
	lifecycleService := lifecycle_service.NewLifecycleService(shutdownChan)

	suiteMetadataSerializingService := New

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

