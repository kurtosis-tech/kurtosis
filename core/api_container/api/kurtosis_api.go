package api

import (
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
)

type KurtosisAPI struct {
	dockerManager *docker.DockerManager
	dockerNetworkId string
	freeIpAddrTracker *networks.FreeIpAddrTracker

	// The Docker container ID of the test suite that will be making calls against this Kurtosis API
	// This is expected to be nil until the "register test suite container" endpoint is called, which should be called
	//  exactly once
	testSuiteContainerId *string
}

func NewKurtosisAPI(
		dockerManager *docker.DockerManager,
		dockerNetworkId string,
		freeIpAddrTracker *networks.FreeIpAddrTracker) *KurtosisAPI {
	return &KurtosisAPI{
		dockerManager:        dockerManager,
		dockerNetworkId:      dockerNetworkId,
		freeIpAddrTracker:    freeIpAddrTracker,
		testSuiteContainerId: nil,
	}
}

func (api *KurtosisAPI) StartService(httpReq *http.Request, args *StartServiceArgs, result *StartServiceResponse) error {
	// TODO Add a UUID for the request so we can trace it through???
	logrus.Tracef("Received request: %v", *httpReq)

	usedPorts := map[nat.Port]bool{}
	for _, portStr := range args.UsedPorts {
		// TODO validation that the port is legit
		castedPort := nat.Port(portStr)
		usedPorts[castedPort] = true
	}

	freeIp, err := api.freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		// TODO give more identifying container details in the log message
		return stacktrace.Propagate(err, "Could not get a free IP to assign the new container")
	}

	// TODO Debug log message saying what we're going to do

	_, err = api.dockerManager.CreateAndStartContainer(
		httpReq.Context(),
		args.ImageName,
		api.dockerNetworkId,
		freeIp,
		usedPorts,
		args.StartCmd,
		args.DockerEnvironmentVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		map[string]string{}, // TODO mount the test volume!
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Docker container for the new service")
	}

	// TODO Debug log message saying we started successfully
	result.IPAddress = freeIp.String()

	return nil
}

func (api *KurtosisAPI) RegisterTestSuite(httpReq *http.Request, args *RegisterTestSuiteContainerArgs, result *struct{}) error {
	// TODO kick off a thread to trigger the hard test timeout (prob with a Go channel)

	// TODO validation on container ID
	// Strings are read-only so it's okay to do this
	api.testSuiteContainerId = &(args.ContainerId)
	return nil
}
