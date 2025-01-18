/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package logs

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"strconv"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = true // don't need to pass this in if they use the return all services flag
	isServiceIdentifierArgGreedy   = true

	shouldFollowLogsFlagKey  = "follow"
	returnAllServiceLogs     = "all-services"
	allServicesWildcard      = "*"
	returnNumLogsFlagKey     = "num"
	returnAllLogsFlagKey     = "all"
	matchTextFilterFlagKey   = "match"
	matchRegexFilterFlagKey  = "regex-match"
	invertMatchFilterFlagKey = "invert-match"

	defaultMatchTextOrRegexFilterFlagValue = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	interruptChanBufferSize = 5

	defaultNumLogLines            = 200
	commonInstructionInMatchFlags = "Important: " + matchTextFilterFlagKey + " and " + matchRegexFilterFlagKey + " flags cannot be used at the same time. You should either use one or the other."
)

type ColorPrinter func(a ...interface{}) string

var colorList = []ColorPrinter{
	color.New(color.FgBlue).SprintFunc(),
	color.New(color.FgCyan).SprintFunc(),
	color.New(color.FgGreen).SprintFunc(),
	color.New(color.FgMagenta).SprintFunc(),
	color.New(color.FgYellow).SprintFunc(),
	color.New(color.FgHiRed).SprintFunc(),
	color.New(color.FgHiBlue).SprintFunc(),
	color.New(color.FgHiWhite).SprintFunc(),
}

var doNotFilterLogLines *kurtosis_context.LogLineFilter = nil

var defaultShouldFollowLogs = strconv.FormatBool(false)
var defaultInvertMatchFilterFlagValue = strconv.FormatBool(false)
var defaultShouldReturnAllLogs = strconv.FormatBool(false)
var defaultShouldReturnAllServiceLog = strconv.FormatBool(false)
var defaultNumLogLinesFlagValue = strconv.Itoa(defaultNumLogLines)

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
			Key:       returnAllLogsFlagKey,
			Usage:     "Gets all logs.",
			Shorthand: "a",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldReturnAllLogs,
		},
		{
			Key:       returnNumLogsFlagKey,
			Usage:     "Get the last X log lines.",
			Shorthand: "n",
			Type:      flags.FlagType_Uint32,
			Default:   defaultNumLogLinesFlagValue,
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
		{
			Key:       returnAllServiceLogs,
			Usage:     "Returns service log streams for all logs in an enclave",
			Shorthand: "x",
			Type:      flags.FlagType_Bool,
			Default:   defaultShouldReturnAllServiceLog,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewHistoricalEnclaveIdentifiersArgWithValidationDisabled(
			enclaveIdentifierArgKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewHistoricalServiceIdentifierArgWithValidationDisabled(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
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

	shouldReturnAllServiceLogs, err := flags.GetBool(returnAllServiceLogs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the 'all-services' flag using key '%v'", returnAllServiceLogs)
	}

	var serviceIdentifiers []string
	serviceIdentifiers, err = args.GetGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier using arg key '%v'", serviceIdentifierArgKey)
	}
	// if no service identifiers were passed or just the wildcard was passed, default to returning all
	if len(serviceIdentifiers) == 0 || (len(serviceIdentifiers) == 1 && serviceIdentifiers[0] == allServicesWildcard) { //
		shouldReturnAllServiceLogs = true
	}

	shouldFollowLogs, err := flags.GetBool(shouldFollowLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the should-follow-logs flag using key '%v'", shouldFollowLogsFlagKey)
	}

	shouldReturnAllLogs, err := flags.GetBool(returnAllLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the 'all' flag using key '%v'", returnAllLogsFlagKey)
	}

	numLogLines, err := flags.GetUint32(returnNumLogsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the 'num' flag using key '%v'", returnNumLogsFlagKey)
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

	if shouldReturnAllServiceLogs {
		enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred retrieving enclave context for '%v'", enclaveIdentifier)
		}

		allServiceIdentifiers, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred retrieving service identifiers for enclave '%v'", enclaveIdentifier)
		}
		serviceIdentifiers = allServiceIdentifiers.GetOrderedListOfNames()
	}

	userServiceUuids := map[services.ServiceUUID]bool{}
	serviceUuids := map[services.ServiceUUID]string{}
	serviceColorPrinterMap := map[string]ColorPrinter{}
	for idx, serviceIdentifier := range serviceIdentifiers {
		serviceUuid := getEnclaveAndServiceUuidForIdentifiers(kurtosisCtx, ctx, enclaveIdentifier, serviceIdentifier)
		serviceUuids[serviceUuid] = serviceIdentifier
		serviceColorPrinterMap[serviceIdentifier] = colorList[idx%len(colorList)]
		userServiceUuids[serviceUuid] = true
	}

	logLineFilter, err := getLogLineFilterFromFilterFlagValues(matchTextStr, matchRegexStr, invertMatch)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the log line filter using these filter flag values '%s=%s', '%s=%s', '%s=%v'", matchTextFilterFlagKey, matchTextStr, matchRegexFilterFlagKey, matchRegexStr, invertMatchFilterFlagKey, invertMatch)
	}

	serviceLogsStreamContentChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, enclaveIdentifier, userServiceUuids, shouldFollowLogs, shouldReturnAllLogs, numLogLines, logLineFilter)
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
			for serviceUuid, serviceIdentifier := range serviceUuids {
				userServiceLogs, found := userServiceLogsByUuid[serviceUuid]
				if !found {
					return stacktrace.NewError("Expected to find logs for user service with UUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", serviceUuid, userServiceLogsByUuid)
				}

				for _, serviceLog := range userServiceLogs {
					colorPrinter := serviceColorPrinterMap[serviceIdentifier]
					out.PrintOutLn(fmt.Sprintf("[%v] %v", colorPrinter(serviceIdentifier), serviceLog.GetContent()))
				}
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

// This function works makes the best effort to get the most accurate enclave uuid and service uuid for the passed values
// defaults to assuming the passed value are uuids
// this function will be a lot cleaner after the object ids are stored in a database
// https://github.com/kurtosis-tech/kurtosis/issues/343
func getEnclaveAndServiceUuidForIdentifiers(kurtosisCtx *kurtosis_context.KurtosisContext, ctx context.Context, enclaveIdentifier string, serviceIdentifier string) services.ServiceUUID {

	serviceUuid := services.ServiceUUID(serviceIdentifier)

	// we need to get the enclave context to get the historical service identifiers for the enclave
	// we also need it to get the `ServiceContext`
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		// we couldn't get the enclaveCtx we go with service id passed as the serviceUuid
		// for the enclave either we got it from the historical values or we return what was passed to the function
		return serviceUuid
	}

	// we see if we can get historical identifiers
	serviceIdentifiers, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	if err == nil {
		// if we can find a match for the service in the historical values, we return that
		serviceUuidFromHistoricalServices, err := serviceIdentifiers.GetServiceUuidForIdentifier(serviceIdentifier)
		if err == nil {
			return serviceUuidFromHistoricalServices
		}
	}

	// otherwise we see if we can find the service in the currently running services
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err == nil {
		return serviceCtx.GetServiceUUID()
	}

	// we return the best matches so far for enclave and service uuids
	return serviceUuid
}
