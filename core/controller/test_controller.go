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
	DEFAULT_SUBNET_MASK = "172.23.0.0/16"
)

type TestController struct {
	subnetMask string
	gatewayIp string
	testControllerIp string
	testSuite testsuite.TestSuite
	testImageName string
}

/*
Creates a new TestController with the given properties
Args:
	subnetMask: Mask of the network that the TestController is living in, and from which it should dole out IPs to the testnet containers
	gatewayIp: The IP of the gateway that's running the Docker network that the TestController and the containers run in
	testControllerIp: The IP address of the controller itself
	testSuite: A pre-defined set of tests that the user will choose to run a single test from
	testImageName: The Docker image representing the version of the node that is being tested
 */
func NewTestController(
			subnetMask string,
			gatewayIp string,
			testControllerIp string,
			testSuite testsuite.TestSuite,
			testImageName string) *TestController {
	return &TestController{
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
 */
func (controller TestController) RunTest(testName string, ) (error) {
	tests := controller.testSuite.GetTests()
	logrus.Debugf("Test configs: %v", tests)
	test, found := tests[testName]
	if !found {
		return stacktrace.NewError("Nonexistent test: %v", testName)
	}

	networkLoader, err := test.GetNetworkLoader()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get network loader")
	}

	// Initialize default environment context.
	dockerCtx := context.Background()
	// Initialize a Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err,"Failed to initialize Docker client from environment.")
	}
	dockerManager, err := docker.NewDockerManager(dockerCtx, dockerClient)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when constructing the Docker manager")
	}

	alreadyTakenIps := []string{controller.gatewayIp, controller.testControllerIp}
	freeIpTracker, err := networks.NewFreeIpAddrTracker(controller.subnetMask, alreadyTakenIps)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	builder := networks.NewServiceNetworkBuilder(testName, dockerManager, freeIpTracker)
	if err := networkLoader.ConfigureNetwork(builder); err != nil {
		return stacktrace.Propagate(err, "Could not configure network")
	}
	network := builder.Build()

	availabilityCheckers, err := networkLoader.BootstrapNetwork(network);
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred bootstrapping the network to its starting state")
	}

	serviceNetwork, err := networkCfg.LoadNetwork(rawServiceNetwork)
	if err != nil {
		return stacktrace.Propagate(err, "Could not load network from raw service information")
	}

	untypedNetwork, err := networkLoader.WrapNetwork(serviceNetwork)
	if err != nil {
		return stacktrace.Propagate(err, "Error occurred wrapping network in user-defined network type")
	}

	testResultChan := make(chan error)

	go func() {
		testResultChan <- runTest(test, untypedNetwork)
	}()

	// Time out the test so a poorly-written test doesn't run forever
	testTimeout := test.GetTimeout()
	var timedOut bool
	var resultErr error
	select {
	case resultErr = <- testResultChan:
		timedOut = false
	case <- time.After(testTimeout):
		timedOut = true
	}

	if timedOut {
		return stacktrace.NewError("Timed out after %v waiting for test to complete", testTimeout)
	}

	if resultErr != nil {
		return stacktrace.Propagate(err, "An error occurred when running the test")
	}

	// Should we return a TestSuiteResults object that provides detailed info about each test?
	return nil
}

// Little helper function meant to be run inside a goroutine that runs the test
func runTest(test testsuite.Test, untypedNetwork interface{}) (resultErr error) {
	defer func() {
		if recoverResult := recover(); recoverResult != nil {
			resultErr = recoverResult.(error)
		}
	}()
	test.Run(untypedNetwork, testsuite.TestContext{})
	return nil
}
