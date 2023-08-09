/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package identifiers

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var ServiceIdentifiersCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                "identifiers",
	ShortDescription:          "Get service identifiers",
	LongDescription:           "Show service identifiers",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		//TODO disabling enclaveID validation and serviceUUID validation for allowing consuming logs from removed or stopped enclaves
		//TODO we should enable them when #879 is ready: https://github.com/kurtosis-tech/kurtosis/issues/879
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
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using arg key '%v'", enclaveIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	// we need to get the enclave context to get the historical service identifiers for the enclave
	// we also need it to get the `ServiceContext`
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		// we couldn't get the enclaveCtx we go with service id passed as the serviceUuid
		// for the enclave either we got it from the historical values or we return what was passed to the function
		return stacktrace.Propagate(err, "An error occurred getting enclave context")
	}

	// we see if we can get historical identifiers
	serviceIdentifiers, err := enclaveCtx.GetExistingAndHistoricalServiceIdentifiers(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service identifiers")
	}

	tablePrinter := output_printers.NewTablePrinter("Service Name", "Service UUID")

	for serviceName, serviceUuids := range serviceIdentifiers.GetServiceNameToUuids() {
		for _, serviceUuid := range serviceUuids {
			if err := tablePrinter.AddRow(string(serviceName), string(serviceUuid)); err != nil {
				return stacktrace.NewError("An error occurred adding row for enclave '%v' to the table printer", enclaveIdentifier)
			}
		}
	}

	tablePrinter.Print()

	return nil
}
