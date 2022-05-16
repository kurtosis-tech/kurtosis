/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package ls

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
)

const (
	enclaveIdColumnHeader     = "EnclaveID"
	enclaveStatusColumnHeader = "Status"
)

// TODO SWITCH TO BE AN ENGINE-CONSUMING COMAMND
var LsCmd = &cobra.Command{
	Use:   command_str_consts.EnclaveLsCmdStr,
	Short: "List Kurtosis enclaves",
	RunE:  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	kurtosisBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend connected to local Docker")
	}
	engineManager := engine_manager.NewEngineManager(kurtosisBackend)

	engineClient, closeClientFunc, err := engineManager.StartEngineIdempotentlyWithDefaultVersion(ctx, defaults.DefaultEngineLogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new Kurtosis engine client")
	}
	defer closeClientFunc()

	response, err := engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves")
	}
	enclaveInfoMap := response.GetEnclaveInfo()

	orderedEnclaveIds := []string{}
	enclaveStatuses := map[string]string{}
	for enclaveId, enclaveInfo := range enclaveInfoMap {
		orderedEnclaveIds = append(orderedEnclaveIds, enclaveId)
		//TODO refactor in order to print users friendly status strings and not the enum value
		enclaveStatuses[enclaveId] = enclaveInfo.GetContainersStatus().String()
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
