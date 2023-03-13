/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"fmt"
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
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	enclaveIdentifierArgKey = "enclave-identifier"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	enclaveUUIDTitleName               = "UUID"
	enclaveNameTitleName               = "Enclave Name"
	enclaveStatusTitleName             = "Enclave Status"
	enclaveCreationTimeTitleName       = "Creation Time"
	apiContainerStatusTitleName        = "API Container Status"
	apiContainerHostGrpcPortTitle      = "API Container Host GRPC Port"
	apiContainerHostGrpcProxyPortTitle = "API Container Host GRPC Proxy Port"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	headerWidthChars = 100
	headerPadChar    = "="

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, showFullUuid bool, isAPIContainerRunning bool) error{
	"User Services": printUserServices,
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
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave identifier arg '%v' but none was found; this is a bug with Kurtosis!", enclaveIdentifierArgKey)
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave for identifier '%v'", enclaveIdentifier)
	}

	enclaveContainersStatus := enclaveInfo.ContainersStatus
	enclaveApiContainerStatus := enclaveInfo.ApiContainerStatus

	keyValuePrinter := output_printers.NewKeyValuePrinter()
	if showFullUuids {
		keyValuePrinter.AddPair(enclaveUUIDTitleName, enclaveInfo.GetEnclaveUuid())
	} else {
		keyValuePrinter.AddPair(enclaveUUIDTitleName, enclaveInfo.GetShortenedUuid())
	}
	keyValuePrinter.AddPair(enclaveNameTitleName, enclaveInfo.GetName())

	enclaveContainersStatusStr, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveContainersStatus)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
	}
	keyValuePrinter.AddPair(enclaveStatusTitleName, enclaveContainersStatusStr)

	enclaveCreationTime := enclaveInfo.GetCreationTime()
	//TODO remove this condition after 2023-01-01 when we are sure that there is not any old enclave created without the creation time label
	//TODO and add a fail loudly check
	if enclaveCreationTime != nil {
		enclaveCreationTimeStr := enclaveCreationTime.AsTime().Local().Format(time.RFC1123)

		keyValuePrinter.AddPair(enclaveCreationTimeTitleName, enclaveCreationTimeStr)
	}

	enclaveApiContainerStatusStr, err := enclave_status_stringifier.EnclaveAPIContainersStatusStringifier(enclaveApiContainerStatus)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when stringify enclave API containers status")
	}
	keyValuePrinter.AddPair(apiContainerStatusTitleName, enclaveApiContainerStatusStr)

	isApiContainerRunning := enclaveApiContainerStatus == kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING
	if isApiContainerRunning {
		apiContainerHostInfo := enclaveInfo.GetApiContainerHostMachineInfo()
		apiContainerHostGrpcPortInfoStr := fmt.Sprintf(
			"%v:%v",
			apiContainerHostInfo.GetIpOnHostMachine(),
			apiContainerHostInfo.GetGrpcPortOnHostMachine(),
		)
		apiContainerHostGrpcProxyPortInfoStr := fmt.Sprintf(
			"%v:%v",
			apiContainerHostInfo.GetIpOnHostMachine(),
			apiContainerHostInfo.GetGrpcProxyPortOnHostMachine(),
		)
		keyValuePrinter.AddPair(apiContainerHostGrpcPortTitle, apiContainerHostGrpcPortInfoStr)
		keyValuePrinter.AddPair(apiContainerHostGrpcProxyPortTitle, apiContainerHostGrpcProxyPortInfoStr)
	}
	keyValuePrinter.Print()
	out.PrintOutLn("")

	sortedEnclaveObjHeaders := []string{}
	for header := range enclaveObjectPrintingFuncs {
		sortedEnclaveObjHeaders = append(sortedEnclaveObjHeaders, header)
	}
	sort.Strings(sortedEnclaveObjHeaders)

	headersWithPrintErrs := []string{}
	for _, header := range sortedEnclaveObjHeaders {
		printingFunc, found := enclaveObjectPrintingFuncs[header]
		if !found {
			return stacktrace.NewError("No printing function found for enclave object '%v'; this is a bug in Kurtosis!", header)
		}

		numRunesInHeader := utf8.RuneCountInString(header) + 2 // 2 because there will be a space before and after the header
		numPadChars := (headerWidthChars - numRunesInHeader) / 2
		padStr := strings.Repeat(headerPadChar, numPadChars)
		fmt.Printf("%v %v %v\n", padStr, header, padStr)

		if err := printingFunc(ctx, kurtosisBackend, enclaveInfo, showFullUuids, isApiContainerRunning); err != nil {
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
