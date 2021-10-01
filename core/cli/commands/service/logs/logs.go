/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_object_labels"
	"github.com/kurtosis-tech/kurtosis/commons/logrus_log_levels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg = "enclave-id"
	guidArg = "guid"
)
var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
	guidArg,
}

var LogsCmd = &cobra.Command{
	Use:   "logs " + strings.Join(positionalArgs, " "),
	Short: "Show service logs by enclave-id and guid",
	RunE:  run,
}

var kurtosisLogLevelStr string

func init() {
	LogsCmd.Flags().StringVarP(
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
	guid, found := parsedPositionalArgs[guidArg]
	if !found {
		return stacktrace.NewError("No '%v' positional args was found in '%+v' - this is very strange!", guidArg, parsedPositionalArgs)
	}

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	labels := getContainerLabelsWithEnclaveIdAndGUID(enclaveId, guid)

	containers, err := dockerManager.GetContainersByLabels(ctx, labels, true)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels: '%+v'", labels)
	}

	if containers != nil && len(containers) > 0 {
		if len(containers) > 1 {
			return stacktrace.NewError("Should exist only one container with enclave-id '%v' and guid '%v' but there are '%v' containers with these properties", enclaveId, guid, len(containers))
		}

		serviceContainer := containers[0]

		functionExitedSuccessfully := false
		readCloserLogs, err := dockerManager.GetContainerLogs(ctx, serviceContainer.GetId(), false)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service logs for container with ID '%v'", serviceContainer.GetId())
		}

		defer func() {
			if !functionExitedSuccessfully {
				readCloserLogs.Close()
			}
		}()

		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, readCloserLogs)
		if err == nil {
			return stacktrace.Propagate(err, "An error occurred executing StdCopy")
		}

		functionExitedSuccessfully = true
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

func getContainerLabelsWithEnclaveIdAndGUID(enclaveId string, guid string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	labels[enclave_object_labels.GUIDLabel] = guid
	return labels
}