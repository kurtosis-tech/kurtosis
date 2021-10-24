package new

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/engine_client"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	apiContainerImageArg     = "api-container-image"
	isPartitioningEnabledArg = "partition-enabled"
	kurtosisLogLevelArg      = "kurtosis-log-level"

	defaultIsPartitioningEnabled = false
	shouldPublishPorts = true
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()

var apiContainerImage string
var isPartitioningEnabled bool
var kurtosisLogLevelStr string

var NewCmd = &cobra.Command{
	Use:                   "new" ,
	Short:                 "Creates a new empty Kurtosis enclave",
	RunE:                  run,
}

func init() {
	NewCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	NewCmd.Flags().StringVarP(
		&apiContainerImage,
		apiContainerImageArg,
		"a",
		defaults.DefaultApiContainerImage,
		"The Kurtosis API Container Docker image that will be used to start the enclave's API container server",
	)
	NewCmd.Flags().BoolVarP(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		"p",
		defaultIsPartitioningEnabled,
		"Enable network partition functionality (repartitioning won't work if this is set to false)",
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

	enclaveId := execution_ids.GetExecutionID()

	engineClient, closeClientFunc, err := engine_client.NewEngineClientFromLocalEngine(ctx, dockerManager)
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

	_, err = engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}

	return nil
}
