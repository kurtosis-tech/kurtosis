package new

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	apiContainerVersionArg   = "api-container-version"
	apiContainerLogLevelArg  = "api-container-log-level"
	isPartitioningEnabledArg = "with-partitioning"

	defaultIsPartitioningEnabled = false
	shouldPublishPorts = true
)

var apiContainerVersion string
var isPartitioningEnabled bool
var kurtosisLogLevelStr string

var NewCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveNewCmdStr,
	Short:                 "Creates a new, empty Kurtosis enclave",
	RunE:                  run,
}

func init() {
	NewCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		apiContainerLogLevelArg,
		"l",
		defaults.DefaultApiContainerLogLevel.String(),
		fmt.Sprintf(
			"The log level that the API container should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	NewCmd.Flags().StringVarP(
		&apiContainerVersion,
		apiContainerVersionArg,
		"a",
		defaults.DefaultAPIContainerVersion,
		"The version of the Kurtosis API container that should be started inside the enclave (blank tells the engine to use the default version)",
	)
	NewCmd.Flags().BoolVarP(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		"p",
		defaultIsPartitioningEnabled,
		"Enable network partitioning functionality (repartitioning won't work if this is set to false)",
	)
}

func run(cmd *cobra.Command, args []string) error {

	ctx := context.Background()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)
	engineManager := engine_manager.NewEngineManager(dockerManager)
	objAttrsProvider := schema.GetObjectAttributesProvider()
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, objAttrsProvider, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	logrus.Info("Creating new enclave...")
	enclaveId := execution_ids.GetExecutionID()
	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId:              enclaveId,
		ApiContainerVersionTag: apiContainerVersion,
		ApiContainerLogLevel:   kurtosisLogLevelStr,
		IsPartitioningEnabled:  isPartitioningEnabled,
		ShouldPublishAllPorts:  shouldPublishPorts,
	}
	_, err = engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}
	logrus.Infof("Successfully created new enclave '%v'", enclaveId)

	return nil
}
