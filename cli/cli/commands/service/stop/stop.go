package stop

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
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/dzobbe/PoTE-kurtosis/cli/cli/helpers/shared_starlark_calls"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/dzobbe/PoTE-kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
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
)

var ServiceStopCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.ServiceStopCmdStr,
	ShortDescription:          "Stops a service",
	LongDescription:           "Stops temporarily a service with the given service identifier in the given enclave",
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
		logrus.Infof("Stopping service '%v'", serviceIdentifier)
		serviceContext, err := enclaveCtx.GetServiceContext(serviceIdentifier)
		if err != nil {
			return stacktrace.NewError("Couldn't validate whether the service exists for identifier '%v'", serviceIdentifier)
		}

		serviceName := serviceContext.GetServiceName()

		err = metricsClient.TrackStopService(enclaveIdentifier, string(serviceName))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred tracking service update metric.")
		}

		if err := shared_starlark_calls.StopServiceStarlarkCommand(ctx, enclaveCtx, serviceName); err != nil {
			return stacktrace.Propagate(err, "An error occurred stopping service '%v' from enclave '%v'", serviceIdentifier, enclaveIdentifier)
		}
	}
	return nil
}
