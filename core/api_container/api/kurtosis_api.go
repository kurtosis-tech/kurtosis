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
)

type KurtosisAPI struct {
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

func NewKurtosisAPI(
		testSuiteContainerId string,
		testEndedBeforeTimeoutChan chan bool,
		dockerManager *docker.DockerManager,
		dockerNetworkId string,
		freeIpAddrTracker *networks.FreeIpAddrTracker) *KurtosisAPI {
	return &KurtosisAPI{
		testSuiteContainerId: testSuiteContainerId,
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
func (api *KurtosisAPI) AddService(httpReq *http.Request, args *AddServiceArgs, result *AddServiceResponse) error {
	// TODO Add a UUID for the request so we can trace it through???
	logrus.Tracef("Received request: %v", *httpReq)

	usedPorts := map[nat.Port]bool{}
	for _, portInt := range args.UsedPorts {
		// TODO add ability to have non-TCP ports
		castedPort := nat.Port(fmt.Sprintf("%v/tcp", portInt))
		usedPorts[castedPort] = true
	}

	freeIp, err := api.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		// TODO give more identifying container details in the log message
		return stacktrace.Propagate(err, "Could not get a free IP to assign the new container")
	}

	// TODO Debug log message saying what we're going to do

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

	containerId, err := api.dockerManager.CreateAndStartContainer(
		httpReq.Context(),
		args.ImageName,
		api.dockerNetworkId,
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

	return nil
}

/*
Removes the service with the given service ID from the network
 */
func (api *KurtosisAPI) RemoveService(httpReq *http.Request, args *RemoveServiceArgs, result *interface{}) error {
	containerId := args.ContainerID
	logrus.Debugf("Removing container ID %v...", containerId)

	// Make a best-effort attempt to stop the container
	err := api.dockerManager.StopContainer(httpReq.Context(), containerId, containerStopTimeout)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the container with ID %v", containerId)
	}
	logrus.Debugf("Successfully removed service with container ID %v", containerId)

	return nil
}

// Registers that the test suite container is going to run a test, and the Kurtosis API container should wait for the
//  given amount of time before calling the test lost
func (api *KurtosisAPI) RegisterTestExecution(httpReq *http.Request, args *RegisterTestExecutionArgs, result *struct{}) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.testExecutionRegistered {
		return stacktrace.NewError("A test execution is already registered with the API container")
	}
	api.testExecutionRegistered = true
	go awaitTestCompletionOrTimeout(api.dockerManager, api.testSuiteContainerId, args.TestTimeoutSeconds, api.testEndedBeforeTimeoutChan)
	return nil
}


// ============================ Private helper functions ==============================================================
/*
Waits for either a) the testsuite container to exit or b) the given timeout to be reached, and pushes a corresponding
	boolean value to the given channel based on which condition was hit
 */
func awaitTestCompletionOrTimeout(dockerManager *docker.DockerManager, testSuiteContainerId string, timeoutSeconds int, testEndedBeforeTimeoutChan chan bool) {
	context, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	// Kick off a thread that will only exit upon a) the testsuite contianer exiting or b) the context getting cancelled
	testSuiteContainerExitedChan := make(chan struct{})
	go awaitTestSuiteContainerExit(context, dockerManager, testSuiteContainerId, testSuiteContainerExitedChan)

	timeout := time.Duration(timeoutSeconds) * time.Second
	select {
	case <- testSuiteContainerExitedChan:
		// Triggered when the channel is closed, which is our signal that the testsuite container exited
		testEndedBeforeTimeoutChan <- true
	case <- time.After(timeout):
		testEndedBeforeTimeoutChan <- false
		cancelFunc() // We hit the timeout, so tell the container-awaiting thread to hara-kiri
	}
	close(testEndedBeforeTimeoutChan)
}

/*
Waits for the container to exit until the context is cancelled
 */
func awaitTestSuiteContainerExit(
		context context.Context,
		dockerManager *docker.DockerManager,
		testSuiteContainerId string,
		testSuiteContainerExitedChan chan struct{}) {
	_, err := dockerManager.WaitForExit(context, testSuiteContainerId)

	if err != nil {
		logrus.Debug("Got an error while waiting for the testsuite container to exit, likely indicating that the timeout was hit and the context was cancelled")
	}

	// If we get to here before the timeout, this will signal that the testsuite container exited; if we get to here
	//  after the timeout, this won't do anything because nobody will be monitoring the other end of the channel
	close(testSuiteContainerExitedChan)
}
