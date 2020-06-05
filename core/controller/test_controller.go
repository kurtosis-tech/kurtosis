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

	networkLoader, err := test.GetNetworkLoader()
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not get network loader")
	}

	builder := networks.NewServiceNetworkConfigBuilder()
	if err := networkLoader.ConfigureNetwork(builder); err != nil {
		return false, stacktrace.Propagate(err, "Could not configure network")
	}
	networkCfg := builder.Build()

	serviceNetwork, err := networkCfg.LoadNetwork(rawServiceNetwork)
	if err != nil {
		return false, stacktrace.Propagate(err, "Could not load network from raw service information")
	}
	untypedNetwork, err := networkLoader.WrapNetwork(serviceNetwork)

	testSucceeded := true
	context := testsuite.TestContext{}

	// TODO test that this panic recovery actually works!
	defer func() {
		if result := recover(); result != nil {
			logrus.Error(stacktrace.Propagate(err, "Error when running test '%v'", testName))
			testSucceeded = false
		}
	}()
	test.Run(untypedNetwork, context)

	// Should we return a TestSuiteResults object that provides detailed info about each test?
	return testSucceeded, nil
}
