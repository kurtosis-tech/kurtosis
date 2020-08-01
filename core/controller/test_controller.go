package controller

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	// How long to wait before force-killing a container
	CONTAINER_STOP_TIMEOUT = 30 * time.Second
)

/*
The user is responsible for providing a Docker image that runs a CLI which in turn does test setup and execution. The user
	will need to define their own CLI class because they might need custom parameters, but this struct will do all the
	heavy lifting of test setup and execution.
 */
type TestController struct {
	// The name of the test Docker volume that will have already been mounted on the controller image by the Kurtosis initializer
	testVolumeName string

	// The location on the controller image where the test volume will have been mounted by the Kurtosis initializer
	testVolumeFilepath string

	// The ID of the Docker network that the controller container is running inside
	networkId string

	// The subnet mask of the Docker network that the controller container is running inside
	subnetMask string

	// The gateway IP of the Docker network that the controller container is running inside
	gatewayIp string

	// The IP address of the Docker container running the controller code itself (this code)
	testControllerIp string

	// The user-defined test suite containing all the user's tests
	testSuite testsuite.TestSuite

	// The name of the specific test this controller is responsible for running (since there's a 1:1 mapping between controller
	// 	and test to execute
	testName string
}

/*
Creates a new TestController with the given properties. All of these parameters should be passed from the Kurtosis initializer
	to the user's CLI in the form of Docker environment variables.

Args:
	testVolumeName: The name of the Docker volume where test data should be stored, which will have been mounted on
		the controller by the initializer and should be mounted on service nodes
	testVolumeFilepath: The filepath where the test volume will have been mounted on the controller container by the initializer
	networkId: The ID of the Docker network that the controller container is running in and which all
		services should be started in
	subnetMask: Mask of the network that the controller container is running in, and from which it should dole out
		IPs to the testnet containers
	gatewayIp: The IP of the gateway that's running the Docker network that the controller container is running in, and
		which test network services will be started in
	testControllerIp: The IP address of the controller container itself
	testSuite: A pre-defined set of tests that the user will choose to run a single test from
	testName: The name of the test to run in the test suite
 */
func NewTestController(
			testVolumeName string,
			testVolumeFilepath string,
			networkId string,
			subnetMask string,
			gatewayIp string,
			testControllerIp string,
			testSuite testsuite.TestSuite,
			testName string) *TestController {
	return &TestController{
		testVolumeName:     testVolumeName,
		testVolumeFilepath: testVolumeFilepath,
		networkId:          networkId,
		subnetMask:         subnetMask,
		gatewayIp:          gatewayIp,
		testControllerIp:   testControllerIp,
		testSuite:          testSuite,
		testName:           testName,
	}
}

/*
Runs the single test from the test suite that the controller is configured to run.

Returns:
	setupErr: Indicates an error setting up the test that prevented the test from running
	testErr: Indicates an error in the test itself, indicating a test failure
 */
func (controller TestController) RunTest() (setupErr error, testErr error) {
	tests := controller.testSuite.GetTests()
	logrus.Debugf("Test configs: %v", tests)
	test, found := tests[controller.testName]
	if !found {
		return stacktrace.NewError("Nonexistent test: %v", controller.testName), nil
	}

	networkLoader, err := test.GetNetworkLoader()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get network loader"), nil
	}

	logrus.Info("Connecting to Docker environment...")
	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err,"Failed to initialize Docker client from environment."), nil
	}
	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when constructing the Docker manager"), nil
	}
	logrus.Info("Connected to Docker environment")

	logrus.Infof("Configuring test network in Docker network %v...", controller.networkId)
	alreadyTakenIps := map[string]bool{
		controller.gatewayIp: true,
		controller.testControllerIp: true,
	}
	freeIpTracker, err := networks.NewFreeIpAddrTracker(logrus.StandardLogger(), controller.subnetMask, alreadyTakenIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the free IP address tracker"), nil
	}

	builder := networks.NewServiceNetworkBuilder(
			dockerManager,
			controller.networkId,
			freeIpTracker,
			controller.testVolumeName,
			controller.testVolumeFilepath)
	if err := networkLoader.ConfigureNetwork(builder); err != nil {
		return stacktrace.Propagate(err, "Could not configure test network in Docker network %v", controller.networkId), nil
	}
	network := builder.Build()
	defer func() {
		logrus.Info("Stopping test network...")
		err := network.RemoveAll(CONTAINER_STOP_TIMEOUT)
		if err != nil {
			logrus.Error("An error occurred stopping the network")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		} else {
			logrus.Info("Successfully stopped the test network")
		}
	}()
	logrus.Info("Test network configured")

	logrus.Info("Initializing test network...")
	availabilityCheckers, err := networkLoader.InitializeNetwork(network);
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred initialized the network to its starting state"), nil
	}
	logrus.Info("Test network initialized")

	// Second pass: wait for all services to come up
	logrus.Info("Waiting for test network to become available...")
	for serviceId, availabilityChecker := range availabilityCheckers {
		logrus.Debugf("Waiting for service %v to become available...", serviceId)
		if err := availabilityChecker.WaitForStartup(); err != nil {
			return stacktrace.Propagate(err, "An error occurred waiting for service with ID %v to start up", serviceId), nil
		}
		logrus.Debugf("Service %v is available", serviceId)
	}
	logrus.Info("Test network is available")

	logrus.Info("Executing test...")
	untypedNetwork, err := networkLoader.WrapNetwork(network)
	if err != nil {
		return stacktrace.Propagate(err, "Error occurred wrapping network in user-defined network type"), nil
	}

	testResultChan := make(chan error)

	go func() {
		testResultChan <- runTest(test, untypedNetwork)
	}()

	// Time out the test so a poorly-written test doesn't run forever
	testTimeout := test.GetExecutionTimeout()
	var timedOut bool
	var testResultErr error
	select {
	case testResultErr = <- testResultChan:
		logrus.Tracef("Test returned result before timeout: %v", testResultErr)
		timedOut = false
	case <- time.After(testTimeout):
		logrus.Tracef("Hit timeout %v before getting a result from the test", testTimeout)
		timedOut = true
	}

	logrus.Tracef("After running test w/timeout: resultErr: %v, timedOut: %v", testResultErr, timedOut)

	if timedOut {
		return nil, stacktrace.NewError("Timed out after %v waiting for test to complete", testTimeout)
	}

	logrus.Info("Test execution completed")

	if testResultErr != nil {
		return nil, stacktrace.Propagate(testResultErr, "An error occurred when running the test")
	}

	return nil, nil
}

// Little helper function meant to be run inside a goroutine that runs the test
func runTest(test testsuite.Test, untypedNetwork interface{}) (resultErr error) {
	// See https://medium.com/@hussachai/error-handling-in-go-a-quick-opinionated-guide-9199dd7c7f76 for details
	defer func() {
		if recoverResult := recover(); recoverResult != nil {
			logrus.Tracef("Caught panic while running test: %v", recoverResult)
			resultErr = recoverResult.(error)
		}
	}()
	test.Run(untypedNetwork, testsuite.TestContext{})
	logrus.Tracef("Test completed successfully")
	return
}
