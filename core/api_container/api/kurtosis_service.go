/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/test_execution_status"
	"github.com/kurtosis-tech/kurtosis/api_container/partitioning"
	"github.com/kurtosis-tech/kurtosis/api_container/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const (

	// How long we'll wait when making a best-effort attempt to stop a container
	containerStopTimeout = 15 * time.Second

	// Prefixes for the two types of threads we'll spin up
	completionTimeoutPrefix = "Test completion/timeout thread"
	completionPrefix        = "Test completion thread"
)

type KurtosisService struct {
	// The Docker container ID of the test suite that will be making calls against this Kurtosis API
	// This is expected to be nil until the "register test suite container" endpoint is called, which should be called
	//  exactly once
	testSuiteContainerId string

	dockerManager *commons.DockerManager

	partitioningEngine *partitioning.PartitioningEngine

	// A value will be pushed to this channel when the status of the execution of a test changes, e.g. via the test suite
	//  registering that execution has started, or the timeout has been hit, etc.
	testExecutionStatusChan chan test_execution_status.TestExecutionStatus

	// The names of the tests inside the suite; will be nil if no test suite has been registered yet
	suiteTestNames []string

	// Flag that will only be switched to true once, indicating that a test execution has been registered
	testExecutionRegistered bool

	// Map of service ID -> container ID
	serviceContainerIds map[ServiceID]string

	mutex *sync.Mutex
}

func NewKurtosisService(
		testSuiteContainerId string,
		testExecutionStatusChan chan test_execution_status.TestExecutionStatus,
		dockerManager *commons.DockerManager,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		testVolumeName string,
		isPartitioningEnabled bool) *KurtosisService {
	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		dockerManager,
		freeIpAddrTracker,
		dockerNetworkId,
		testVolumeName)
	partitioningEngine := partitioning.NewPartitioningEngine(
		isPartitioningEnabled,
		dockerNetworkId,
		freeIpAddrTracker,
		dockerManager,
		userServiceLauncher)
	return &KurtosisService{
		testSuiteContainerId:    testSuiteContainerId,
		testExecutionStatusChan: testExecutionStatusChan,
		dockerManager:           dockerManager,
		partitioningEngine: partitioningEngine,
		suiteTestNames:          nil,
		mutex:                   &sync.Mutex{},
	}
}

/*
Adds a service with the given parameters to the network
 */
func (service *KurtosisService) AddService(httpReq *http.Request, args *AddServiceArgs, result *AddServiceResponse) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	logrus.Infof("Received request to add a service with the following args: %v", *args)

	requestedServiceId := ServiceID(args.ServiceID)
	if _, found := service.serviceContainerIds[requestedServiceId]; found {
		return stacktrace.NewError("Could not create service with ID '%v'; ID is already in use", requestedServiceId)
	}

	usedPorts := map[nat.Port]bool{}
	for _, portSpecStr := range args.UsedPorts {
		// NOTE: this function, frustratingly, doesn't return an error on failure - just emptystring
		protocol, portNumberStr := nat.SplitProtoPort(portSpecStr)
		if protocol == "" {
			return stacktrace.NewError(
				"Could not split port specification string '%s' into protocol & number strings",
				portSpecStr)
		}
		portObj, err := nat.NewPort(protocol, portNumberStr)
		if err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred constructing a port object out of protocol '%v' and port number string '%v'",
				protocol,
				portNumberStr)
		}
		usedPorts[portObj] = true
	}

	service.partitioningEngine.CreateServiceInPartition(
		httpReq.Context(),
		)



	logrus.Infof("Successfully added service")
	return nil
}

/*
Removes the service with the given service ID from the network
 */
func (service *KurtosisService) RemoveService(httpReq *http.Request, args *RemoveServiceArgs, result *interface{}) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	serviceId := ServiceID(args.ServiceID)
	containerId, found := service.serviceContainerIds[serviceId]
	if !found {
		return stacktrace.NewError("Could not remove service with ID '%v'; no such ID exists", serviceId)
	}

	logrus.Debugf("Removing service ID '%v' with container ID '%v'...", containerId)

	// Make a best-effort attempt to stop the container
	err := service.dockerManager.StopContainer(httpReq.Context(), containerId, containerStopTimeout)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", containerId)
	}
	logrus.Debugf("Successfully removed service with container ID %v", containerId)

	// TODO need to stop the iproute2 sidecar container too!

	return nil
}

// Registers that the test suite container is going to run a test, and the Kurtosis API container should wait for the
//  given amount of time before calling the test lost
func (service *KurtosisService) RegisterTestExecution(httpReq *http.Request, args *RegisterTestExecutionArgs, result *struct{}) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	logrus.Infof("Received request to register a test execution with timeout of %v seconds...", args.TestTimeoutSeconds)
	if service.testExecutionRegistered {
		return stacktrace.NewError("A test execution is already registered with the API container")
	}

	service.testExecutionStatusChan <- test_execution_status.Running
	service.testExecutionRegistered = true

	logrus.Info("Launching thread to await test completion or timeout...")
	go awaitTestCompletionOrTimeout(service.dockerManager, service.testSuiteContainerId, args.TestTimeoutSeconds, service.testExecutionStatusChan)
	logrus.Info("Launched thread to await test completion or timeout")
	return nil
}


// ============================ Private helper functions ==============================================================
/*
Waits for either a) the testsuite container to exit or b) the given timeout to be reached, and pushes a corresponding
	boolean value to the given channel based on which condition was hit
 */
func awaitTestCompletionOrTimeout(
			dockerManager *commons.DockerManager,
			testSuiteContainerId string,
			timeoutSeconds int,
			testExecutionStatusChan chan test_execution_status.TestExecutionStatus) {
	logrus.Debugf("[%v] Thread started", completionTimeoutPrefix)

	context, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Kick off a thread that will only exit upon a) the testsuite container exiting or b) the context getting cancelled
	testSuiteContainerExitedChan := make(chan struct{})

	logrus.Debugf("[%v] Launching thread to await testsuite container exit...", completionTimeoutPrefix)
	go awaitTestSuiteContainerExit(context, dockerManager, testSuiteContainerId, testSuiteContainerExitedChan)
	logrus.Debugf("[%v] Launched thread to await testsuite container exit", completionTimeoutPrefix)

	logrus.Debugf("[%v] Blocking until either the test suite exits or the test timeout is hit...", completionTimeoutPrefix)
	timeout := time.Duration(timeoutSeconds) * time.Second
	select {
	case <- testSuiteContainerExitedChan:
		// Triggered when the channel is closed, which is our signal that the testsuite container exited
		logrus.Debugf("[%v] Received signal that test suite container exited", completionTimeoutPrefix)
		testExecutionStatusChan <- test_execution_status.CompletedBeforeTimeout
	case <- time.After(timeout):
		logrus.Debugf("[%v] Hit test timeout (%v) before test suite container exited", completionTimeoutPrefix, timeout)
		testExecutionStatusChan <- test_execution_status.HitTimeout
		cancelFunc() // We hit the timeout, so tell the container-awaiting thread to hara-kiri
	}
	close(testExecutionStatusChan)
	logrus.Debugf("[%v] Thread is exiting", completionTimeoutPrefix)
}

/*
Waits for the container to exit until the context is cancelled
 */
func awaitTestSuiteContainerExit(
		context context.Context,
		dockerManager *commons.DockerManager,
		testSuiteContainerId string,
		testSuiteContainerExitedChan chan struct{}) {
	logrus.Debugf("[%v] Thread started", completionPrefix)

	_, err := dockerManager.WaitForExit(context, testSuiteContainerId)
	if err != nil {
		logrus.Debugf(
			"[%v] Got an error while waiting for the testsuite container to exit, likely indicating that the timeout was hit and the context was cancelled",
			completionPrefix)
	} else {
		logrus.Debugf("[%v] The test suite container has exited", completionPrefix)
	}

	// If we get to here before the timeout, this will signal that the testsuite container exited; if we get to here
	//  after the timeout, this won't do anything because nobody will be monitoring the other end of the channel
	close(testSuiteContainerExitedChan)
	logrus.Debugf("[%v] Thread is exiting", completionPrefix)
}
