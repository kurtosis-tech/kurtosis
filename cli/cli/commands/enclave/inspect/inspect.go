/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package inspect

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	enclaveIdTitleName                 = "Enclave ID"
	enclaveStatusTitleName             = "Enclave Status"
	apiContainerStatusTitleName        = "API Container Status"
	apiContainerHostGrpcPortTitle      = "API Container Host GRPC Port"
	apiContainerHostGrpcProxyPortTitle = "API Container Host GRPC Proxy Port"

	headerWidthChars = 100
	headerPadChar    = "="

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var enclaveObjectPrintingFuncs = map[string]func(ctx context.Context, kurtosisBackend backend_interface.KurtosisBackend, enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo, isAPIContainerRunning bool) error{
	"User Services":    printUserServices,
	"Kurtosis Modules": printModules,
}

var EnclaveInspectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveInspectCmdStr,
	ShortDescription:          "Inspect an enclave",
	LongDescription:           "List information about the enclave's status and contents",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
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
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdStr, err := args.GetNonGreedyArg(enclaveIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for non-greedy enclave ID arg '%v' but none was found; this is a bug with Kurtosis!", enclaveIdArgKey)
	}

	getEnclavesResp, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves, which is necessary to display the state for enclave '%v'", enclaveIdStr)
	}

	enclaveInfo, found := getEnclavesResp.EnclaveInfo[enclaveIdStr]
	if !found {
		return stacktrace.NewError("No enclave with ID '%v' exists", enclaveIdStr)
	}

	enclaveContainersStatus := enclaveInfo.ContainersStatus
	enclaveApiContainerStatus := enclaveInfo.ApiContainerStatus

	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(enclaveIdTitleName, enclaveIdStr)
	enclaveContainersStatusStr, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveContainersStatus)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
	}
	keyValuePrinter.AddPair(enclaveStatusTitleName, enclaveContainersStatusStr)
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
	fmt.Fprintln(logrus.StandardLogger().Out, "")

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
		fmt.Printf("%v %v %v", padStr, header, padStr)

		if err := printingFunc(ctx, kurtosisBackend, enclaveInfo, isApiContainerRunning); err != nil {
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
