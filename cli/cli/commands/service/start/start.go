package start

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/sirupsen/logrus"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
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
	isServiceIdentifierArgGreedy   = true

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	starlarkScript = `
def run(plan, args):
	plan.start_service(name=args["service_name"])
`
)

var ServiceStartCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceStartCmdStr,
	ShortDescription:          "Restarts a stopped service",
	LongDescription:           "Restarts the temporarily stopped service with the given service identifier in the given enclave",
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
			isServiceIdentifierArgOptional,
			isServiceIdentifierArgGreedy,
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

	serviceIdentifiers, err := args.GetGreedyArg(serviceIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the service identifier value using key '%v'", serviceIdentifierArgKey)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}

	metricsClient, closeMetricsClientFunc, err := metrics_client_factory.GetMetricsClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting metrics client.")
	}
	defer func() {
		if err = closeMetricsClientFunc(); err != nil {
			logrus.Warnf("An error occurred closing metrics client:\n%v", closeMetricsClientFunc())
		}
	}()

	for _, serviceIdentifier := range serviceIdentifiers {
		logrus.Infof("Starting service '%v'", serviceIdentifier)
		serviceContext, err := enclaveCtx.GetServiceContext(serviceIdentifier)
		if err != nil {
			return stacktrace.NewError("Couldn't validate whether the service exists for identifier '%v'", serviceIdentifier)
		}

		serviceName := serviceContext.GetServiceName()

		err = metricsClient.TrackStartService(enclaveIdentifier, string(serviceName))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred tracking service update metric.")
		}

		if err := startServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
			return stacktrace.Propagate(err, "An error occurred starting service '%v' from enclave '%v'", serviceIdentifier, enclaveIdentifier)
		}

	}
	return nil
}

func startServiceStarlarkCommand(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName) error {
	serviceNameString := fmt.Sprintf(`{"service_name": "%s"}`, serviceName)
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(serviceNameString)))
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred on Starlark for starting service")
	}
	if runResult.ExecutionError != nil {
		return stacktrace.NewError("An error occurred during Starlark script execution for starting service: %s", runResult.ExecutionError.GetErrorMessage())
	}
	if runResult.InterpretationError != nil {
		return stacktrace.NewError("An error occurred during Starlark script interpretation for starting service: %s", runResult.InterpretationError.GetErrorMessage())
	}
	if len(runResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error occurred during Starlark script validation for starting service: %v", runResult.ValidationErrors)
	}
	return nil
}
