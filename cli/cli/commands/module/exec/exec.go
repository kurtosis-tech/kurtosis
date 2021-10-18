/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package exec

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/enclave_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/binding_constructors"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	loadParamsStrArg = "load-params"
	executeParamsStrArg = "execute-params"
	apiContainerImageArg = "api-container-image"

	moduleImageArg = "module-image"

	defaultLoadParams              = "{}"
	defaultExecuteParams           = "{}"

	shouldEnablePartitioning = true
	shouldPublishAllPorts = true

	moduleId = "my-module"
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var positionalArgs = []string{
	moduleImageArg,
}

var kurtosisLogLevelStr string
var loadParamsStr string
var executeParamsStr string
var apiContainerImage string

var ExecCmd = &cobra.Command{
	Use:   "exec [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Creates a new enclave and loads & executes the given executable module inside it",
	RunE:  run,
}

func init() {
	ExecCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	ExecCmd.Flags().StringVar(
		&loadParamsStr,
		loadParamsStrArg,
		defaultLoadParams,
		"The serialized params that should be passed to the module when loading it",
	)
	ExecCmd.Flags().StringVar(
		&executeParamsStr,
		executeParamsStrArg,
		defaultExecuteParams,
		"The serialized params that should be passed to the module when executing it",
	)
	ExecCmd.Flags().StringVar(
		&apiContainerImage,
		apiContainerImageArg,
		defaults.DefaultApiContainerImage,
		"The image of the API container that should be started inside the enclave where the module will execute",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	moduleImage := parsedPositionalArgs[moduleImageArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)

	best_effort_image_puller.PullImageBestEffort(ctx, dockerManager, apiContainerImage)
	best_effort_image_puller.PullImageBestEffort(ctx, dockerManager, moduleImage)

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient)

	logrus.Info("Creating enclave for the module to execute inside...")
	executionId := execution_ids.GetExecutionID()
	enclaveCtx, err := enclaveManager.CreateEnclave(ctx, logrus.StandardLogger(), apiContainerImage, kurtosisLogLevel, executionId, shouldEnablePartitioning, shouldPublishAllPorts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave to execute the module in")
	}
	shouldDestroyEnclave := true
	defer func() {
		if shouldDestroyEnclave {
			if err := enclaveManager.DestroyEnclave(context.Background(), logrus.StandardLogger(), enclaveCtx); err != nil {
				logrus.Errorf(
					"The module didn't execute correctly so we tried to destroy the created enclave, but destroying the enclave threw an error:\n%v",
					err,
				)
				logrus.Errorf("ACTION NEEDED: You'll need to destroy enclave '%v' manually!!", enclaveCtx.GetEnclaveID())
			}
		}
	}()
	logrus.Infof("Enclave '%v' created successfully", enclaveCtx.GetEnclaveID())

	apiContainerHostPortBinding := enclaveCtx.GetAPIContainerHostPortBinding()
	apiContainerHostUrl := fmt.Sprintf(
		"%v:%v",
		apiContainerHostPortBinding.HostIP,
		apiContainerHostPortBinding.HostPort,
	)
	conn, err := grpc.Dial(apiContainerHostUrl, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container at '%v' in enclave '%v'",
			apiContainerHostUrl,
			enclaveCtx.GetEnclaveID(),
		)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	logrus.Infof(
		"Loading module '%v' with parameters '%v' inside the enclave...",
		moduleImage,
		loadParamsStr,
	)
	loadModuleArgs := binding_constructors.NewLoadModuleArgs(moduleId, moduleImage, loadParamsStr)
	if _, err := apiContainerClient.LoadModule(ctx, loadModuleArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the module with image '%v'", moduleImage)
	}
	logrus.Info("Module loaded successfully")

	logrus.Infof("Executing the module with parameters '%v'...", executeParamsStr)
	executeModuleArgs := binding_constructors.NewExecuteModuleArgs(moduleId, executeParamsStr)
	executeModuleResult, err := apiContainerClient.ExecuteModule(ctx, executeModuleArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the module with params '%v'", executeParamsStr)
	}
	logrus.Infof(
		"Module executed successfully and returned the following result:\n%v",
		executeModuleResult.SerializedResult,
	)

	shouldDestroyEnclave = false
	return nil
}
