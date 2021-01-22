/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	// The amount of time a testsuite container has after registering itself with the API container to register
	//  a test execution (there should be no reason that registering test execution doesn't happen immediately)
	testExecutionRegistrationTimeout = 10 * time.Second

	awaitCompletionOrTimeoutThreadName = "Await completion/timeout thread"
	awaitCompletionThreadName = "Await completion thread"
)


type TestExecutionService struct {
	dockerManager *docker_manager.DockerManager

	testSuiteContainerId string
	stateMachine *testExecutionServiceStateMachine
	shutdownChan chan exit_codes.ApiContainerExitCode
}

func NewTestExecutionService(shutdownChan chan exit_codes.ApiContainerExitCode) *TestExecutionService {
	return &TestExecutionService{
		stateMachine: newTestExecutionServiceStateMachine(),
		shutdownChan: shutdownChan,
	}
}

func (service *TestExecutionService) HandleSuiteRegistrationEvent() error {
	if err := service.stateMachine.assertAndAdvance(waitingForSuiteRegistration); err != nil {
		return stacktrace.Propagate(
			err,
			"Cannot register test suite; an error occurred advancing the state machine")
	}

	// Launch timeout thread that will error if a test execution isn't registered soon
	go func() {
		time.Sleep(testExecutionRegistrationTimeout)
		if err := service.stateMachine.assert(waitingForTestExecutionRegistration); err == nil {
			service.shutdownChan <- exit_codes.NoTestExecutionRegisteredExitCode
		}
	}()

	return nil
}

func (service *TestExecutionService) RegisterTestExecution(ctx context.Context, args *bindings.RegisterTestExecutionArgs) (*emptypb.Empty, error) {
	if err := service.stateMachine.assertAndAdvance(waitingForTestExecutionRegistration); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register test execution; an error occurred advancing the state machine")
	}

	timeoutSeconds := args.TimeoutSeconds
	timeout := time.Duration(timeoutSeconds) * time.Second

	// Launch timeout thread that will error if the test execution doesn't complete within the allotted time limit
	go func() {
		time.Sleep(timeout)
		if err := service.stateMachine.assert(waitingForExecutionCompletion); err == nil {
			service.shutdownChan <- exit_codes.TestHitTimeoutExitCode
		}
	}()

	// Launch thread to monitor the state of the testsuite container
	go func() {
		// We use the background context so that waiting continues even when the request finishes
		if _, err := service.dockerManager.WaitForExit(context.Background(), service.testSuiteContainerId); err != nil {
			logrus.Errorf("An error occurred waiting for the testsuite container to exit:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			service.shutdownChan <- exit_codes.ErrWaitingForSuiteContainerExitExitCode
			return
		}
		if err := service.stateMachine.assertAndAdvance(waitingForExecutionCompletion); err != nil {
			logrus.Warnf("The testsuite container exited, but an error occurred advancing the state machine to its final state")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
		service.shutdownChan <- exit_codes.SuccessExitCode // TODO Rename this to "testsuite exited within timeout"
	}()

	return &emptypb.Empty{}, nil
}

func (service *TestExecutionService) RegisterService(ctx context.Context, args *bindings.RegisterServiceArgs) (*bindings.RegisterServiceResponse, error) {
	if err := service.stateMachine.assert(waitingForExecutionCompletion); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register service; test execution service wasn't in expected state '%v'", waitingForExecutionCompletion)
	}


}

func (service *TestExecutionService) StartService(ctx context.Context, args *bindings.StartServiceArgs) (*emptypb.Empty, error) {
	panic("implement me")
}

