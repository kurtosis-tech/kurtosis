/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package shell

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/command_str_consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey  = "service"
	isServiceGuidArgOptional = false
	isServiceGuidArgGreedy   = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var ServiceShellCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceShellCmdStr,
	ShortDescription:          "Gets a shell on a service",
	LongDescription:           "Starts a shell on the specified service",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		service_identifier_arg.NewServiceIdentifierArg(
			serviceIdentifierArgKey,
			enclaveIdentifierArgKey,
			isServiceGuidArgOptional,
			isServiceGuidArgGreedy,
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

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier using arg key '%v'", serviceIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting enclave context for enclave with identifier '%v' exists", enclaveIdentifier)
	}

	enclaveUuid := enclave.EnclaveUUID(enclaveCtx.GetEnclaveUuid())

	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting service context for service with identifier '%v'", serviceIdentifier)
	}
	serviceUuid := service.ServiceUUID(serviceCtx.GetServiceUUID())

	if err = kurtosisBackend.GetShellOnUserService(ctx, enclaveUuid, serviceUuid); err != nil {
		return stacktrace.Propagate(err, "An error occurred getting shell on user service with UUID '%v' in enclave '%v'", serviceUuid, enclaveIdentifier)
	}

	return nil
}
