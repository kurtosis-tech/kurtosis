/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	enclaveUUIDTitleName         = "UUID"
	enclaveNameTitleName         = "Name"
	enclaveStatusTitleName       = "Status"
	enclaveCreationTimeTitleName = "Creation Time"
	flagsTitleName               = "Flags"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	headerWidthChars = 100
	headerPadChar    = "="

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	userServicesArtifactsHeader = "User Services"
	filesArtifactsHeader        = "Files Artifacts"

	productionEnclaveFlagStr = "production"
)

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuid bool, isAPIContainerRunning bool) error{
	userServicesArtifactsHeader: printUserServices,
	filesArtifactsHeader:        printFilesArtifacts,
}

var EnclaveInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveInspectCmdStr,
	ShortDescription:          "Inspect an enclave",
	LongDescription:           "List information about the enclave's status and contents",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     fullUuidsFlagKey,
			Usage:   "If true then Kurtosis prints full UUIDs instead of shortened UUIDs. Default false.",
			Type:    flags.FlagType_Bool,
			Default: fullUuidFlagKeyDefault,
		},
	},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
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
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug in the Kurtosis CLI!", enclaveIdentifierArgKey)
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	if err = PrintEnclaveInspect(ctx, kurtosisCtx, enclaveIdentifier, showFullUuids); err != nil {
		// this is already wrapped up
		return err
	}
	return nil
}

func PrintEnclaveInspect(ctx context.Context, kurtosisCtx *kurtosis_context.KurtosisContext, enclaveIdentifier string, showFullUuids bool) error {
	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave for identifier '%v'", enclaveIdentifier)
	}

	keyValuePrinter := output_printers.NewKeyValuePrinter()

	// Add title row
	keyValuePrinter.AddPair(enclaveNameTitleName, enclaveInfo.GetName())

	// Add UUID row
	if showFullUuids {
		keyValuePrinter.AddPair(enclaveUUIDTitleName, enclaveInfo.GetEnclaveUuid())
	} else {
		keyValuePrinter.AddPair(enclaveUUIDTitleName, enclaveInfo.GetShortenedUuid())
	}

	// Add status row
	enclaveContainersStatusStr, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
	}
	keyValuePrinter.AddPair(enclaveStatusTitleName, enclaveContainersStatusStr)

	// Add creation time row
	enclaveCreationTime := enclaveInfo.GetCreationTime()
	if enclaveCreationTime == nil {
		return stacktrace.NewError("Expected to get the enclave creation time from the enclave info received but it was not received, this is a bug in Kurtosis")
	}
	enclaveCreationTimeStr := enclaveCreationTime.AsTime().Local().Format(time.RFC1123)
	keyValuePrinter.AddPair(enclaveCreationTimeTitleName, enclaveCreationTimeStr)

	// Add flags row
	allEnclaveFlagsStr := getAllEnclaveFlagsStr(enclaveInfo)

	keyValuePrinter.AddPair(flagsTitleName, allEnclaveFlagsStr)

	isApiContainerRunning := enclaveInfo.GetApiContainerStatus() == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING

	keyValuePrinter.Print()
	out.PrintOutLn("")

	sortedEnclaveObjHeaders := []string{}
	for header := range enclaveObjectPrintingFuncs {
		sortedEnclaveObjHeaders = append(sortedEnclaveObjHeaders, header)
	}
	sort.Strings(sortedEnclaveObjHeaders)

	headersWithPrintErrs := []string{}
	for _, header := range sortedEnclaveObjHeaders {
		if header == filesArtifactsHeader && !isApiContainerRunning {
			// can't fetch files artifact information if APIC isn't running
			continue
		}

		printingFunc, found := enclaveObjectPrintingFuncs[header]
		if !found {
			return stacktrace.NewError("No printing function found for enclave object '%v'; this is a bug in Kurtosis!", header)
		}

		numRunesInHeader := utf8.RuneLen(' ') + utf8.RuneCountInString(header) + utf8.RuneLen(' ') // there will be a space before and after the header
		numPadChars := (headerWidthChars - numRunesInHeader) / 2                                   //nolint:mnd
		padStr := strings.Repeat(headerPadChar, numPadChars)
		fmt.Printf("%v %v %v\n", padStr, header, padStr)

		if err := printingFunc(ctx, kurtosisCtx, enclaveInfo, showFullUuids, isApiContainerRunning); err != nil {
			logrus.Error(err)
			headersWithPrintErrs = append(headersWithPrintErrs, header)
		}
		fmt.Println("")
	}

	if len(headersWithPrintErrs) > 0 {
		return stacktrace.NewError(
			"Errors occurred printing the following enclave elements: %v",
			strings.Join(headersWithPrintErrs, ", "),
		)
	}

	return nil
}

func getAllEnclaveFlagsStr(enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo) string {
	allEnclaveFragsStr := ""

	// we only one enclave flag added so far, but we could have more in the future, so we should add them here
	// and return all of them together in just one string
	currentModeStr := strings.ToLower(enclaveInfo.GetMode().String())

	if currentModeStr == productionEnclaveFlagStr {
		allEnclaveFragsStr = currentModeStr
	}

	return allEnclaveFragsStr
}
