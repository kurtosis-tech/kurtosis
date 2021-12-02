/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package exec

import (
	"context"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/best_effort_image_puller"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_liveness_validator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"regexp"
	"strings"
	"time"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	loadParamsStrArg = "load-params"
	executeParamsStrArg = "execute-params"
	apiContainerVersionArg = "api-container-version"
	moduleImageArg = "module-image"
	enclaveIdArg = "enclave-id"
	isPartitioningEnabledArg = "with-partitioning"

	defaultLoadParams              = "{}"
	defaultExecuteParams           = "{}"
	defaultEnclaveId               = ""
	defaultIsPartitioningEnabled   = false

	shouldPublishAllPorts = true

	moduleId = "my-module"

	// TODO Extract this validation into a centralized location for all commands that use an enclave ID
	allowedEnclaveIdCharsRegexStr = `^[A-Za-z0-9._-]+$`
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var positionalArgs = []string{
	moduleImageArg,
}

var kurtosisLogLevelStr string
var loadParamsStr string
var executeParamsStr string
var apiContainerVersion string
var userRequestedEnclaveId string
var isPartitioningEnabled bool

var ExecCmd = &cobra.Command{
	Use:   command_str_consts.ModuleExecCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
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
		&apiContainerVersion,
		apiContainerVersionArg,
		defaults.DefaultAPIContainerVersion,
		"The image of the API container that should be started inside the enclave where the module will execute (blank will use the engine's default version)",
	)
	ExecCmd.Flags().StringVar(
		&userRequestedEnclaveId,
		enclaveIdArg,
		defaultEnclaveId,
		fmt.Sprintf(
			"The ID to give the enclave that will be created to execute the module inside, which must match regex '%v' (default: use the module image and the current Unix time)",
			allowedEnclaveIdCharsRegexStr,
		),
	)
	ExecCmd.Flags().BoolVar(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		defaultIsPartitioningEnabled,
		"If set to true, the enclave that the module executes in will have partitioning enabled so network partitioning simulations can be run",
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

	best_effort_image_puller.PullImageBestEffort(ctx, dockerManager, moduleImage)

	logrus.Info("Creating enclave for the module to execute inside...")
	enclaveId := userRequestedEnclaveId
	if enclaveId == defaultEnclaveId {
		enclaveId = getEnclaveId(moduleImage)
	}
	validEnclaveId, err := regexp.Match(allowedEnclaveIdCharsRegexStr, []byte(enclaveId))
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred validating that enclave ID '%v' matches allowed enclave ID regex '%v'",
			enclaveId,
			allowedEnclaveIdCharsRegexStr,
		)
	}
	if !validEnclaveId {
		return stacktrace.NewError(
			"Enclave ID '%v' doesn't match allowed enclave ID regex '%v'",
			enclaveId,
			allowedEnclaveIdCharsRegexStr,
		)
	}

	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId:              enclaveId,
		ApiContainerVersionTag: apiContainerVersion,
		ApiContainerLogLevel:   kurtosisLogLevelStr,
		IsPartitioningEnabled:  isPartitioningEnabled,
		ShouldPublishAllPorts:  shouldPublishAllPorts,
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
			if  _, err := engineClient.StopEnclave(ctx, destroyEnclaveArgs); err != nil {
				logrus.Errorf(
					"The module didn't execute correctly so we tried to stop the created enclave, but doing so threw an error:\n%v",
					err,
				)
				logrus.Errorf("ACTION NEEDED: You'll need to stop enclave '%v' manually!!", enclaveId)
			}
		}
	}()
	logrus.Infof("Enclave '%v' created successfully", enclaveId)

	apicHostMachineIp, apicHostMachinePort, err := enclave_liveness_validator.ValidateEnclaveLiveness(enclaveInfo)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying that the enclave was running")
	}

	apiContainerHostUrl := fmt.Sprintf(
		"%v:%v",
		apicHostMachineIp,
		apicHostMachinePort,
	)
	conn, err := grpc.Dial(apiContainerHostUrl, grpc.WithInsecure())
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred connecting to the API container at '%v' in enclave '%v'",
			apiContainerHostUrl,
			enclaveId,
		)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(conn)

	logrus.Infof(
		"Loading module '%v' with load params '%v'...",
		moduleImage,
		loadParamsStr,
	)
	loadModuleArgs := binding_constructors.NewLoadModuleArgs(moduleId, moduleImage, loadParamsStr)
	if _, err := apiContainerClient.LoadModule(ctx, loadModuleArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the module with image '%v'", moduleImage)
	}
	logrus.Info("Module loaded successfully")

	logrus.Infof("Executing the module with execute params '%v'...", executeParamsStr)
	executeModuleArgs := binding_constructors.NewExecuteModuleArgs(moduleId, executeParamsStr)
	executeModuleResult, err := apiContainerClient.ExecuteModule(ctx, executeModuleArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the module with params '%v'", executeParamsStr)
	}
	logrus.Infof(
		"Module executed successfully and returned the following result:\n%v",
		executeModuleResult.SerializedResult,
	)

	shouldStopEnclave = false
	return nil
}


// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================
func getEnclaveId(moduleImage string) string {
	defaultEnclaveId := execution_ids.GetExecutionID()
	parsedModuleImage, err := reference.Parse(moduleImage)
	if err != nil {
		logrus.Warn("Couldn't parse the module image string '%v'; using enclave ID '%v'", moduleImage, defaultEnclaveId)
		return defaultEnclaveId
	}

	namedModuleImage, ok := parsedModuleImage.(reference.Named)
	if !ok {
		logrus.Warn("Module image string '%v' couldn't be cast to a named reference; using enclave ID '%v'", moduleImage, defaultEnclaveId)
		return defaultEnclaveId
	}
	pathElement := reference.Path(namedModuleImage)

	return fmt.Sprintf(
		"%v_%v",
		pathElement,
		time.Now().Unix(),
	)
}
