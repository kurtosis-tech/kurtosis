/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/enclave_status_stringifier"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"sort"
	"strings"
	"time"
)

const (
	enclaveUuidColumnHeader         = "UUID"
	enclaveStatusColumnHeader       = "Status"
	enclaveNameColumnHeader         = "Name"
	enclaveCreationTimeColumnHeader = "Creation Time"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	fullUuidsFlagKey       = "full-uuids"
	fullUuidFlagKeyDefault = "false"

	emptyTimeForOldEnclaves = ""
)

var EnclaveLsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveLsCmdStr,
	ShortDescription:          "Lists enclaves",
	LongDescription:           "Lists the enclaves running in the Kurtosis engine",
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
	Args:    nil,
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	_ *args.ParsedArgs,
) error {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves")
	}

	showFullUuids, err := flags.GetBool(fullUuidsFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for the '%v' flag but failed to get it", fullUuidsFlagKey)
	}

	tablePrinter := output_printers.NewTablePrinter(enclaveUuidColumnHeader, enclaveNameColumnHeader, enclaveStatusColumnHeader, enclaveCreationTimeColumnHeader)
	orderedEnclaveInfoMaps, enclaveWithoutCreationTimeInfoMap := getOrderedEnclaveInfoMapAndEnclaveWithoutCreationTimeMap(enclaves.GetEnclavesByUuid())

	//TODO remove this iteration after 2023-01-01 when we are sure that there is not any old enclave created without the creation time label
	//This is for retro-compatibility, for those old enclave did not track enclave's creation time
	for _, enclaveInfo := range enclaveWithoutCreationTimeInfoMap {
		enclaveUuid := enclaveInfo.GetEnclaveUuid()
		uuidToPrint := enclaveInfo.GetShortenedUuid()
		if showFullUuids {
			uuidToPrint = enclaveUuid
		}

		enclaveStatus, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status '%v'", enclaveInfo.GetContainersStatus())
		}

		if err := tablePrinter.AddRow(uuidToPrint, enclaveInfo.Name, enclaveStatus, emptyTimeForOldEnclaves); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveUuid)
		}
	}
	//Retro-compatibility ends

	for _, enclaveInfo := range orderedEnclaveInfoMaps {

		enclaveUuid := enclaveInfo.GetEnclaveUuid()
		uuidToPrint := enclaveInfo.GetShortenedUuid()
		if showFullUuids {
			uuidToPrint = enclaveUuid
		}

		enclaveStatus, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status '%v'", enclaveInfo.GetContainersStatus())
		}

		// The extra space is a hack till we figure out the table printer color + formatting story
		enclaveCreationTime := " " + enclaveInfo.CreationTime.AsTime().Local().Format(time.RFC1123)

		enclaveName := enclaveInfo.GetName()

		if err := tablePrinter.AddRow(uuidToPrint, enclaveName, enclaveStatus, enclaveCreationTime); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveUuid)
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

		//If creation time is equal order by enclave name
		firstItemEnclaveNameStr := orderedEnclaveInfoMaps[firstItemIndex].GetName()
		secondItemEnclaveNameStr := orderedEnclaveInfoMaps[secondItemIndex].GetName()

		return strings.Compare(firstItemEnclaveNameStr, secondItemEnclaveNameStr) <= 0
	})

	return orderedEnclaveInfoMaps, enclaveWithoutCreationTimeInfoMap
}
