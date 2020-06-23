package controller

import (
	"encoding/gob"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type TestController struct {
	testSuite testsuite.TestSuite
}

func NewTestController(testSuite testsuite.TestSuite) *TestController {
	return &TestController{testSuite: testSuite}
}

func (controller TestController) RunTest(testName string, networkInfoFilepath string) (error) {
	tests := controller.testSuite.GetTests()
	logrus.Debugf("Test configs: %v", tests)
	test, found := tests[testName]
	if !found {
		return stacktrace.NewError("Nonexistent test: %v", testName)
	}

	if _, err := os.Stat(networkInfoFilepath); err != nil {
		return stacktrace.Propagate(err, "Nonexistent file: %v", networkInfoFilepath)
	}

	fp, err := os.Open(networkInfoFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "Could not open file for reading: %v", networkInfoFilepath)
	}
	decoder := gob.NewDecoder(fp)

	var rawServiceNetwork networks.RawServiceNetwork
	err = decoder.Decode(&rawServiceNetwork)
	if err != nil {
		return stacktrace.Propagate(err, "Decoding raw service network information failed for file: %v", networkInfoFilepath)
	}

	networkLoader, err := test.GetNetworkLoader()
	if err != nil {
		return stacktrace.Propagate(err, "Could not get network loader")
	}

	builder := networks.NewServiceNetworkConfigBuilder()
	if err := networkLoader.ConfigureNetwork(builder); err != nil {
		return stacktrace.Propagate(err, "Could not configure network")
	}
	networkCfg := builder.Build()

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
		logrus.Tracef("Test returned result before timeout: %v", resultErr)
		timedOut = false
	case <- time.After(testTimeout):
		logrus.Tracef("Hit timeout %v before getting a result from the test", testTimeout)
		timedOut = true
	}

	logrus.Tracef("After running test w/timeout: resultErr: %v, timedOut: %v", resultErr, timedOut)

	if timedOut {
		return stacktrace.NewError("Timed out after %v waiting for test to complete", testTimeout)
	}

	if resultErr != nil {
		return stacktrace.Propagate(resultErr, "An error occurred when running the test")
	}

	// Should we return a TestSuiteResults object that provides detailed info about each test?
	return nil
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
