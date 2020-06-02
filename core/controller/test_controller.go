package controller

import (
	"encoding/gob"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/kurtosis-tech/kurtosis/commons/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
)

type TestController struct {
	testSuite testsuite.TestSuite
}

func NewTestController(testSuite testsuite.TestSuite) *TestController {
	return &TestController{testSuite: testSuite}
}

func (controller TestController) RunTests(testName string, networkInfoFilepath string) (bool, error) {
	// TODO create a TestSuiteContext object for returning the state of all the tests

	// TODO run multiple tests
	tests := controller.testSuite.GetTests()
	logrus.Debugf("Test configs: %v", tests)
	test, found := tests[testName]
	if !found {
		return false, stacktrace.NewError("Nonexistent test: %v", testName)
	}

	if _, err := os.Stat(networkInfoFilepath); err != nil {
		return false, stacktrace.Propagate(err, "Nonexistent file: %v", networkInfoFilepath)
	}

	fp, err := os.Open(networkInfoFilepath)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not open file for reading: %v", networkInfoFilepath)
	}
	decoder := gob.NewDecoder(fp)

	var rawServiceNetwork networks.RawServiceNetwork
	err = decoder.Decode(&rawServiceNetwork)
	if err != nil {
		return false, stacktrace.Propagate(err, "Decoding raw service network information failed for file: %v", networkInfoFilepath)
	}
	untypedNetwork, err := test.GetNetworkLoader().LoadNetwork(rawServiceNetwork.ServiceIPs)
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to load network from service IPs")
	}

	testSucceeded := true
	context := testsuite.TestContext{}
	test.Run(untypedNetwork, context)
	defer func() {
		if result := recover(); result != nil {
			testSucceeded = false
		}
	}()

	// TODO return a TestSuiteResults object that provides detailed info about each test?
	return testSucceeded, nil
}
