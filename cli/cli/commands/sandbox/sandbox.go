/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package sandbox

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_client"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/repl_runner"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	shouldPublishPorts = true

	isPartitioningEnabled = true

	apiContainerImageArg = "kurtosis-api-image"
	javascriptReplImageArg = "javascript-repl-image"
	kurtosisLogLevelArg = "kurtosis-log-level"

)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var SandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Creates a new Kurtosis enclave and attaches a REPL for manipulating it",
	RunE:  run,
}

var kurtosisLogLevelStr string
var apiContainerImage string
var jsReplImage string


func init() {
	SandboxCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)

	SandboxCmd.Flags().StringVarP(
		&apiContainerImage,
		apiContainerImageArg,
		"a",
		defaults.DefaultApiContainerImage,
		"The image of the Kurtosis API container to use inside the enclave",
	)

	SandboxCmd.Flags().StringVarP(
		&jsReplImage,
		javascriptReplImageArg,
		"r",
		defaults.DefaultJavascriptReplImage,
		"The image of the Javascript REPL to connect to the enclave with",
	)
}

func run(cmd *cobra.Command, args []string) error {

	ctx := context.Background()

	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	best_effort_image_puller.PullImageBestEffort(context.Background(), dockerManager, jsReplImage)

	enclaveId := execution_ids.GetExecutionID()

	engineClient, closeClientFunc, err := engine_client.NewEngineClientFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new engine client")
	}
	defer closeClientFunc()

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId: enclaveId,
		ApiContainerImage: apiContainerImage,
		ApiContainerLogLevel: kurtosisLogLevelStr,
		IsPartitioningEnabled: isPartitioningEnabled,
		ShouldPublishAllPorts: shouldPublishPorts,
	}

	response, err := engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}
	enclaveInfo := response.GetEnclaveInfo()

	defer func() {
		// Ensure we don't leak enclaves
		logrus.Info("Removing enclave...")
		destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{
			EnclaveId: enclaveId,
		}
		if err, _ := engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
			logrus.Errorf("An error occurred destroying enclave '%v' that the interactive environment was connected to:", enclaveId)
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			logrus.Errorf("ACTION REQUIRED: You'll need to clean this up manually!!!!")
		} else {
			logrus.Info("Enclave removed")
		}
	}()

	logrus.Debug("Running REPL...")
	if err := repl_runner.RunREPL(
		enclaveInfo.GetEnclaveId(),
		enclaveInfo.GetNetworkId(),
		enclaveInfo.GetApiContainerInfo().GetIpInsideEnclave(),
		enclaveInfo.GetApiContainerInfo().GetIpOnHostMachine(),
		enclaveInfo.GetApiContainerInfo().GetPortOnHostMachine(),
		jsReplImage,
		dockerManager); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")

	return nil
}
