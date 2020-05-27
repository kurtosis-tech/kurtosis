package main

import (
	"flag"
	"fmt"
	"github.com/gmarchetti/kurtosis/ava_commons/testsuite"
	"github.com/gmarchetti/kurtosis/initializer"
)

const DEFAULT_SUBNET_MASK = "172.23.0.0/16"
const DEFAULT_STARTING_PORT = 9650
const DEFAULT_ENDING_PORT = 9670

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"gecko-f290f73", // by default, pick commit that was on master May 14, 2020.
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

	subnetMaskArg := flag.String(
		"subnet-mask",
		DEFAULT_SUBNET_MASK,
		"Subnet mask used for assigning Public IPs to nodes. Must be a valid private internet address subnet according to RFC-1918",)
	flag.Parse()

	testSuiteRunner := initializer.NewTestSuiteRunner(
		testsuite.AvaTestSuite{},
		*geckoImageNameArg,
		*subnetMaskArg,
		*portRangeStartArg,
		*portRangeEndArg)

	// Create the container based on the configurations, but don't start it yet.
	fmt.Println("I'm going to run a Gecko testnet, and hang while it's running! Kill me and then clear your docker containers.")
	error := testSuiteRunner.RunTests()
	if error != nil {
		panic(error)
	}
}
