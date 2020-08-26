package api

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// TODO Move this somewhere else?
	KurtosisAPIContainerPort = 7443


	// How long we'll wait when making a best-effort attempt to stop a container
	containerStopTimeout = 15 * time.Second

	// Prefixes for the two types of threads we'll spin up
	completionTimeoutPrefix = "Test completion/timeout thread"
	completionPrefix        = "Test completion thread"
)

// TODO for all the methods on this class, make the logging log with a request ID
type KurtosisService struct {
	// The Docker container ID of the test suite that will be making calls against this Kurtosis API
	// This is expected to be nil until the "register test suite container" endpoint is called, which should be called
	//  exactly once
	testSuiteContainerId string

	dockerManager *docker.DockerManager

	dockerNetworkId string

	freeIpAddrTracker *networks.FreeIpAddrTracker

	// A value will be pushed to this channel iff a test execution has been registered with the API container and the
	//  execution either completes successfully (as indicated by the testsuite container exiting) or hits the timeout
	//  that was registered with the API container
	testEndedBeforeTimeoutChan chan bool

	// The names of the tests inside the suite; will be nil if no test suite has been registered yet
	suiteTestNames []string

	// Flag that will only be switched to true once, indicating that a test execution has been registered
	testExecutionRegistered bool

	mutex *sync.Mutex
}

func NewKurtosisService(
		testSuiteContainerId string,
		testEndedBeforeTimeoutChan chan bool,
		dockerManager *docker.DockerManager,
		dockerNetworkId string,
		freeIpAddrTracker *networks.FreeIpAddrTracker) *KurtosisService {
	return &KurtosisService{
		testSuiteContainerId:       testSuiteContainerId,
		testEndedBeforeTimeoutChan: testEndedBeforeTimeoutChan,
		dockerManager:              dockerManager,
		dockerNetworkId:            dockerNetworkId,
		freeIpAddrTracker:          freeIpAddrTracker,
		suiteTestNames:             nil,
		mutex:                      &sync.Mutex{},
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
	for _, portInt := range args.UsedPorts {
		// TODO add ability to have non-TCP ports
		castedPort := nat.Port(fmt.Sprintf("%v/tcp", portInt))
		usedPorts[castedPort] = true
	}

	freeIp, err := service.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		// TODO give more identifying container details in the log message
		return stacktrace.Propagate(err, "Could not get a free IP to assign the new container")
	}

	// The user won't know the IP address, so we'll need to replace all the IP address placeholders with the actual
	//  IP
	ipPlaceholderStr := args.IPPlaceholder
	replacedStartCmd := []string{}
	for _, cmdFragment := range args.StartCmd {
		replacedCmdFragment := strings.ReplaceAll(cmdFragment, ipPlaceholderStr, freeIp.String())
		replacedStartCmd = append(replacedStartCmd, replacedCmdFragment)
	}
	replacedEnvVars := map[string]string{}
	for key, value := range args.DockerEnvironmentVars {
		replacedValue := strings.ReplaceAll(value, ipPlaceholderStr, freeIp.String())
		replacedEnvVars[key] = replacedValue
	}

	containerId, err := service.dockerManager.CreateAndStartContainer(
		httpReq.Context(),
		args.ImageName,
		service.dockerNetworkId,
		freeIp,
		usedPorts,
		replacedStartCmd,
		replacedEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{}, // TODO mount the test volume!
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Docker container for the new service")
	}

	// TODO Debug log message saying we started successfully
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

	testTimeoutSeconds := args.TestTimeoutSeconds
	logrus.Infof("Received request to register a test execution with timeout of %v seconds...", testTimeoutSeconds)


	if service.testExecutionRegistered {
		return stacktrace.NewError("A test execution is already registered with the API container")
	}
	service.testExecutionRegistered = true
	logrus.Info("Launching thread to await test completion or timeout...")
	go awaitTestCompletionOrTimeout(service.dockerManager, service.testSuiteContainerId, testTimeoutSeconds, service.testEndedBeforeTimeoutChan)
	logrus.Info("Launched thread to await test completion or timeout")
	return nil
}


// ============================ Private helper functions ==============================================================
/*
Waits for either a) the testsuite container to exit or b) the given timeout to be reached, and pushes a corresponding
	boolean value to the given channel based on which condition was hit
 */
func awaitTestCompletionOrTimeout(dockerManager *docker.DockerManager, testSuiteContainerId string, timeoutSeconds int, testEndedBeforeTimeoutChan chan bool) {
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
		testEndedBeforeTimeoutChan <- true
	case <- time.After(timeout):
		logrus.Debugf("[%v] Hit test timeout (%v) before test suite container exited", completionTimeoutPrefix, timeout)
		testEndedBeforeTimeoutChan <- false
		cancelFunc() // We hit the timeout, so tell the container-awaiting thread to hara-kiri
	}
	close(testEndedBeforeTimeoutChan)
	logrus.Debugf("[%v] Thread is exiting", completionTimeoutPrefix)
}

/*
Waits for the container to exit until the context is cancelled
 */
func awaitTestSuiteContainerExit(
		context context.Context,
		dockerManager *docker.DockerManager,
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
