package main

import (
	"flag"
	"fmt"
	"github.com/gmarchetti/kurtosis/initializer"
)

func main() {
	fmt.Println("Welcome to Kurtosis E2E Testing for Ava.")

	// Define and parse command line flags.
	geckoImageNameArg := flag.String(
		"gecko-image-name", 
		"gecko-f290f73", // by default, pick commit that was on master May 14, 2020.
		"the name of a pre-built gecko image in your docker engine.",
	)
	flag.Parse()

	testSuiteRunner := initializer.NewTestSuiteRunner(*geckoImageNameArg)
	testSuiteRunner.RunTests()
}
