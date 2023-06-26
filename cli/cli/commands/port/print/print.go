package print

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/service_identifier_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	serviceIdentifierArgKey        = "service"
	isServiceIdentifierArgOptional = false
	isServiceIdentifierArgGreedy   = false

	portIdentifierArgKey        = "port_id"
	isPortIdentifierArgOptional = false
	isPortIdentifierArgGreedy   = false

	ipAddress = "127.0.0.1"
)

var PortPrintCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.PortPrintCmdStr,
	ShortDescription:          "Get service logs",
	LongDescription:           "Show logs for a service inside an enclave",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     []*flags.FlagConfig{},
	Args: []*args.ArgConfig{
		enclave_id_arg.NewHistoricalEnclaveIdentifiersArgWithValidationDisabled(
			enclaveIdentifierArgKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
		// TODO use the `NewServiceIdentifierArg` instead when we start storing identifiers in DB
		// TODO we should fix this after https://github.com/kurtosis-tech/kurtosis/issues/879
		service_identifier_arg.NewHistoricalServiceIdentifierArgWithValidationDisabled(
			serviceIdentifierArgKey,
			isServiceIdentifierArgOptional,
			isServiceIdentifierArgGreedy,
		),
		{
			Key:        portIdentifierArgKey,
			IsOptional: isPortIdentifierArgOptional,
			IsGreedy:   isPortIdentifierArgGreedy,
		},
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

	portIdentifier, err := args.GetNonGreedyArg(portIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the port identifier using arg key '%v'", portIdentifier)
	}

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting to the local Kurtosis engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting enclave context for enclave with identifier '%v' exists", enclaveIdentifier)
	}

	serviceCtx, err := enclaveCtx.GetServiceContext(serviceIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while getting service context for service with identifier '%v'", serviceIdentifier)
	}

	publicPorts := serviceCtx.GetPublicPorts()

	if publicPorts[portIdentifier] == nil {
		return stacktrace.NewError(
			fmt.Sprintf("Port Identifier: %v is not found for service: %v in enclave %v", portIdentifier, serviceIdentifier, enclaveIdentifier),
		)
	}

	publicPort := publicPorts[portIdentifier]

	fullUrl := fmt.Sprintf("%v:%v", ipAddress, publicPort.GetNumber())
	maybeApplicationProtocol := publicPort.GetMaybeApplicationProtocol()

	if maybeApplicationProtocol != "" {
		fullUrl = fmt.Sprintf("%v://%v", maybeApplicationProtocol, fullUrl)
	}

	outputString := fmt.Sprintf("Here is the port information for port: %v for %v in %v",
		portIdentifier, serviceIdentifier, enclaveIdentifier)

	out.PrintOutLn(outputString)
	out.PrintOutLn(fmt.Sprintf("The url is:  %v", fullUrl))
	return nil
}
