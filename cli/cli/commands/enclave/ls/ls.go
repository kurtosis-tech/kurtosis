/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"strings"
	"time"
)

const (
	enclaveIdColumnHeader           = "EnclaveID"
	enclaveStatusColumnHeader       = "Status"
	enclaveCreationTimeColumnHeader = "Creation Time"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	emptyTimeForOldEnclaves = ""
)

var EnclaveLsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveLsCmdStr,
	ShortDescription:          "Lists enclaves",
	LongDescription:           "Lists the enclaves running in the Kurtosis engine",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     nil,
	Args:                      nil,
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

	orderedEnclaveInfoMaps, enclaveWithoutCreationTimeInfoMap := getOrderedEnclaveInfoMapAndEnclaveWithoutCreationTimeMap(response.GetEnclaveInfo())
	tablePrinter := output_printers.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader, enclaveCreationTimeColumnHeader)

	//TODO remove this iteration after 2023-01-01 when we are sure that there is not any old enclave created without the creation time label
	//This is for retro-compatibility, for those old enclave did not track enclave's creation time
	for _, enclaveInfo := range enclaveWithoutCreationTimeInfoMap {
		enclaveId := enclaveInfo.GetEnclaveId()

		enclaveStatus, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status '%v'", enclaveInfo.GetContainersStatus())
		}

		if err := tablePrinter.AddRow(enclaveId, enclaveStatus, emptyTimeForOldEnclaves); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}
	//Retro-compatibility ends

	for _, enclaveInfo := range orderedEnclaveInfoMaps {

		enclaveId := enclaveInfo.GetEnclaveId()

		enclaveStatus, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status '%v'", enclaveInfo.GetContainersStatus())
		}

		enclaveCreationTime := enclaveInfo.CreationTime.AsTime().Local().Format(time.RFC1123)

		if err := tablePrinter.AddRow(enclaveId, enclaveStatus, enclaveCreationTime); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}

	tablePrinter.Print()

	return nil
}

func getOrderedEnclaveInfoMapAndEnclaveWithoutCreationTimeMap(
	enclaveInfoMap map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) (
	[]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
	map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) {

	orderedEnclaveInfoMaps := []*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}

	enclaveWithoutCreationTimeInfoMap := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}

	for enclaveIdStr, enclaveInfo := range enclaveInfoMap {
		//TODO remove this condition after 2023-01-01 when we are sure that there is not any old enclave created without the creation time label
		//This is for retro-compatibility, for those old enclave did not track enclave's creation time
		if enclaveInfo.GetCreationTime() == nil {
			enclaveWithoutCreationTimeInfoMap[enclaveIdStr] = enclaveInfo
			continue
		}
		orderedEnclaveInfoMaps = append(orderedEnclaveInfoMaps, enclaveInfo)
	}

	sort.Slice(orderedEnclaveInfoMaps, func(firstItemIndex, secondItemIndex int) bool {

		firstItemEnclaveCreationTime := orderedEnclaveInfoMaps[firstItemIndex].CreationTime.AsTime()
		secondItemEnclaveCreationTime := orderedEnclaveInfoMaps[secondItemIndex].CreationTime.AsTime()

		//First order by creation time
		if firstItemEnclaveCreationTime.Before(secondItemEnclaveCreationTime) {
			return true
		} else if firstItemEnclaveCreationTime.After(secondItemEnclaveCreationTime) {
			return false
		}

		//If creation time is equal order by enclave Id
		firstItemEnclaveIdStr := orderedEnclaveInfoMaps[firstItemIndex].EnclaveId
		secondItemEnclaveIdStr := orderedEnclaveInfoMaps[secondItemIndex].EnclaveId

		return strings.Compare(firstItemEnclaveIdStr, secondItemEnclaveIdStr) <= 0
	})

	return orderedEnclaveInfoMaps, enclaveWithoutCreationTimeInfoMap
}
