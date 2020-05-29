package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/ava_commons/testsuite"
	"github.com/kurtosis-tech/kurtosis/initializer"
)


const DEFAULT_STARTING_PORT = 9650
const DEFAULT_ENDING_PORT = 10650

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"",
		"the name of a pre-built gecko image in your docker engine.",
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
		*portRangeStartArg,
		*portRangeEndArg)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko testnet, and hang while it's running! Kill me and then clear your docker containers.")
	error := testSuiteRunner.RunTests()
	if error != nil {
		panic(error)
	}
}
