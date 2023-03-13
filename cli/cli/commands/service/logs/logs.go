/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/kurtosis_config_getter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_config/resolved_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/cli/cli/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface/objects/service"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/signal"
	"strconv"
)

const (
	enclaveIdentifierArgKey = "enclave-identifier"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service-identifier"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false

	shouldFollowLogsFlagKey  = "follow"
	matchTextFilterFlagKey   = "match"
	matchRegexFilterFlagKey  = "regex-match"
	invertMatchFilterFlagKey = "invert-match"

	defaultMatchTextOrRegexFilterFlagValue = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	interruptChanBufferSize = 5

	commonInstructionInMatchFlags = "Important: " + matchTextFilterFlagKey + " and " + matchRegexFilterFlagKey + " flags cannot be used at the same time. You should either use one or the other."
)

var doNotFilterLogLines *kurtosis_context.LogLineFilter = nil

var defaultShouldFollowLogs = strconv.FormatBool(false)
var defaultInvertMatchFilterFlagValue = strconv.FormatBool(false)

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
		{
			Key: matchTextFilterFlagKey,
			Usage: fmt.Sprintf(
				"Filter the log lines returning only those containing this match. %s",
				commonInstructionInMatchFlags,
			),
			Default: defaultMatchTextOrRegexFilterFlagValue,
		},
		{
			Key: matchRegexFilterFlagKey,
			Usage: fmt.Sprintf(
				"Filter the log lines returning only those containing this regex expression match (re2 syntax regex may be used, more here: %s). This filter will always work but will have degraded performance for tokens. %s",
				user_support_constants.GoogleRe2SyntaxDocumentation,
				commonInstructionInMatchFlags,
			),
			Default: defaultMatchTextOrRegexFilterFlagValue,
		},
		{
			Key: invertMatchFilterFlagKey,
			Usage: fmt.Sprintf(
				"Inverts the filter condition specified by either '%s' or '%s'. Log lines NOT containing %s/%s will be returned",
				matchTextFilterFlagKey,
				matchRegexFilterFlagKey,
				matchTextFilterFlagKey,
				matchRegexFilterFlagKey,
			),
			Shorthand: "v",
			Type:      flags.FlagType_Bool,
			Default:   defaultInvertMatchFilterFlagValue,
		},
	},
	Args: []*args.ArgConfig{
		//TODO disabling enclaveID validation and serviceUUID validation for allowing consuming logs from removed or stopped enclaves
		//TODO we should enable them when #879 is ready: https://github.com/kurtosis-tech/kurtosis/issues/879
		enclave_id_arg.NewHistoricalEnclaveIdentifiersArgWithValidationDisabled(
			enclaveIdentifierArgKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		// TODO use the `NewServiceIdentifierArg` instead when we start storing identifiers in DB
		// TODO we should fix this after https://github.com/kurtosis-tech/kurtosis/issues/879
		service_identifier_arg.NewHistoricalServiceIdentifierArgWithValidationDisabled(
			serviceIdentifierArgKey,
			isServiceIdentifierArgOptional,
			isServiceIdentifierArgGreedy,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using arg key '%v'", enclaveIdentifierArgKey)
	}

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier using arg key '%v'", serviceIdentifierArgKey)
	}

	shouldFollowLogs, err := flags.GetBool(shouldFollowLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the should-follow-logs flag using key '%v'", shouldFollowLogsFlagKey)
	}

	matchTextStr, err := flags.GetString(matchTextFilterFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the match flag using key '%v'", matchTextFilterFlagKey)
	}

	matchRegexStr, err := flags.GetString(matchRegexFilterFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the regex-match flag using key '%v'", matchRegexFilterFlagKey)
	}

	invertMatch, err := flags.GetBool(invertMatchFilterFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the invert match flag using key '%v'", invertMatchFilterFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveUuid, serviceUuid := getEnclaveAndServiceUuidForIdentifiers(kurtosisCtx, ctx, enclaveIdentifier, serviceIdentifier)

	userServiceUuids := map[services.ServiceUUID]bool{
		serviceUuid: true,
	}

	clusterConfig, err := kurtosis_config_getter.GetKurtosisClusterConfig()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis cluster config")
	}

	clusterType := clusterConfig.GetClusterType()

	//Print logs for Kubernetes cluster
	//TODO HUGE HACK!
	//TODO this is momentarily used here until we implement the GetServiceLogs method in the Kurtosis engine's gateway
	//TODO related ticket: https://github.com/kurtosis-tech/kurtosis/issues/1069
	if clusterType == resolved_config.KurtosisClusterType_Kubernetes {
		//These Kurtosis primitives came from the backend (container-engine-lib) and this is the reason
		//why are different from the same defined earlier (which came from the Kurtosis SDK)
		kurtosisBackendServiceUUID := service.ServiceUUID(serviceIdentifier)

		userServiceFilters := &service.ServiceFilters{
			Names: nil,
			UUIDs: map[service.ServiceUUID]bool{
				kurtosisBackendServiceUUID: true,
			},
			Statuses: nil,
		}

		successfulUserServiceLogs, erroredUserServiceUuids, err := kurtosisBackend.GetUserServiceLogs(ctx, enclaveUuid, userServiceFilters, shouldFollowLogs)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting user service logs using filters '%+v'", userServiceFilters)
		}
		defer func() {
			for _, userServiceLogsReadCloser := range successfulUserServiceLogs {
				if err := userServiceLogsReadCloser.Close(); err != nil {
					logrus.Warnf("We tried to close the user service logs read-closer-objects after we're done using it, but doing so threw an error:\n%v", err)
				}
			}
		}()

		if len(erroredUserServiceUuids) > 0 {
			err, found := erroredUserServiceUuids[kurtosisBackendServiceUUID]
			if !found {
				return stacktrace.NewError("Expected to find an error for user service with UUID '%v' on user service error map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", kurtosisBackendServiceUUID, erroredUserServiceUuids)
			}
			return stacktrace.Propagate(err, "An error occurred getting user service logs for user service with UUID '%v'", kurtosisBackendServiceUUID)
		}

		userServiceReadCloserLog, found := successfulUserServiceLogs[kurtosisBackendServiceUUID]
		if !found {
			return stacktrace.NewError("Expected to find logs for user service with UUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", kurtosisBackendServiceUUID, userServiceReadCloserLog)
		}

		if _, err := io.Copy(out.GetOut(), userServiceReadCloserLog); err != nil {
			return stacktrace.Propagate(err, "An error occurred copying the service logs to STDOUT")
		}

		return nil
	}

	//Print logs for Docker cluster which uses the Kurtosis centralized logs feature
	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the log line filter using these filter flag values '%s=%s', '%s=%s', '%s=%v'", matchTextFilterFlagKey, matchTextStr, matchRegexFilterFlagKey, matchRegexStr, invertMatchFilterFlagKey, invertMatch)
	}

	serviceLogsStreamContentChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, enclaveIdentifier, userServiceUuids, shouldFollowLogs, logLineFilter)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting user service logs from user services with UUIDs '%+v' in enclave '%v' and with follow logs value '%v'", userServiceUuids, enclaveIdentifier, shouldFollowLogs)
	}
	defer cancelStreamUserServiceLogsFunc()

	// This channel will receive a signal when the user presses an interrupt
	interruptChan := make(chan os.Signal, interruptChanBufferSize)
	signal.Notify(interruptChan, os.Interrupt)

	for {
		select {
		case serviceLogsStreamContent, isChanOpen := <-serviceLogsStreamContentChan:
			if !isChanOpen {
				return nil
			}

			notFoundServiceUuids := serviceLogsStreamContent.GetNotFoundServiceUuids()

			for notFoundServiceUuid := range notFoundServiceUuids {
				logrus.Warnf("The Kurtosis centralized logs system does not contains any logs for the requested service UUID '%v'. This means that a service with that UUID either doesn't exist, or hasn't sent any logs. You can wait for further responses, or cancel the stream with Ctrl + C.", notFoundServiceUuid)
			}

			userServiceLogsByUuid := serviceLogsStreamContent.GetServiceLogsByServiceUuids()

			userServiceLogs, found := userServiceLogsByUuid[serviceUuid]
			if !found {
				return stacktrace.NewError("Expected to find logs for user service with UUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceUuid, userServiceLogsByUuid)
			}

			for _, serviceLog := range userServiceLogs {
				out.PrintOutLn(serviceLog.GetContent())
			}
		case <-interruptChan:
			logrus.Debugf("Received signal interruption in service logs Kurtosis CLI command")
			return nil
		}
	}
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func getLogLineFilterFromFilterFlagValues(matchTextStr string, matchRegexStr string, invertMatch bool) (*kurtosis_context.LogLineFilter, error) {

	if matchTextStr == defaultMatchTextOrRegexFilterFlagValue && matchRegexStr == defaultMatchTextOrRegexFilterFlagValue {
		return doNotFilterLogLines, nil
	}

	if matchTextStr != defaultMatchTextOrRegexFilterFlagValue && matchRegexStr != defaultMatchTextOrRegexFilterFlagValue {
		return nil, stacktrace.NewError("Both filter match have being sent '%s', '%s' and it is not allowed, it's allowed to send only one of them", matchTextFilterFlagKey, matchRegexFilterFlagKey)
	}

	if matchTextStr != defaultMatchTextOrRegexFilterFlagValue && !invertMatch {
		return kurtosis_context.NewDoesContainTextLogLineFilter(matchTextStr), nil
	}

	if matchTextStr != defaultMatchTextOrRegexFilterFlagValue && invertMatch {
		return kurtosis_context.NewDoesNotContainTextLogLineFilter(matchTextStr), nil
	}

	if matchRegexStr != defaultMatchTextOrRegexFilterFlagValue && !invertMatch {
		return kurtosis_context.NewDoesContainMatchRegexLogLineFilter(matchRegexStr), nil
	}

	if matchRegexStr != defaultMatchTextOrRegexFilterFlagValue && invertMatch {
		return kurtosis_context.NewDoesNotContainMatchRegexLogLineFilter(matchRegexStr), nil
	}

	return nil, stacktrace.NewError(
		"Something goes wrong during the log line filter definition using these filter flag values '%s=%s', '%s=%s' and '%s=%v', then it was not able to define it, this should never happens; this is a bug in Kurtosis",
		matchTextFilterFlagKey, matchTextStr, matchRegexFilterFlagKey, matchRegexStr, invertMatchFilterFlagKey, invertMatch,
	)
}

// This function works makes a best effort to get the most accurate enclave uuid and service uuid for the passed valeus
// defaults to assuming the passed value are uuids
// this function will be a lot cleaner after the object ids are stored in a database
// https://github.com/kurtosis-tech/kurtosis/issues/343
func getEnclaveAndServiceUuidForIdentifiers(kurtosisCtx *kurtosis_context.KurtosisContext, ctx context.Context, enclaveIdentifier string, serviceIdentifier string) (enclave.EnclaveUUID, services.ServiceUUID) {
	enclaveUuid := enclave.EnclaveUUID(enclaveIdentifier)
	serviceUuid := services.ServiceUUID(serviceIdentifier)
	// check if we can can find the UUID in the historical enclave identifiers
	historicalEnclaveIdentifiers, err := kurtosisCtx.GetExistingAndHistoricalEnclaveIdentifiers(ctx)
	if err == nil {
		enclaveUuidFromHistoricalValues, err := historicalEnclaveIdentifiers.GetEnclaveUuidForIdentifier(enclaveIdentifier)
		if err == nil {
			// set the enclaveUuid to the one from the historical identifiers in case there is a match
			enclaveUuid = enclave.EnclaveUUID(enclaveUuidFromHistoricalValues)
		}
	}

	// we need to get the enclave context to get the historical service identifiers for the enclave
	// we also need it to get the `ServiceContext`
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		// we couldn't get the enclaveCtx we go with service id passed as the serviceUuid
		// for the enclave either we got it from the historical values or we return what was passed to the function
		return enclaveUuid, serviceUuid
	}

	// we see if we can get historical identifiers
	serviceIdentifiers, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	if err == nil {
		// if we can find a match for the service in the historical values, we return that
		serviceUuidFromHistoricalServices, err := serviceIdentifiers.GetServiceUuidForIdentifier(serviceIdentifier)
		if err == nil {
			return enclaveUuid, serviceUuidFromHistoricalServices
		}
	}

	// otherwise we see if we can find the service in the currently running services
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err == nil {
		return enclaveUuid, serviceCtx.GetServiceUUID()
	}

	// we return the best matches so far for enclave and service uuids
	return enclaveUuid, serviceUuid
}
