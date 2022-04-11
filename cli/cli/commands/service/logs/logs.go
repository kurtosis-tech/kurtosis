/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	shouldFollowLogsFlag = "follow"
	enclaveIdArg        = "enclave-id"
	guidArg             = "guid"

	shouldShowStoppedUserServiceContainers = true

	defaultShouldFollowLogs = false
)

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

var shouldFollowLogs bool

func init() {
	LogsCmd.Flags().BoolVarP(
		&shouldFollowLogs,
		shouldFollowLogsFlag,
		"f",
		defaultShouldFollowLogs,
		"Continues to follow the logs until stopped",
	)
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveIdStr := parsedPositionalArgs[enclaveIdArg]
	enclaveId := enclave.EnclaveID(enclaveIdStr)
	guidStr := parsedPositionalArgs[guidArg]
	guid := service.ServiceGUID(guidStr)

	// TODO Once KurtosisBackend can print logs
	/*dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker client")
	}
	dockerManager := docker_manager.NewDockerManager(
		dockerClient,
	)

	labels := labels_helper.GetUserServiceContainerLabelsWithEnclaveID(enclaveId)

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
		containerGUID, found := labelsMap[schema.GUIDLabel]
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
	serviceContainerId := serviceContainer.GetId()

	// TODO vvvvvvvvvvvv Abstract everything below this point into KurtosisBackend when it's ready!!!! vvvvvvvvvvvvv
	inspectResult, err := dockerClient.ContainerInspect(ctx, serviceContainerId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred inspecting service container with ID '%v' to determine if it's running a TTY or not", serviceContainerId)
	}

	readCloserLogs, err := dockerManager.GetContainerLogs(ctx, serviceContainerId, shouldFollowLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service logs for container with ID '%v'", serviceContainer.GetId())
	}
	defer readCloserLogs.Close()

	// TODO This is already copied from 'enclave dump'; this logic should be centralized
	// If we don't have this, reading the logs from REPL container breaks
	stdout := logrus.StandardLogger().Out
	if inspectResult.Config.Tty {
		if _, err := io.Copy(stdout, readCloserLogs); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the TTY container logs stream to STDOUT for container with ID '%v'",
				serviceContainerId,
			)
		}
	} else {
		if _, err := stdcopy.StdCopy(stdout, stdout, readCloserLogs); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred copying the non-TTY container logs stream to STDOUT for container with name '%v' and ID '%v'",
				serviceContainer.GetName(),
				serviceContainerId,
			)
		}
	}
	// TODO ^^^^^^^^^^^^ Abstract everything below this point into KurtosisBackend when it's ready!!!! ^^^^^^^^^^^^^^^^
	 */

	kurtosisBackend, err := lib.GetLocalDockerKurtosisBackend()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
	}

	userServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			guid: true,
		},
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := kurtosisBackend.GetUserServiceLogs(ctx, userServiceFilters, shouldFollowLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}

	if len(erroredUserServiceGuids) > 0 {
		err, found := erroredUserServiceGuids[guid]
		if !found {
			return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; it should never happens, it is a bug in Kurtosis", guid, erroredUserServiceGuids)
		}
		return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with GUID '%v'", guid)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[guid]
	if !found {
		return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; it should never happens, it is a bug in Kurtosis", guid, userServiceReadCloserLog)
	}

	stdout := logrus.StandardLogger().Out
	if _, err := stdcopy.StdCopy(stdout, stdout, userServiceReadCloserLog); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the user service logs stream to STDOUT for user service with GUID '%v'",
			guid,
		)
	}

	return nil
}
