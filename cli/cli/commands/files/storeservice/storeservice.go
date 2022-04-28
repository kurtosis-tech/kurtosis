package storeservice

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/api/golang/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)
const (
	enclaveIdArgKey        = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy   = false

	serviceIdArgKey = "service-id"
	absoluteFilepathArgKey = "filepath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"
)


var FilesStoreServiceCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.FilesStoreServiceCmdStr,
	ShortDescription:          "Copy files from a user service's filepath",
	LongDescription:           fmt.Sprintf(
		"Instructs Kurtosis to copy an archive file or an entire folder from the given absolute filepath in " +
			" user service identified by its ID and store it in the enclave for later use (e.g. with '%v %v')",
		command_str_consts.ServiceCmdStr,
		command_str_consts.ServiceAddCmdStr,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		{
			Key: serviceIdArgKey,
		},
		{
			Key: absoluteFilepathArgKey,
		},
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
		return stacktrace.Propagate(err, "An error occurred getting the enclave ID using key '%v'", enclaveIdArgKey)
	}
	enclaveId := enclaves.EnclaveID(enclaveIdStr)

	serviceIdStr, err := args.GetNonGreedyArg(serviceIdArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service ID value using key '%v'", serviceIdArgKey)
	}
	serviceId := services.ServiceID(serviceIdStr)

	filepath, err := args.GetNonGreedyArg(absoluteFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the absolute filepath value using key '%v'", absoluteFilepathArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveId)
	}
	filesArtifactId, err := enclaveCtx.StoreFilesFromService(ctx, serviceId, filepath)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying content from filepath '%v' in user service with ID '%v' to enclave '%v'",
			filepath,
			serviceId,
			enclaveId,
		)
	}
	logrus.Infof("Files package ID: %v", filesArtifactId)
	return nil
}