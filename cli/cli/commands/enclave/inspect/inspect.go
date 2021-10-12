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
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg        = "enclave-id"

	tabWriterMinwidth = 0
	tabWriterTabwidth = 0
	tabWriterPadding  = 3
	tabWriterPadchar  = ' '
	tabWriterFlags    = 0

	headerWidthChars = 100
	headerPadChar = "="
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
}

var enclaveObjectPrintingFuncs = map[string]func(context.Context, *docker_manager.DockerManager, string) error {
	"Interactive REPLs": printInteractiveRepls,
	"User Services": printUserServices,
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

	for header, printingFunc := range enclaveObjectPrintingFuncs {
		numRunesInHeader := utf8.RuneCountInString(header) + 2	// 2 because there will be a space before and after the header
		numPadChars := (headerWidthChars - numRunesInHeader) / 2
		padStr := strings.Repeat(headerPadChar, numPadChars)
		fmt.Println(fmt.Sprintf("%v %v %v", padStr, header, padStr))

		printingFunc()
	}

	if err := printUserServices(ctx, dockerManager, enclaveId); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the user services for enclave '%v'", enclaveId)
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getTabWriterForPrinting() *tabwriter.Writer {
	return tabwriter.NewWriter(
		os.Stdout,
		tabWriterMinwidth,
		tabWriterTabwidth,
		tabWriterPadding,
		tabWriterPadchar,
		tabWriterFlags,
	)
}

func writeElemsToTabWriter(writer *tabwriter.Writer, elems... string) {
	fmt.Fprintln(
		writer,
		strings.Join(elems, "\t"),
	)
}

func getContainersSortedByGUID(containers []*types.Container) ([]*types.Container, error) {
	containersSet := map[string]*types.Container{}
	for _, container := range containers {
		if container != nil {
			containerGUID, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
			if !found {
				return nil, stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
			}
			containersSet[containerGUID] = container
		}
	}

	containersResult := make([]*types.Container, 0, len(containersSet))
	for _, container := range containersSet {
		containersResult = append(containersResult, container)
	}

	sort.Slice(containersResult, func(i, j int) bool {
		return containersResult[i].GetLabels()[enclave_object_labels.GUIDLabel] < containersResult[j].GetLabels()[enclave_object_labels.GUIDLabel]
	})

	return containersResult, nil
}
