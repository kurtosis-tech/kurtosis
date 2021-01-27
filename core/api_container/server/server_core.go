/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"google.golang.org/grpc"
)

type ApiContainerServerCore interface {
	// The action that will be returned to the testsuite container that registers itself with the server
	GetSuiteAction() bindings.SuiteAction

	// Creates a server using the given shutdownChan, and registers it with the given grpc server
	CreateAndRegisterService(shutdownChan chan int, grpcServer *grpc.Server) ApiContainerServerService
}
