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
	"github.com/kurtosis-tech/kurtosis/api_container/server_core_creator"
	"github.com/kurtosis-tech/kurtosis/api_container/suite_metadata_serializing_mode"
	"github.com/kurtosis-tech/kurtosis/api_container/test_execution_mode"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/sirupsen/logrus"
	"os"
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

	modeArg := flag.String(
		"mode",
		"",
		"Mode that the API container should run in",
	)

	// NOTE: We take this in as JSON so that each mode can have their own independent args
	paramsJsonArg := flag.String(
		"params-json",
		"",
		"JSON string containing the params to the API container",
	)

	flag.Parse()

	logLevel, err := logrus.ParseLevel(*logLevelArg)
	if err != nil {
		logrus.Errorf("An error occurred parsing the log level string '%v':")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(int(exit_codes.StartupErrorExitCode))
	}
	logrus.SetLevel(logLevel)

	mode := api_container_env_vars.ApiContainerMode(*modeArg)
	paramsJson := *paramsJsonArg
	serverCore, err := server_core_creator.Create(mode, paramsJson)
	if err != nil {
		logrus.Errorf("An error occurred creating the service core for mode '%v' with params JSON '%v':", mode, paramsJson)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(int(exit_codes.StartupErrorExitCode))
	}

	server :=

	exitCode, err := codepath.Execute()
	if err != nil {
		logrus.Errorf("An error occurred running the codepath for mode '%v':", mode)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
	} else {
		logrus.Infof("Successfully ran codepath for mode '%v'", mode)
	}
	os.Exit(exitCode)
}

