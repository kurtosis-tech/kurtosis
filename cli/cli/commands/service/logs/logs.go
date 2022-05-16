/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	enclaveIdArgKey   = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy = false

	serviceGuidArgKey = "service-guid"

	shouldFollowLogsFlagKey = "follow"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"
)

var defaultShouldFollowLogs = strconv.FormatBool(false)

var ServiceLogsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceLogsCmdStr,
	ShortDescription:          "Get service logs",
	LongDescription:           "Show logs for a service inside an enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{
		{
			Key:       shouldFollowLogsFlagKey,
			Usage:     "Continues to follow the logs until stopped",
			Shorthand: "f",
			Type:      flags.FlagType{},
			Default:   defaultShouldFollowLogs,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
	},
	RunFunc:                   run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdStr, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using arg key '%v'", enclaveIdArgKey)
	}
	enclaveId := enclave.EnclaveID(enclaveIdStr)

	serviceGuidStr, err := args.GetNonGreedyArg(serviceGuidArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service GUID using arg key '%v'", serviceGuidArgKey)
	}
	serviceGuid := service.ServiceGUID(serviceGuidStr)

	shouldFollowLogs, err := flags.GetBool(shouldFollowLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the should-follow-logs flag using key '%v'", shouldFollowLogsFlagKey)
	}

	userServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			serviceGuid: true,
		},
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := kurtosisBackend.GetUserServiceLogs(ctx, enclaveId, userServiceFilters, shouldFollowLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}

	if len(erroredUserServiceGuids) > 0 {
		err, found := erroredUserServiceGuids[serviceGuid]
		if !found {
			return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, erroredUserServiceGuids)
		}
		return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with GUID '%v'", serviceGuid)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[serviceGuid]
	if !found {
		return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceReadCloserLog)
	}

	stdout := logrus.StandardLogger().Out
	// TODO This is a Docker library call; this should be pushed down into KurtosisBackend
	if _, err := stdcopy.StdCopy(stdout, stdout, userServiceReadCloserLog); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying the user service logs stream to STDOUT for user service with GUID '%v'",
			serviceGuid,
		)
	}

	return nil
}
