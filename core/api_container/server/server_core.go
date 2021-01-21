/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"google.golang.org/grpc"
)

type ApiContainerServerCore interface {
	RegisterServices(grpcServer *grpc.Server)
}

type ApiContainerServerCoreFactory interface {
	Create(shutdownChan chan exit_codes.ApiContainerExitCode) ApiContainerServerCore
}
