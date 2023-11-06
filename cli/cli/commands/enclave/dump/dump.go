package dump

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/file_system_path_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	outputDirpathArg = "output-dirpath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	defaultEnclaveDumpDir = "kurtosis-dump"
	enclaveDumpSeparator  = "--"
	outputDirIsOptional   = true
)

var EnclaveDumpCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveDumpCmdStr,
	ShortDescription:          "Dumps information about an enclave to disk",
	LongDescription:           "Dumps all information about the enclave to the given directory",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     nil,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifierArgKey,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		file_system_path_arg.NewDirpathArg(
			outputDirpathArg,
			outputDirIsOptional,
			defaultEnclaveDumpDir,
			file_system_path_arg.BypassDefaultValidationFunc,
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
		return stacktrace.Propagate(err, "An error occurred getting enclave identifier using arg key '%v'", enclaveIdentifierArgKey)
	}
	enclaveOutputDirpath, err := args.GetNonGreedyArg(outputDirpathArg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting output dirpath using arg key '%v'", outputDirpathArg)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting enclave for identifier '%v'", enclaveIdentifier)
	}

	enclaveUuid := enclaveInfo.GetEnclaveUuid()

	if enclaveOutputDirpath == defaultEnclaveDumpDir {
		enclaveName := enclaveInfo.GetName()
		enclaveOutputDirpath = fmt.Sprintf("%s%s%s", enclaveName, enclaveDumpSeparator, enclaveUuid)
	}

	if err := kurtosisBackend.DumpEnclave(ctx, enclave.EnclaveUUID(enclaveUuid), enclaveOutputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to '%v'", enclaveIdentifier, enclaveOutputDirpath)
	}

	logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveIdentifier, enclaveOutputDirpath)
	return nil
}
