/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/test_execution/service_network"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"google.golang.org/grpc"
)

type TestExecutionServerCore struct {
	dockerManager *docker_manager.DockerManager
	serviceNetwork *service_network.ServiceNetwork
	testName string
	testSetupTimeout uint32
	testExecutionTimeout uint32
	testSuiteContainerId string
}

func NewTestExecutionServerCore(dockerManager *docker_manager.DockerManager, serviceNetwork *service_network.ServiceNetwork, testSetupTimeout uint32, testExecutionTimeout uint32, testName string, testSuiteContainerId string) *TestExecutionServerCore {
	return &TestExecutionServerCore{dockerManager: dockerManager, serviceNetwork: serviceNetwork, testName: testName, testSetupTimeout: testSetupTimeout, testExecutionTimeout: testExecutionTimeout, testSuiteContainerId: testSuiteContainerId}
}


func (core TestExecutionServerCore) GetSuiteAction() bindings.SuiteAction {
	return bindings.SuiteAction_EXECUTE_TEST
}

func (core TestExecutionServerCore) CreateAndRegisterService(
		shutdownChan chan int,
		grpcServer *grpc.Server) server.ApiContainerServerService {
	service := newTestExecutionService(
		core.dockerManager,
		core.serviceNetwork,
		core.testName,
		core.testExecutionTimeout,
		core.testSetupTimeout,
		core.testSuiteContainerId,
		shutdownChan)
	bindings.RegisterTestExecutionServiceServer(grpcServer, service)
	return service
}
