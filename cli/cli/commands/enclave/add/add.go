package add

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	enclave_consts "github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/enclave"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	apiContainerVersionArg   = "api-container-version"
	apiContainerLogLevelArg  = "api-container-log-level"
	isPartitioningEnabledArg = "with-partitioning"
	// TODO(deprecation) remove enclave ids in favor of names
	enclaveIdArg   = "id"
	enclaveNameArg = "name"

	defaultIsPartitioningEnabled = false

	// Signifies that an enclave ID should be auto-generated
	autogenerateEnclaveIdKeyword = ""

	// Signifies that an enclave name should be auto-generated
	autogenerateEnclaveNameKeyword = ""
)

var apiContainerVersion string
var isPartitioningEnabled bool
var kurtosisLogLevelStr string
var enclaveIdStr string
var enclaveName string

// EnclaveAddCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
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
		fmt.Sprintf(
			"The enclave ID to give the new enclave, which must match regex '%v' "+
				"(emptystring will autogenerate an enclave ID). Note this will be deprecated in favor of '%v'",
			enclave_consts.AllowedEnclaveNameCharsRegexStr,
			enclaveNameArg,
		),
	)
	EnclaveAddCmd.Flags().StringVarP(
		&enclaveName,
		enclaveNameArg,
		"n",
		autogenerateEnclaveNameKeyword,
		fmt.Sprintf(
			"The enclave name to give the new enclave, which must match regex '%v' "+
				"(emptystring will autogenerate an enclave name)",
			enclave_consts.AllowedEnclaveNameCharsRegexStr,
		),
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

	// if the enclave id is provider but name isn't we go with the supplied id
	// TODO deprecate ids
	if enclaveIdStr != autogenerateEnclaveIdKeyword && enclaveName == autogenerateEnclaveNameKeyword {
		enclaveName = enclaveIdStr
	}

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:            enclaveName,
		ApiContainerVersionTag: apiContainerVersion,
		ApiContainerLogLevel:   kurtosisLogLevelStr,
		IsPartitioningEnabled:  isPartitioningEnabled,
	}
	createdEnclaveResponse, err := engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveIdStr)
	}

	enclaveInfo := createdEnclaveResponse.GetEnclaveInfo()
	enclaveName = enclaveInfo.Name

	defer output_printers.PrintEnclaveName(enclaveName)

	return nil
}
