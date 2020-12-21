/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/test_execution_status"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
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

	dockerNetworkId string

	freeIpAddrTracker *commons.FreeIpAddrTracker

	// A value will be pushed to this channel when the status of the execution of a test changes, e.g. via the test suite
	//  registering that execution has started, or the timeout has been hit, etc.
	testExecutionStatusChan chan test_execution_status.TestExecutionStatus

	// The name of the Docker volume for this test that will be mounted on all services
	testVolumeName string

	// The names of the tests inside the suite; will be nil if no test suite has been registered yet
	suiteTestNames []string

	// Flag that will only be switched to true once, indicating that a test execution has been registered
	testExecutionRegistered bool

	mutex *sync.Mutex
}

func NewKurtosisService(
		testSuiteContainerId string,
		testExecutionStatusChan chan test_execution_status.TestExecutionStatus,
		dockerManager *commons.DockerManager,
		dockerNetworkId string,
		freeIpAddrTracker *commons.FreeIpAddrTracker,
		testVolumeName string) *KurtosisService {
	return &KurtosisService{
		testSuiteContainerId:    testSuiteContainerId,
		testExecutionStatusChan: testExecutionStatusChan,
		dockerManager:           dockerManager,
		dockerNetworkId:         dockerNetworkId,
		freeIpAddrTracker:       freeIpAddrTracker,
		testVolumeName: testVolumeName,
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

	freeIp, err := service.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred when getting an IP to give the container running the new service with Docker image '%v'",
			args.ImageName)
	}
	logrus.Debugf("Giving new service the following IP: %v", freeIp.String())

	// The user won't know the IP address, so we'll need to replace all the IP address placeholders with the actual
	//  IP
	replacedStartCmd, replacedEnvVars := replaceIpPlaceholderForDockerParams(
		args.IPPlaceholder,
		freeIp,
		args.StartCmd,
		args.DockerEnvironmentVars)

	portBindings := map[nat.Port]*nat.PortBinding{}
	for port, _ := range usedPorts {
		portBindings[port] = nil
	}

	containerId, err := service.dockerManager.CreateAndStartContainer(
		httpReq.Context(),
		args.ImageName,
		service.dockerNetworkId,
		freeIp,
		portBindings,
		replacedStartCmd,
		replacedEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{
			service.testVolumeName: args.TestVolumeMountFilepath,
		},
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Docker container for the new service")
	}

	result.IPAddress = freeIp.String()
	result.ContainerID = containerId

	logrus.Infof("Successfully added service")
	return nil
}

/*
Removes the service with the given service ID from the network
 */
func (service *KurtosisService) RemoveService(httpReq *http.Request, args *RemoveServiceArgs, result *interface{}) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	containerId := args.ContainerID
	logrus.Debugf("Removing container ID %v...", containerId)

	// Make a best-effort attempt to stop the container
	err := service.dockerManager.StopContainer(httpReq.Context(), containerId, containerStopTimeout)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", containerId)
	}
	logrus.Debugf("Successfully removed service with container ID %v", containerId)

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

/*
Small helper function to replace the IP placeholder with the real IP string in the start command and Docker environment
	variables.
 */
func replaceIpPlaceholderForDockerParams(
		ipPlaceholder string,
		realIp net.IP,
		startCmd []string,
		envVars map[string]string) ([]string, map[string]string) {
	ipPlaceholderStr := ipPlaceholder
	replacedStartCmd := []string{}
	for _, cmdFragment := range startCmd {
		replacedCmdFragment := strings.ReplaceAll(cmdFragment, ipPlaceholderStr, realIp.String())
		replacedStartCmd = append(replacedStartCmd, replacedCmdFragment)
	}
	replacedEnvVars := map[string]string{}
	for key, value := range envVars {
		replacedValue := strings.ReplaceAll(value, ipPlaceholderStr, realIp.String())
		replacedEnvVars[key] = replacedValue
	}
	return replacedStartCmd, replacedEnvVars
}
