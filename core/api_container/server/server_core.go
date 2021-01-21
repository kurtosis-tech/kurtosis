/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"google.golang.org/grpc"
)

type ApiContainerServerCore interface {
	// The action that will be returned to the testsuite container that registers itself with the server
	GetSuiteAction() bindings.SuiteAction

	// Creates a server using the given shutdownChan, and registers it with the given grpc server
	CreateAndRegisterService(shutdownChan chan exit_codes.ApiContainerExitCode, grpcServer *grpc.Server) ApiContainerServerService
}
