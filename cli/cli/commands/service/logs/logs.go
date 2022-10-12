/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"bytes"
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
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
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

	//TODO This is a huge temporary hack until we implement stream logs from the centralized logs database
	if !shouldFollowLogs {
		userServiceGuids := map[services.ServiceGUID]bool{
			serviceGuid: true,
		}

		kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
		}

		userServiceLogs, err := kurtosisCtx.GetUserServiceLogs(ctx, enclaveId, userServiceGuids)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting user service logs from user services with GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
		}

		serviceLogs, found := userServiceLogs[serviceGuid]
		if !found {
			return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceLogs)
		}

		logLineBuffer := bytes.NewBuffer([]byte{})
		for _, logLine := range serviceLogs {
			logLineWithLineBreak := fmt.Sprintf("%v\n", logLine)
			if _, err := logLineBuffer.WriteString(logLineWithLineBreak); err != nil {
				return stacktrace.Propagate(err, "An error occurred writing the service logs to the buffer")
			}
		}
		if _, err := logrus.StandardLogger().Out.Write(logLineBuffer.Bytes()); err != nil {
			return stacktrace.Propagate(err, "An error occurred writing the service logs to STDOUT")
		}

		return nil
	}

	//These Kurtosis primitives came from the backend (container-engine-lib) and this is the reason
	//why are different from the same defined on top (which came from the Kurtosis SDK)
	//TODO these are momentarily used here until we implement stream logs from the centralized logs database
	kurtosisBackendEnclaveId := enclave.EnclaveID(enclaveIdStr)
	kurtosisBackendServiceGuid := service.ServiceGUID(serviceGuidStr)

	userServiceFilters := &service.ServiceFilters{
		GUIDs: map[service.ServiceGUID]bool{
			kurtosisBackendServiceGuid: true,
		},
	}

	successfulUserServiceLogs, erroredUserServiceGuids, err := kurtosisBackend.GetUserServiceLogs(ctx, kurtosisBackendEnclaveId, userServiceFilters, shouldFollowLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
	}
	defer func() {
		for _, readCloser := range successfulUserServiceLogs {
			if err := readCloser.Close(); err != nil{
				logrus.Warnf("We tried to close the user service logs read-closer-objects after we're done using it, but doing so threw an error:\n%v", err)
			}
		}
	}()

	if len(erroredUserServiceGuids) > 0 {
		err, found := erroredUserServiceGuids[kurtosisBackendServiceGuid]
		if !found {
			return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, erroredUserServiceGuids)
		}
		return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with GUID '%v'", serviceGuid)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[kurtosisBackendServiceGuid]
	if !found {
		return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceReadCloserLog)
	}

	if _, err := io.Copy(logrus.StandardLogger().Out, userServiceReadCloserLog); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying the service logs to STDOUT")
	}

	return nil
}
