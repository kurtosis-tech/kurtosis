/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/execution_impl"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/execution"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {
	customParamsJsonArg := flag.String(
		"custom-params-json",
		"{}",
		"JSON string containing custom data that the testsuite will deserialize to modify runtime behaviour",
	)

	kurtosisApiSocketArg := flag.String(
		"kurtosis-api-socket",
		"",
		"Socket in the form of address:port of the Kurtosis API container",
	)

	logLevelArg := flag.String(
		"log-level",
		"",
		"String indicating the loglevel that the test suite should output with",
	)

	flag.Parse()

	// >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<
	configurator := execution_impl.NewExampleTestsuiteConfigurator()
	// >>>>>>>>>>>>>>>>>>> REPLACE WITH YOUR OWN CONFIGURATOR <<<<<<<<<<<<<<<<<<<<<<<<

	suiteExecutor := execution.NewTestSuiteExecutor(*kurtosisApiSocketArg, *logLevelArg, *customParamsJsonArg, configurator)
	if err := suiteExecutor.Run(); err != nil {
		logrus.Errorf("An error occurred running the test suite executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
