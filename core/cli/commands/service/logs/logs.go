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
	"github.com/kurtosis-tech/kurtosis/cli/positional_arg_parser"
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
	Use:   "logs [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short: "Show logs for a service inside of an enclave",
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

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgs(positionalArgs, args)
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

		readCloserLogs, err := dockerManager.GetContainerLogs(ctx, serviceContainer.GetId(), false)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service logs for container with ID '%v'", serviceContainer.GetId())
		}
		defer readCloserLogs.Close()

		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, readCloserLogs)
		if err == nil {
			return stacktrace.Propagate(err, "An error occurred executing StdCopy")
		}

	}
	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getContainerLabelsWithEnclaveIdAndGUID(enclaveId string, guid string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	labels[enclave_object_labels.GUIDLabel] = guid
	return labels
}
