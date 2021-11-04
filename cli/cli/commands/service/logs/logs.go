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
	docker_manager_types "github.com/kurtosis-tech/container-engine-lib/lib/docker_manager/types"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	kurtosisLogLevelArg = "kurtosis-log-level"
	enclaveIdArg        = "enclave-id"
	guidArg             = "guid"

	shouldShowStoppedUserServiceContainers = true
	shouldFollowContainerLogs              = false
)

var defaultKurtosisLogLevel = logrus.InfoLevel.String()
var positionalArgs = []string{
	enclaveIdArg,
	guidArg,
}

var LogsCmd = &cobra.Command{
	Use:                   command_str_consts.ServiceLogsCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Show logs for a service inside of an enclave",
	RunE:                  run,
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

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIdArg]
	guid := parsedPositionalArgs[guidArg]

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		logrus.StandardLogger(),
		dockerClient,
	)

	labels := getUserServiceContainerLabelsWithEnclaveId(enclaveId)

	containers, err := dockerManager.GetContainersByLabels(ctx, labels, shouldShowStoppedUserServiceContainers)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting containers by labels: '%+v'", labels)
	}

	if containers == nil || len(containers) == 0 {
		logrus.Errorf("There is not any service container with enclave ID '%v'", enclaveId)
		return nil
	}

	var containersWithSearchedGUID = []*docker_manager_types.Container{}
	for _, container := range containers {
		labelsMap := container.GetLabels()
		containerGUID, found := labelsMap[enclave_object_labels.GUIDLabel]
		if found && containerGUID == guid {
			containersWithSearchedGUID = append(containersWithSearchedGUID, container)
		}
	}

	if len(containersWithSearchedGUID) == 0 {
		logrus.Errorf("There is not any service container with GUID '%v'", guid)
		return nil
	}

	if len(containersWithSearchedGUID) > 1 {
		return stacktrace.NewError("Only one container with enclave-id '%v' and GUID '%v' should exist but there are '%v' containers with these properties", enclaveId, guid, len(containers))
	}

	serviceContainer := containersWithSearchedGUID[0]

	readCloserLogs, err := dockerManager.GetContainerLogs(ctx, serviceContainer.GetId(), shouldFollowContainerLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service logs for container with ID '%v'", serviceContainer.GetId())
	}
	defer readCloserLogs.Close()

	_, err = stdcopy.StdCopy(logrus.StandardLogger().Out, logrus.StandardLogger().Out, readCloserLogs)
	if err == nil {
		return stacktrace.Propagate(err, "An error occurred copying the container logs to STDOUT")
	}

	return nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func getUserServiceContainerLabelsWithEnclaveId(enclaveId string) map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	labels[enclave_object_labels.EnclaveIDContainerLabel] = enclaveId
	return labels
}
