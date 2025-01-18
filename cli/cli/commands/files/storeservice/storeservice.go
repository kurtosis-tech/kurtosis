package storeservice

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false
	absoluteFilepathArgKey         = "filepath"

	nameFlagKey = "name"
	defaultName = ""

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	starlarkTemplateWithArtifactName = `
def run(plan, args):
	plan.store_service_files(
		src = args["src"],
		name = args["name"],
		service_name = args["service_name"],
	)
`

	starlarkTemplateWithoutArtifactName = `
def run(plan, args):
	plan.store_service_files(
		src = args["src"],
		service_name = args["service_name"],
	)
`
)

var FilesStoreServiceCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:       command_str_consts.FilesStoreServiceCmdStr,
	ShortDescription: "Store files from a service",
	LongDescription: fmt.Sprintf(
		"Instructs Kurtosis to copy a file or folder from the given absolute filepath in "+
			"the given service and store it in the enclave for later use (e.g. with '%v %v')",
		command_str_consts.ServiceCmdStr,
		command_str_consts.ServiceAddCmdStr,
	),
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key:     nameFlagKey,
			Usage:   "The name to be given to the produced of the artifact, auto generated if not passed",
			Type:    flags.FlagType_String,
			Default: defaultName,
		},
	},
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
			isServiceIdentifierArgOptional,
			isServiceIdentifierArgGreedy,
		),
		{
			Key: absoluteFilepathArgKey,
		},
	},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	flags *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier using key '%v'", enclaveIdentifierArgKey)
	}

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier value using key '%v'", serviceIdentifierArgKey)
	}

	filepath, err := args.GetNonGreedyArg(absoluteFilepathArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the absolute filepath value using key '%v'", absoluteFilepathArgKey)
	}

	artifactName, err := flags.GetString(nameFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the name to be given to the produced artifact")
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}
	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave context for enclave '%v'", enclaveIdentifier)
	}
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service context for service with identifier '%v'", serviceIdentifier)
	}
	serviceName := serviceCtx.GetServiceName()

	runResult, err := storeServiceFileStarlarkCommand(ctx, enclaveCtx, serviceName, filepath, enclaveIdentifier, artifactName)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying content from filepath '%v' in user service with name '%v' to enclave '%v'",
			filepath,
			serviceName,
			enclaveIdentifier,
		)
	}
	logrus.Info(runResult.RunOutput)
	return nil
}

func storeServiceFileStarlarkCommand(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName, filePath string, enclaveIdentifier string, artifactName string) (*enclaves.StarlarkRunResult, error) {
	template := starlarkTemplateWithArtifactName
	if artifactName == defaultName {
		template = starlarkTemplateWithoutArtifactName
	}
	params := fmt.Sprintf(`{"service_name": "%s", "src": "%s", "name": "%s"}`, serviceName, filePath, artifactName)
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, template, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(params)))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An unexpected error occurred running command for copying content from filepath '%v' in user service with Name '%v' to enclave '%v'. This is a bug in Kurtosis, please report.",
			filePath,
			serviceName,
			enclaveIdentifier)
	}
	if runResult.InterpretationError != nil {
		return nil, stacktrace.NewError(
			"An error occurred interpreting command for copying content from filepath '%v' in user service with name '%v' to enclave '%v': %s\nThis is a bug in Kurtosis, please report.",
			filePath,
			serviceName,
			enclaveIdentifier,
			runResult.InterpretationError.GetErrorMessage(),
		)
	}
	if len(runResult.ValidationErrors) > 0 {
		return nil, stacktrace.NewError(
			"An error occurred validating command for copying content from filepath '%v' in user service with name '%v' to enclave '%v': %v",
			filePath,
			serviceName,
			enclaveIdentifier,
			runResult.ValidationErrors,
		)
	}
	if runResult.ExecutionError != nil {
		return nil, stacktrace.NewError(
			"An error occurred executing command for copying content from filepath '%v' in user service with name '%v' to enclave '%v': %s",
			filePath,
			serviceName,
			enclaveIdentifier,
			runResult.ExecutionError.GetErrorMessage(),
		)
	}
	return runResult, err
}
