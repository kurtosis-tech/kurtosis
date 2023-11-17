package rm

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	starlarkScript = `
def run(plan, args):
	plan.remove_service(name=args["service_name"])
`
)

var ServiceRmCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceRmCmdStr,
	ShortDescription:          "Removes a service from an enclave",
	LongDescription:           "Removes the service with the given identifier from the given enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
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
			isServiceIdentifierArgGreedy,
			isServiceIdentifierArgOptional,
		),
	},
	Flags:   []*flags.FlagConfig{},
	RunFunc: run,
}

func run(
	ctx context.Context,
	_ backend_interface.KurtosisBackend,
	_ kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier value using key '%v'", enclaveIdentifierArgKey)
	}

	serviceIdentifier, err := args.GetNonGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service ID value using key '%v'", serviceIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	serviceContext, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err != nil {
		return stacktrace.NewError("Couldn't validate whether the service exists for identifier '%v'", serviceIdentifier)
	}

	serviceName := serviceContext.GetServiceName()

	if err := removeServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
		return stacktrace.Propagate(err, "An error occurred removing service '%v' from enclave '%v'", serviceIdentifier, enclaveIdentifier)
	}
	return nil
}

func removeServiceStarlarkCommand(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName) error {
	params := fmt.Sprintf(`{"service_name": "%s"}`, serviceName)
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(params)))
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred on Starlark for rendering template")
	}
	if runResult.ExecutionError != nil {
		return stacktrace.NewError("An error occurred during Starlark script execution for rendering template: %s", runResult.ExecutionError.GetErrorMessage())
	}
	if runResult.InterpretationError != nil {
		return stacktrace.NewError("An error occurred during Starlark script interpretation for rendering template: %s", runResult.InterpretationError.GetErrorMessage())
	}
	if len(runResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error occurred during Starlark script validation for rendering template: %v", runResult.ValidationErrors)
	}
	return nil
}
