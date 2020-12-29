/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package api

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/execution/test_execution_status"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/service_engine/topology_types"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

const (

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

	serviceNetworkEngine *service_engine.ServiceNetworkEngine

	// A value will be pushed to this channel when the status of the execution of a test changes, e.g. via the test suite
	//  registering that execution has started, or the timeout has been hit, etc.
	testExecutionStatusChan chan test_execution_status.TestExecutionStatus

	// Flag that will only be switched to true once, indicating that a test execution has been registered
	testExecutionRegistered bool
}

func NewKurtosisService(
		testSuiteContainerId string,
		testExecutionStatusChan chan test_execution_status.TestExecutionStatus,
		dockerManager *commons.DockerManager,
		serviceNetworkEngine *service_engine.ServiceNetworkEngine) *KurtosisService {

	return &KurtosisService{
		testSuiteContainerId:    testSuiteContainerId,
		dockerManager:           dockerManager,
		serviceNetworkEngine:    serviceNetworkEngine,
		testExecutionStatusChan: testExecutionStatusChan,
		testExecutionRegistered: false,
	}
}

/*
Adds a service with the given parameters to the network
 */
func (service *KurtosisService) AddService(httpReq *http.Request, args *AddServiceArgs, result *AddServiceResponse) error {
	logrus.Infof("Received request to add a service with the following args: %v", *args)

	serviceIdStr := strings.TrimSpace(args.ServiceID)
	if serviceIdStr == "" {
		return stacktrace.NewError("Service ID cannot be empty or whitespace")
	}
	serviceId := topology_types.ServiceID(serviceIdStr)

	partitionId := topology_types.PartitionID(args.PartitionID)

	imageNameStr := strings.TrimSpace(args.ImageName)
	if imageNameStr == "" {
		return stacktrace.NewError("Image name cannot be empty or whitespace")
	}

	ipPlaceholderStr := strings.TrimSpace(args.IPPlaceholder)
	if ipPlaceholderStr == "" {
		return stacktrace.NewError("IP placeholder string cannot be empty or whitespace")
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

	serviceIp, err := service.serviceNetworkEngine.AddServiceInPartition(
		httpReq.Context(),
		serviceId,
		imageNameStr,
		usedPorts,
		partitionId,
		ipPlaceholderStr,
		args.StartCmd,
		args.DockerEnvironmentVars,
		args.TestVolumeMountDirpath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating service '%v' inside partition '%v'", serviceId, partitionId)
	}
	logrus.Infof("Successfully added service '%v'", serviceId)

	result.IPAddress = serviceIp.String()
	return nil
}

/*
Removes the service with the given service ID from the network
 */
func (service *KurtosisService) RemoveService(httpReq *http.Request, args *RemoveServiceArgs, _ *interface{}) error {
	logrus.Infof("Received request to remove a service with the following args: %v", *args)

	serviceIdStr := strings.TrimSpace(args.ServiceID)
	if serviceIdStr == "" {
		return stacktrace.NewError("Service ID cannot be empty or whitespace")
	}
	serviceId := topology_types.ServiceID(serviceIdStr)

	if args.ContainerStopTimeoutSeconds <= 0 {
		return stacktrace.NewError("Container stop timeout seconds cannot be <= 0")
	}

	containerStopTimeout := time.Duration(args.ContainerStopTimeoutSeconds) * time.Second
	if err := service.serviceNetworkEngine.RemoveService(httpReq.Context(), serviceId, containerStopTimeout); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}

	return nil
}

func (service *KurtosisService) Repartition(httpReq *http.Request, args *RepartitionArgs, _ *interface{}) error {
	logrus.Infof("Received request to repartition the test network with the following args: %v", args)

	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetworkEngine.Repartition (as it should)
	partitionServices := map[topology_types.PartitionID]*topology_types.ServiceIDSet{}
	for partitionIdStr, serviceIdStrSet := range args.PartitionServices {
		partitionId := topology_types.PartitionID(partitionIdStr)
		serviceIdSet := topology_types.NewServiceIDSet()
		for serviceIdStr := range serviceIdStrSet {
			serviceId := topology_types.ServiceID(serviceIdStr)
			serviceIdSet.AddElem(serviceId)
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[topology_types.PartitionConnectionID]partition_topology.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := topology_types.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection {
			partitionBId := topology_types.PartitionID(partitionBStr)
			partitionConnectionId := *topology_types.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology.PartitionConnection{
				IsBlocked: connectionInfo.IsBlocked,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		IsBlocked: defaultConnectionInfo.IsBlocked,
	}

	if err := service.serviceNetworkEngine.Repartition(
			httpReq.Context(),
			partitionServices,
			partitionConnections,
			defaultConnection); err != nil {
		return stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return nil
}


// Registers that the test suite container is going to run a test, and the Kurtosis API container should wait for the
//  given amount of time before calling the test lost
func (service *KurtosisService) RegisterTestExecution(_ *http.Request, args *RegisterTestExecutionArgs, _ *struct{}) error {
	logrus.Infof("Received request to register a test execution with timeout of %v seconds", args.TestTimeoutSeconds)

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

	cancellableCtx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Kick off a thread that will only exit upon a) the testsuite container exiting or b) the context getting cancelled
	testSuiteContainerExitedChan := make(chan struct{})

	logrus.Debugf("[%v] Launching thread to await testsuite container exit...", completionTimeoutPrefix)
	go awaitTestSuiteContainerExit(cancellableCtx, dockerManager, testSuiteContainerId, testSuiteContainerExitedChan)
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
