/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"google.golang.org/grpc"
)

type TestExecutionServerCore struct {
	
}

// TODO constructor

func (t TestExecutionServerCore) GetSuiteAction() bindings.SuiteAction {
	return bindings.SuiteAction_EXECUTE_TEST
}

func (t TestExecutionServerCore) CreateAndRegisterService(shutdownChan chan exit_codes.ApiContainerExitCode, grpcServer *grpc.Server) server.ApiContainerServerService {
	panic("implement me")
}
