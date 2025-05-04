/*
 * Copyright (c) 2024 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package snapshot

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/shared_starlark_calls"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey        = "enclave-identifier"
	enclaveIdentifierArgIsOptional = false
	enclaveIdentifierArgIsGreedy   = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var EnclaveSnapshotCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveSnapshotCmdStr,
	ShortDescription:          "Takes a snapshot of an enclave",
	LongDescription:           "Takes a snapshot of the specified enclave, capturing its current state",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			false,
			false,
		),
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "Expected a value for enclave identifier but none was found")
	}
	if enclaveIdentifier == "" {
		return stacktrace.NewError("Enclave identifier cannot be empty")
	}

	err = stopAllEnclaveServices(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping all services in enclave '%v'", enclaveIdentifier)
	}

	// err = snapshotEnclave(ctx, enclaveIdentifier, kurtosisBackend, dockerManager)
	// if err != nil {
	// 	return stacktrace.Propagate(err, "An error occurred snapshotting enclave '%v'", enclaveIdentifier)
	// }

	// 3. Handle response and output

	// get services in enclave

	// kurtosis backend.snapshot enclave (enclave identifier)?

	return nil
}

// func snapshotEnclave(ctx context.Context, enclaveIdentifier string, kurtosisBackend backend_interface.KurtosisBackend, dockerManager *docker_manager.DockerManager) error {
// 	enclaveUuid := enclave.EnclaveUUID(enclaveIdentifier)

// 	err := stopAllEnclaveServices(ctx, enclaveIdentifier)

// 	// Stop enclave
// 	_, _, err = kurtosisBackend.StopEnclaves(ctx, &enclave.EnclaveFilters{
// 		UUIDs: map[enclave.EnclaveUUID]bool{
// 			enclaveUuid: true,
// 		},
// 	})
// 	if err != nil {
// 		return stacktrace.Propagate(err, "An error occurred stopping enclave '%v' while snapshotting", enclaveUuid)
// 	}
// 	logrus.Infof("Enclave '%v' stopped", enclaveUuid)

// 	// Commit all the services in the enclave
// 	logrus.Infof("Getting services in enclave '%v'", enclaveUuid)
// 	services, err := kurtosisBackend.GetUserServices(ctx, enclaveUuid, &service.ServiceFilters{
// 		Names:    nil,
// 		UUIDs:    map[service.ServiceUUID]bool{},
// 		Statuses: nil,
// 	})
// 	if err != nil {
// 		return stacktrace.Propagate(err, "An error occurred getting services in enclave '%v' while snapshotting", enclaveUuid)
// 	}
// 	for _, service := range services {
// 		logrus.Infof("Commits service '%v'", service.GetRegistration().GetName())
// 		err = dockerManager.CommitContainer(ctx, string(service.GetRegistration().GetName()))
// 		if err != nil {
// 			return stacktrace.Propagate(err, "An error occurred committing service '%v' while snapshotting", service.GetRegistration().GetName())
// 		}
// 	}

// 	// save the images in tar files

// 	// create a tar file containing the tars of all the images

// 	// save this tar file to output dir path

// 	return nil
// }

func stopAllEnclaveServices(ctx context.Context, enclaveIdentifier string) error {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	allEnclaveServices, err := enclaveCtx.GetServices()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting all enclave services")
	}

	for serviceName := range allEnclaveServices {
		if err := shared_starlark_calls.StopServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping service '%s'", serviceName)
		}
	}
	return nil
}
