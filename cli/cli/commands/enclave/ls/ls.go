/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
)

const (
	enclaveIdColumnHeader     = "EnclaveID"
	enclaveStatusColumnHeader = "Status"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var EnclaveLsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveLsCmdStr,
	ShortDescription:          "Lists enclaves",
	LongDescription:           "Lists the enclaves running in the Kurtosis engine",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	RunFunc:                   run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
	response, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves")
	}
	enclaveInfoMap := response.GetEnclaveInfo()

	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]string{}
	for enclaveId, enclaveInfo := range enclaveInfoMap {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		enclaveStatuses[enclaveId], err = enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
		}
	}
	sort.Strings(orderedEnclaveIds)

	tablePrinter := output_printers.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader)
	for _, enclaveId := range orderedEnclaveIds {
		enclaveStatus, found := enclaveStatuses[enclaveId]
		if !found {
			return stacktrace.NewError("We're about to print enclave '%v', but it doesn't have a status; this is a bug in Kurtosis!", enclaveId)
		}
		if err := tablePrinter.AddRow(enclaveId, string(enclaveStatus)); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}
	tablePrinter.Print()

	return nil
}
