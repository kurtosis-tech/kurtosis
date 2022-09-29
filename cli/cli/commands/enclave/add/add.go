package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	apiContainerVersionArg   = "api-container-version"
	apiContainerLogLevelArg  = "api-container-log-level"
	isPartitioningEnabledArg = "with-partitioning"
	enclaveIdArg             = "id"

	defaultIsPartitioningEnabled = false

	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdKeyword = ""
)

var apiContainerVersion string
var isPartitioningEnabled bool
var kurtosisLogLevelStr string
var enclaveIdStr string

var EnclaveAddCmd = &cobra.Command{
	Use:     command_str_consts.EnclaveAddCmdStr,
	Short:   "Creates an enclave",
	Long:    "Creates a new, empty Kurtosis enclave",
	RunE:    run,
	Aliases: []string{"new"}, // TODO remove this after 2022-08-16 when everyone should be using "add"
}

func init() {
	EnclaveAddCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		apiContainerLogLevelArg,
		"l",
		defaults.DefaultApiContainerLogLevel.String(),
		fmt.Sprintf(
			"The log level that the API container should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)
	EnclaveAddCmd.Flags().StringVarP(
		&apiContainerVersion,
		apiContainerVersionArg,
		"a",
		defaults.DefaultAPIContainerVersion,
		"The version of the Kurtosis API container that should be started inside the enclave (blank tells the engine to use the default version)",
	)
	EnclaveAddCmd.Flags().BoolVarP(
		&isPartitioningEnabled,
		isPartitioningEnabledArg,
		"p",
		defaultIsPartitioningEnabled,
		"Enable network partitioning functionality (repartitioning won't work if this is set to false)",
	)
	EnclaveAddCmd.Flags().StringVarP(
		&enclaveIdStr,
		enclaveIdArg,
		"i",
		autogenerateEnclaveIdKeyword,
		"The enclave ID to give the new enclave (emptystring will autogenerate an enclave ID)",
	)
}

func run(cmd *cobra.Command, args []string) error {

	ctx := context.Background()

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager.")
	}
	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer func() {
		if err = closeClientFunc(); err != nil {
			logrus.Errorf("Error closing the engine client")
		}
	}()

	logrus.Info("Creating new enclave...")
	var enclaveId string
	if enclaveIdStr == autogenerateEnclaveIdKeyword {
		enclaveId = execution_ids.GetExecutionID()
	} else {
		enclaveId = enclaveIdStr
	}
	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId:              enclaveId,
		ApiContainerVersionTag: apiContainerVersion,
		ApiContainerLogLevel:   kurtosisLogLevelStr,
		IsPartitioningEnabled:  isPartitioningEnabled,
	}
	_, err = engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}
	logrus.Infof("Successfully created new enclave '%v'", enclaveId)

	return nil
}
