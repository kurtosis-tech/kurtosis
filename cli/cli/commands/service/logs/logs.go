/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
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

	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	kurtosisBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
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
			return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", guid, erroredUserServiceGuids)
		}
		return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with GUID '%v'", guid)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[guid]
	if !found {
		return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", guid, userServiceReadCloserLog)
	}

	stdout := logrus.StandardLogger().Out
	// TODO This is a Docker library call; this should be pushed down into KurtosisBackend
	if _, err := stdcopy.StdCopy(stdout, stdout, userServiceReadCloserLog); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the user service logs stream to STDOUT for user service with GUID '%v'",
			guid,
		)
	}

	return nil
}
