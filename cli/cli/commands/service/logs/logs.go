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
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
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

	userServiceGuids := map[services.ServiceGUID]bool{
		serviceGuid: true,
	}

	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return stacktrace.Propagate(err, "And error occurred getting the Kurtosis cluster config")
	}

	clusterType := clusterConfig.GetClusterType()


	//Print logs for Kubernetes cluster
	//TODO HUGE HACK!
	//TODO this is momentarily used here until we implement centralized logs for Kubernetes
	if clusterType == resolved_config.KurtosisClusterType_Kubernetes {
		//These Kurtosis primitives came from the backend (container-engine-lib) and this is the reason
		//why are different from the same defined earlier (which came from the Kurtosis SDK)
		kurtosisBackendEnclaveId := enclave.EnclaveID(enclaveIdStr)
		kurtosisBackendServiceGUID := service.ServiceGUID(serviceGuidStr)

		userServiceFilters := &service.ServiceFilters{
			GUIDs: map[service.ServiceGUID]bool{
				kurtosisBackendServiceGUID: true,
			},
		}

		successfulUserServiceLogs, erroredUserServiceGuids, err := kurtosisBackend.GetUserServiceLogs(ctx, kurtosisBackendEnclaveId, userServiceFilters, shouldFollowLogs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
		}
		defer func() {
			for _, userServiceLogsReadCloser := range successfulUserServiceLogs {
				if err := userServiceLogsReadCloser.Close(); err != nil{
					logrus.Warnf("We tried to close the user service logs read-closer-objects after we're done using it, but doing so threw an error:\n%v", err)
				}
			}
		}()

		if len(erroredUserServiceGuids) > 0 {
			err, found := erroredUserServiceGuids[kurtosisBackendServiceGUID]
			if !found {
				return stacktrace.NewError("Expected to find an error for user service with GUID '%v' on user service error map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", kurtosisBackendServiceGUID, erroredUserServiceGuids)
			}
			return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with GUID '%v'", kurtosisBackendServiceGUID)
		}

		userServiceReadCloserLog, found := successfulUserServiceLogs[kurtosisBackendServiceGUID]
		if !found {
			return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", kurtosisBackendServiceGUID, userServiceReadCloserLog)
		}

		if _, err := io.Copy(logrus.StandardLogger().Out, userServiceReadCloserLog); err != nil {
			return stacktrace.Propagate(err, "An error occurred copying the service logs to STDOUT")
		}

		return nil
	}

	//Print logs for Docker cluster which uses the Kurtosis centralized logs feature
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

	successfulUserServiceLogs, err := kurtosisCtx.StreamUserServiceLogs(ctx, enclaveId, userServiceGuids)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service logs from user services with GUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	}

	userServiceReadCloserLog, found := successfulUserServiceLogs[serviceGuid]
	if !found {
		return stacktrace.NewError("Expected to find logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceGuid, userServiceReadCloserLog)
	}

	if _, err := io.Copy(logrus.StandardLogger().Out, userServiceReadCloserLog); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying the service logs to STDOUT")
	}

	return nil
}
