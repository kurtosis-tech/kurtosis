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
	"github.com/kurtosis-tech/kurtosis/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg = "enclave-id"
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
}

var InspectCmd = &cobra.Command{
	Use:   "inspect " + strings.Join(positionalArgs, " "),
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

	parsedPositionalArgs, err := parsePositionalArgs(args)
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
		fmt.Println("GUID\tName")
		sortedContainers := getContainersSortedByGUID(containers)
		for _, container := range sortedContainers {
			fmt.Printf("%v\t%v\n", container.GetLabels()[enclave_object_labels.GUIDLabel], container.GetName())
		}
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
// Parses the args into a map of positional_arg_name -> value
func parsePositionalArgs(args []string) (map[string]string, error) {
	if len(args) != len(positionalArgs) {
		return nil, stacktrace.NewError("Expected %v positional arguments but got %v", len(positionalArgs), len(args))
	}

	result := map[string]string{}
	for idx, argValue := range args {
		arg := positionalArgs[idx]
		result[arg] = argValue
	}
	return result, nil
}

func getLabelsForListEnclaveUserServices(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}

func getContainersSortedByGUID(containers []*docker_manager.Container) []*docker_manager.Container{
	containersSet := map[string]*docker_manager.Container{}
	for _, container := range containers {
		if container != nil {
			containerGUID := container.GetLabels()[enclave_object_labels.GUIDLabel]
			containersSet[containerGUID] = container
		}
	}

	containersGUID := []string{}
	for _, container := range containersSet {
		guid := container.GetLabels()[enclave_object_labels.GUIDLabel]
		containersGUID = append(containersGUID, guid)
	}

	sort.Strings(containersGUID)

	sortedContainers := make([]*docker_manager.Container, 0, len(containersGUID))
	for _, containerGUID := range containersGUID {
		container := containersSet[containerGUID]
		sortedContainers = append(sortedContainers, container)
	}

	return sortedContainers
}
