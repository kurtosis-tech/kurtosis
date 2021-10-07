/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package exec

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
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

	lambdaImageArg = "lambda-image"

	defaultLoadParams = "{}"
	defaultExecuteParams = "{}"

	shouldEnablePartitioning = true
	shouldPublishAllPorts = true

	lambdaId = "my-lambda"
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var positionalArgs = []string{
	lambdaImageArg,
}

var kurtosisLogLevelStr string
var loadParamsStr string
var executeParamsStr string
var apiContainerImage string

var ExecCmd = &cobra.Command{
	Use:   "exec [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Creates a new enclave and loads & executes the given Lambda inside it",
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
		"The serialized params that should be passed to the Lambda when loading it",
	)
	ExecCmd.Flags().StringVar(
		&executeParamsStr,
		executeParamsStrArg,
		defaultExecuteParams,
		"The serialized params that should be passed to the Lambda when executing it",
	)
	ExecCmd.Flags().StringVar(
		&apiContainerImage,
		apiContainerImageArg,
		defaults.DefaultApiContainerImage,
		"The image of the API container that should be started inside the enclave where the Lambda will execute",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	kurtosisLogLevel, err := logrus.ParseLevel(kurtosisLogLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing Kurtosis loglevel string '%v' to a log level object", kurtosisLogLevelStr)
	}
	logrus.SetLevel(kurtosisLogLevel)

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgs(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	lambdaImage := parsedPositionalArgs[lambdaImageArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}

	enclaveManager := enclave_manager.NewEnclaveManager(dockerClient, apiContainerImage)

	logrus.Info("Creating enclave for the Lambda to execute inside...")
	executionId := execution_ids.GetExecutionID()
	enclaveCtx, err := enclaveManager.CreateEnclave(ctx, logrus.StandardLogger(), kurtosisLogLevel, executionId, shouldEnablePartitioning, shouldPublishAllPorts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave to execute the Lambda in")
	}
	shouldDestroyEnclave := true
	defer func() {
		if shouldDestroyEnclave {
			if err := enclaveManager.DestroyEnclave(context.Background(), logrus.StandardLogger(), enclaveCtx); err != nil {
				logrus.Errorf(
					"The Lambda didn't execute correctly so we tried to destroy the created enclave, but destroying the enclave threw an error:\n%v",
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
		"Loading Lambda '%v' with parameters '%v' inside the enclave...",
		lambdaImage,
		loadParamsStr,
	)
	loadLambdaArgs := binding_constructors.NewLoadLambdaArgs(lambdaId, lambdaImage, loadParamsStr)
	if _, err := apiContainerClient.LoadLambda(ctx, loadLambdaArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the Lambda with image '%v'", lambdaImage)
	}
	logrus.Info("Lambda loaded successfully")

	logrus.Infof("Executing the Lambda with parameters '%v'...", executeParamsStr)
	executeLambdaArgs := binding_constructors.NewExecuteLambdaArgs(lambdaId, executeParamsStr)
	executeLambdaResult, err := apiContainerClient.ExecuteLambda(ctx, executeLambdaArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing the Lambda with params '%v'", executeParamsStr)
	}
	logrus.Infof(
		"Lambda executed successfully and returned the following result:\n%v",
		executeLambdaResult.SerializedResult,
	)

	shouldDestroyEnclave = false
	return nil
}
