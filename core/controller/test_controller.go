package controller

import (
	"context"
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

type TestController struct {
	testVolumeName string
	testVolumeFilepath string
	networkName string
	subnetMask string
	gatewayIp string
	testControllerIp string
	testSuite testsuite.TestSuite
	testImageName string
}

/*
Creates a new TestController with the given properties
Args:
	testVolumeName: The name of the volume where test data should be stored, which will have been mountd on the controller by the initializer and should be mounted on service nodes
	testVolumeFilepath: The filepath where the test volume will have been mounted on the controller by the initializer
	networkName: The name of the Docker network that the test controller is running in and which all services should be started in
	subnetMask: Mask of the network that the TestController is living in, and from which it should dole out IPs to the testnet containers
	gatewayIp: The IP of the gateway that's running the Docker network that the TestController and the containers run in
	testControllerIp: The IP address of the controller itself
	testSuite: A pre-defined set of tests that the user will choose to run a single test from
	testImageName: The Docker image representing the version of the node that is being tested
 */
func NewTestController(
			testVolumeName string,
			testVolumeFilepath string,
			networkName string,
			subnetMask string,
			gatewayIp string,
			testControllerIp string,
			testSuite testsuite.TestSuite,
			testImageName string) *TestController {
	return &TestController{
		testVolumeName: testVolumeName,
		testVolumeFilepath: testVolumeFilepath,
		networkName:      networkName,
		subnetMask:       subnetMask,
		gatewayIp:        gatewayIp,
		testControllerIp: testControllerIp,
		testSuite:        testSuite,
		testImageName:    testImageName,
	}
}

/*
Creates a new TestController that will run one of the tests from the test suite given at construction time
Args:
	testName: Name of test to run
	testImageName: Name of the Docker image representing a node being tested

Returns:
	setupErr: Indicates an error setting up the test that prevented the test from running
	testErr: Indicates an error in the test itself, indicating a test failure
 */
func (controller TestController) RunTest(testName string) (setupErr error, testErr error) {
	tests := controller.testSuite.GetTests()
	logrus.Debugf("Test configs: %v", tests)
	test, found := tests[testName]
	if !found {
		return stacktrace.NewError("Nonexistent test: %v", testName), nil
	}

	networkLoader, err := test.GetNetworkLoader()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get network loader"), nil
	}

	logrus.Info("Connecting to Docker environment...")
	// Initialize default environment context.
	dockerCtx := context.Background()
	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err,"Failed to initialize Docker client from environment."), nil
	}
	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when constructing the Docker manager"), nil
	}
	logrus.Info("Connected to Docker environment")

	logrus.Infof("Configuring test network in Docker network %v...", controller.networkName)
	alreadyTakenIps := []string{controller.gatewayIp, controller.testControllerIp}
	freeIpTracker, err := networks.NewFreeIpAddrTracker(controller.subnetMask, alreadyTakenIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the free IP address tracker"), nil
	}

	builder := networks.NewServiceNetworkBuilder(
			controller.testImageName,
			dockerManager,
			controller.networkName,
			freeIpTracker,
			controller.testVolumeName,
			controller.testVolumeFilepath)
	if err := networkLoader.ConfigureNetwork(builder); err != nil {
		return stacktrace.Propagate(err, "Could not configure test network in Docker network %v", controller.networkName), nil
	}
	network := builder.Build()
	defer func() {
		logrus.Info("Stopping test network...")
		err := network.RemoveAll(CONTAINER_STOP_TIMEOUT)
		if err != nil {
			logrus.Error("An error occurred stopping the network")
			logrus.Error(err)
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
	testTimeout := test.GetTimeout()
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
