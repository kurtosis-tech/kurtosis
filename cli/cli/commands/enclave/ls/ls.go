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
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
	"time"
)

const (
	enclaveIdColumnHeader          = "EnclaveID"
	enclaveStatusColumnHeader      = "Status"
	enclaveCreationTimeColumnHeader = "Creation Time"

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

/*
	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]string{}
	for enclaveId, enclaveInfo := range enclaveInfoMap {
		enclaveInfo.GetCreationTime()
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		enclaveStatuses[enclaveId], err = enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status")
		}
	}
	sort.Strings(orderedEnclaveIds)*/

	//TODO remove this, printing the received times

	for _, v := range  response.GetEnclaveInfo() {
		logrus.Infof("Creation times...")
		logrus.Infof("%v", v.GetCreationTime())
		logrus.Infof("%v", v.GetCreationTime().String())
		logrus.Infof("%v", v.GetCreationTime().AsTime())
		logrus.Infof("%v", v.GetCreationTime().AsTime().String())
	}

	orderedEnclaveCreationTimes, enclaveInfoByCreationTime := getOrderedEnclaveCreationTimesAndEnclaveInfoMap(response.GetEnclaveInfo())

	tablePrinter := output_printers.NewTablePrinter(enclaveIdColumnHeader, enclaveStatusColumnHeader, enclaveCreationTimeColumnHeader)
	for _, enclaveCreationTime := range orderedEnclaveCreationTimes {
		enclaveInfo, found := enclaveInfoByCreationTime[enclaveCreationTime]
		if !found {
			return stacktrace.NewError("Not found error")//TODO fix this message
		}
		enclaveId := enclaveInfo.GetEnclaveId()

		enclaveStatus, err := enclave_status_stringifier.EnclaveContainersStatusStringifier(enclaveInfo.GetContainersStatus())
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred when stringify enclave containers status '%v'", enclaveInfo.GetContainersStatus())
		}

		if err := tablePrinter.AddRow(enclaveId, enclaveStatus, enclaveCreationTime.String()); err != nil {
			return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveId)
		}
	}
	tablePrinter.Print()

	return nil
}

func getOrderedEnclaveCreationTimesAndEnclaveInfoMap(
	enclaveInfoMap map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) (
	[]time.Time,
	map[time.Time]*kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) {

	orderedEnclaveCreationTimes := []time.Time{}

	enclaveInfoByCreationTime := map[time.Time]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}

	for _, enclaveInfo := range enclaveInfoMap {
		enclaveCreationTime := enclaveInfo.GetCreationTime().AsTime()
		orderedEnclaveCreationTimes = append(orderedEnclaveCreationTimes, enclaveCreationTime)
		enclaveInfoByCreationTime[enclaveCreationTime] = enclaveInfo
	}

	sort.Slice(orderedEnclaveCreationTimes, func(i, j int) bool {
		return orderedEnclaveCreationTimes[i].Before(orderedEnclaveCreationTimes[j])
	})


	return orderedEnclaveCreationTimes, enclaveInfoByCreationTime
}