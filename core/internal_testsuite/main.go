/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/execution"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/execution_impl"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	successExitCode = 0
	failureExitCode = 1
)

func main() {
	configurator := execution_impl.NewInternalTestsuiteConfigurator()
	suiteExecutor := execution.NewTestSuiteExecutor(configurator)
	if err := suiteExecutor.Run(); err != nil {
		logrus.Errorf("An error occurred running the test suite executor:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}
