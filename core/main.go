package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"github.com/kurtosis-tech/kurtosis/initializer"
	"github.com/sirupsen/logrus"
)


const DEFAULT_STARTING_PORT = 9650
const DEFAULT_ENDING_PORT = 9670

func main() {
	// TODO make this configurable
	logrus.SetLevel(logrus.TraceLevel)

	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"gecko-f290f73", // by default, pick commit that was on master May 14, 2020.
		"The name of a pre-built Gecko image, either on the local Docker engine or in Docker Hub",
	)

	testControllerImageNameArg := flag.String(
		"test-controller-image-name",
		"ava-test-controller",
		"The name of a pre-built test controller image, either on the local Docker engine or in Docker Hub",
	)
	portRangeStartArg := flag.Int(
		"port-range-start",
		DEFAULT_STARTING_PORT,
		"Beginning of port range to be used by testnet on the local environment. Must be between 1024-65535",
	)

	portRangeEndArg := flag.Int(
		"port-range-end",
		DEFAULT_ENDING_PORT,
		"End of port range to be used by testnet on the local environment. Must be between 1024-65535",
	)

	flag.Parse()

	testSuiteRunner := initializer.NewTestSuiteRunner(
		testsuite.AvaTestSuite{},
		*geckoImageNameArg,
		*testControllerImageNameArg,
		*portRangeStartArg,
		*portRangeEndArg)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko testnet, and hang while it's running! Kill me and then clear your docker containers.")
	error := testSuiteRunner.RunTests()
	if error != nil {
		panic(error)
	}
}
