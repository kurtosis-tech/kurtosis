/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg        = "enclave-id"

	tabWriterMinwidth = 0
	tabWriterTabwidth = 0
	tabWriterPadding  = 3
	tabWriterPadchar  = ' '
	tabWriterFlags    = 0

	guidHeader             = "GUID"
	nameHeader             = "Name"
	hostPortBindingsHeader = "HostPortBindings"
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
}

var InspectCmd = &cobra.Command{
	Use:   "inspect " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Inspect Kurtosis enclaves",
	RunE:  run,
}

var kurtosisLogLevelStr string

func init() {
	InspectCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
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
	enclaveId, found := parsedPositionalArgs[enclaveIdArg]
	if !found {
		return stacktrace.NewError("No '%v' positional args was found in '%+v' - this is very strange!", enclaveIdArg, parsedPositionalArgs)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	if err := printUserServices(ctx, dockerManager, enclaveId); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the user services for enclave '%v'", enclaveId)
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
