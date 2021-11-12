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
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/repl_runner"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	shouldPublishPorts = true

	isPartitioningEnabled = true

	apiContainerVersionArg = "api-container-version"
	javascriptReplImageArg = "javascript-repl-image"
	kurtosisLogLevelArg    = "kurtosis-log-level"

)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var SandboxCmd = &cobra.Command{
	Use:   command_str_consts.SandboxCmdStr,
	Short: "Creates a new Kurtosis enclave and attaches a REPL for manipulating it",
	RunE:  run,
}

var kurtosisLogLevelStr string
var apiContainerVersion string
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

	SandboxCmd.Flags().StringVar(
		&apiContainerVersion,
		apiContainerVersionArg,
		defaults.DefaultAPIContainerVersion,
		"The version of the Kurtosis API container to use inside the enclave (blank will use the engine server default version)",
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

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	logrus.Info("Creating new enclave...")
	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId:              enclaveId,
		ApiContainerVersionTag: apiContainerVersion,
		ApiContainerLogLevel:   kurtosisLogLevelStr,
		IsPartitioningEnabled:  isPartitioningEnabled,
		ShouldPublishAllPorts:  shouldPublishPorts,
	}
	response, err := engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}
	enclaveInfo := response.GetEnclaveInfo()
	shouldStopEnclave := true
	defer func() {
		if shouldStopEnclave {
			destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
				EnclaveId: enclaveId,
			}
			if _, err := engineClient.StopEnclave(ctx, destroyEnclaveArgs); err != nil {
				logrus.Errorf("An error occurred so we tried to stop enclave '%v' that we started, but an error was thrown:", enclaveId)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to stop enclave '%v' manually!!!!", enclaveId)
			} else {
				logrus.Info("Enclave stopped")
			}
		}
	}()
	apicHostMachineIp, apicHostMachinePort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "Cannot create sandbox; an error occurred verifying enclave liveness")
	}
	// The enclave was set up successfully, so from this point on the user should be using enclave lifecycle management
	//  tools to get rid of it
	shouldStopEnclave = false
	logrus.Infof("New enclave '%v' created successfully", enclaveId)

	logrus.Debug("Running REPL...")
	enclaveObjAttrsProvider := objAttrsProvider.ForEnclave(enclaveId)
	if err := repl_runner.RunREPL(
		enclaveInfo.GetEnclaveId(),
		enclaveInfo.GetNetworkId(),
		enclaveInfo.GetApiContainerInfo().GetIpInsideEnclave(),
		enclaveInfo.GetApiContainerInfo().GetPortInsideEnclave(),
		apicHostMachineIp,
		apicHostMachinePort,
		jsReplImage,
		dockerManager,
		enclaveObjAttrsProvider,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the REPL container")
	}
	logrus.Debug("REPL exited")
	logrus.Infof(
		"NOTE: Enclave '%v' will be left running until stopped by running '%v %v %v %v'",
		enclaveId,
		command_str_consts.KurtosisCmdStr,
		command_str_consts.EnclaveCmdStr,
		command_str_consts.EnclaveStopCmdStr,
		enclaveId,
	)

	return nil
}
