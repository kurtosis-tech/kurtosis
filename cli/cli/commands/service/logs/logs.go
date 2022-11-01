/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strconv"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	serviceGuidArgKey = "service-guid"

	shouldFollowLogsFlagKey = "follow"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var defaultShouldFollowLogs = strconv.FormatBool(false)

var ServiceLogsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceLogsCmdStr,
	ShortDescription:          "Get service logs",
	LongDescription:           "Show logs for a service inside an enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:       shouldFollowLogsFlagKey,
			Usage:     "Continues to follow the logs until stopped",
			Shorthand: "f",
			Type:      flags.FlagType_Bool,
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
		// TODO Create a NewServiceIDArg that adds autocomplete
		{
			Key: serviceGuidArgKey,
		},
	},
	RunFunc: run,
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
	enclaveId := enclaves.EnclaveID(enclaveIdStr)

	serviceGuidStr, err := args.GetNonGreedyArg(serviceGuidArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service GUID using arg key '%v'", serviceGuidArgKey)
	}
	serviceGuid := services.ServiceGUID(serviceGuidStr)

	shouldFollowLogs, err := flags.GetBool(shouldFollowLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the should-follow-logs flag using key '%v'", shouldFollowLogsFlagKey)
	}

	userServiceGuids := map[services.ServiceGUID]bool{
		serviceGuid: true,
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	if !shouldFollowLogs {

		userServiceLogs, err := kurtosisCtx.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting user service logs from user services with GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
		}

		serviceLogs, found := userServiceLogs[serviceGuid]
		if !found {
			return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceLogs)
		}

		for _, logLine := range serviceLogs {
			if _, err := fmt.Fprintln(logrus.StandardLogger().Out, logLine); err != nil {
				logrus.Errorf("We tried to print the user service log line '%v', but doing so threw an error:\n%v",logLine, err)
			}
		}

		return nil
	}

	userServiceLogsByGuidChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.StreamUserServiceLogs(ctx, enclaveId, userServiceGuids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred streaming user service logs from user services with GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	}
	defer cancelStreamUserServiceLogsFunc()

	for {
		userServiceLogsByGuid, isChanOpen := <-userServiceLogsByGuidChan
		if !isChanOpen {
			break
		}

		userServiceLogs, found := userServiceLogsByGuid[serviceGuid]
		if !found {
			return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceLogsByGuid)
		}

		for _, serviceLog := range userServiceLogs {
			fmt.Println(serviceLog.GetContent())
		}

	}

	return nil
}
