/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_vars"
	"github.com/kurtosis-tech/kurtosis/api_container/execution_codepath"
	"github.com/kurtosis-tech/kurtosis/api_container/exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/print_suite_metadata_mode"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (

	successExitCode = 0
	failureExitCode = 1
)

func main() {
	// NOTE: we'll want to chnage the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	logLevelArg := flag.String(
		"log-level",
		"info",
		fmt.Sprintf("Log level to use for the API container (%v)", logrus_log_levels.AcceptableLogLevels),
	)

	acceptedModesSlice := []string{}
	for mode := range api_container_env_vars.AllModes {
		acceptedModesSlice = append(acceptedModesSlice, mode)
	}
	modeArg := flag.String(
		"mode",
		"",
		fmt.Sprintf(
			"Mode that the API container should run in (allowed: %v)",
			strings.Join(acceptedModesSlice, ", "),
		),
	)

	paramsJsonArg := flag.String(
		"params-json",
		"",
		"JSON string containing the params to the API container",
	)

	flag.Parse()

	logLevel, err := logrus.ParseLevel(*logLevelArg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occurred parsing the log level string: %v\n", err)
		os.Exit(exit_codes.StartupErrorExitCode)
	}
	logrus.SetLevel(logLevel)

	paramsJsonBytes := []byte(*paramsJsonArg)
	mode := *modeArg

	var codepath execution_codepath.ExecutionCodepath
	switch mode {
	case api_container_env_vars.SuiteMetadataPrintingMode:
		var args print_suite_metadata_mode.PrintSuiteMetadataArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			logrus.Errorf("An error occurred deserializing the suite metadata printer args:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			// TODO switch to proper exit code
			os.Exit(failureExitCode)
		}
		codepath = print_suite_metadata_mode.NewPrintSuiteMetadataCodepath(args)
	case api_container_env_vars.TestExecutionMode:
		var args test_execution_mode.TestExecutionArgs
		if err := json.Unmarshal(paramsJsonBytes, &args); err != nil {
			logrus.Errorf("An error occurred deserializing the test execution args:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			// TODO switch to proper exit code
			os.Exit(failureExitCode)
		}
		codepath = test_execution_mode.NewTestExecutionCodepath(args)
	}

	exitCode, err := codepath.Execute()
	if err != nil {
		logrus.Errorf("An error occurred running the codepath for mode '%v':", mode)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	} else {
		logrus.Infof("Successfully ran codepath for mode '%v'", mode)
	}
	os.Exit(exitCode)
}

