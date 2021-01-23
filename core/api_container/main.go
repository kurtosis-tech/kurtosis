/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_env_var_values/api_container_modes"
	"github.com/kurtosis-tech/kurtosis/api_container/server"
	"github.com/kurtosis-tech/kurtosis/api_container/server/server_core_creator"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

const (
	listenProtocol = "tcp"
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
		fmt.Sprintf(
			"Log level to use for the API container (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
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
		logrus.Errorf("An error occurred parsing the log level string '%v':", *logLevelArg)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(int(api_container_exit_codes.StartupErrorExitCode))
	}
	logrus.SetLevel(logLevel)

	mode := api_container_modes.ApiContainerMode(*modeArg)
	paramsJson := *paramsJsonArg
	serverCore, err := server_core_creator.Create(mode, paramsJson)
	if err != nil {
		logrus.Errorf("An error occurred creating the service core for mode '%v' with params JSON '%v':", mode, paramsJson)
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(int(api_container_exit_codes.StartupErrorExitCode))
	}

	server := server.NewApiContainerServer(serverCore)

	logrus.Info("Running server...")
	exitCode := server.Run()
	logrus.Infof("Server exited with exit code '%v'", exitCode)
	os.Exit(int(exitCode))
}

