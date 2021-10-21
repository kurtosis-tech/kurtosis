package new

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/execution_ids"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/lib/kurtosis_context"
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

	enclaveId := execution_ids.GetExecutionID()

	kurtosisContext, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis Context")
	}

	_, err = kurtosisContext.CreateEnclave(
		ctx,
		enclaveId,
		apiContainerImage,
		kurtosisLogLevelStr,
		isPartitioningEnabled,
		shouldPublishPorts)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an enclave, make sure that you already started Kurtosis Engine Sever with `kurtosis engine start` command")
	}

	return nil
}
