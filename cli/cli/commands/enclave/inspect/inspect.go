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
	"github.com/kurtosis-tech/kurtosis-cli/cli/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis-cli/cli/logrus_log_levels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
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

	labels := getLabelsForListEnclaveUserServices(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, labels, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels: '%+v'", labels)
	}

	if containers != nil {
		tabWriter := tabwriter.NewWriter(os.Stdout, tabWriterMinwidth, tabWriterTabwidth, tabWriterPadding, tabWriterPadchar, tabWriterFlags)
		fmt.Fprintln(tabWriter, guidHeader + "\t" + nameHeader + "\t" + hostPortBindingsHeader)
		sortedContainers, err := getContainersSortedByGUID(containers)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting containers sorted by GUID")
		}
		for _, container := range sortedContainers {
			containerGUIDLabel, found := container.GetLabels()[enclave_object_labels.GUIDLabel]
			if !found {
				return stacktrace.NewError("No '%v' container label was found in container ID '%v' with labels '%+v'", enclave_object_labels.GUIDLabel, container.GetId(), container.GetLabels())
			}
			hostPortBindingsStrings := getContainerHostPortBindingStrings(container)

			var firstHostPortBinding string
			if hostPortBindingsStrings != nil  {
				firstHostPortBinding = hostPortBindingsStrings[0]
				hostPortBindingsStrings = hostPortBindingsStrings[1:]
			}
			line := containerGUIDLabel + "\t" + container.GetName() + "\t" + firstHostPortBinding
			fmt.Fprintln(tabWriter, line)

			for _, hostPortBindingsString := range hostPortBindingsStrings {
				line = "\t\t" + hostPortBindingsString
				fmt.Fprintln(tabWriter, line)
			}
		}
		tabWriter.Flush()
	}

	return nil
}

func getContainerHostPortBindingStrings(container *types.Container) []string {
	var allHosPortBindings []string
	hostPortBindings := container.GetHostPortBindings()
	for hostPortBindingKey, hostPortBinding := range hostPortBindings {
		hostPortBindingString := fmt.Sprintf("%v -> %v:%v", hostPortBindingKey, hostPortBinding.HostIP, hostPortBinding.HostPort)
		allHosPortBindings = append(allHosPortBindings, hostPortBindingString)
	}
	return allHosPortBindings
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getLabelsForListEnclaveUserServices(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
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
