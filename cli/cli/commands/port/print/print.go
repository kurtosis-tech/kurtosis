package print

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
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
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
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

	formatFlagKey = "format"
	protocolStr   = "protocol"
	ipStr         = "ip"
	numberStr     = "number"

	ipAddress = "127.0.0.1"
)

var expectedRelativeOrder = map[string]int{
	protocolStr: 0,
	ipStr:       1,
	numberStr:   2, //nolint:gomnd
}

var (
	formatFlagKeyDefault  = fmt.Sprintf("%s,%s,%s", protocolStr, ipStr, numberStr)
	formatFlagKeyExamples = []string{
		fmt.Sprintf("'%s'", ipStr),
		fmt.Sprintf("'%s'", numberStr),
		fmt.Sprintf("'%s,%s'", protocolStr, ipStr),
	}
)

var PortPrintCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.PortPrintCmdStr,
	ShortDescription:          "Get information about port",
	LongDescription:           "Get information for port using port id",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags: []*flags.FlagConfig{
		{
			Key: formatFlagKey,
			Usage: fmt.Sprintf(
				"Allows selecting what pieces of port are printed, using comma separated values (examples: %s). Default '%s'.",
				strings.Join(formatFlagKeyExamples, ", "), formatFlagKeyDefault),
			Type:    flags.FlagType_String,
			Default: formatFlagKeyDefault,
		},
	},
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
			enclaveIdentifierArgKey,
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

	format, err := flags.GetString(formatFlagKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the output flag key '%v'", formatFlagKey)
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

	publicPort, found := publicPorts[portIdentifier]
	if !found {
		return stacktrace.NewError(
			fmt.Sprintf("Port Identifier: '%v' is not found for service: '%v' in enclave '%v'", portIdentifier, serviceIdentifier, enclaveIdentifier),
		)
	}

	fullUrl, err := formatPortOutput(format, ipAddress, publicPort)
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't format the output according to formatting string '%v'", format)
	}
	out.PrintOutLn(fullUrl)
	return nil
}

func formatPortOutput(format string, ipAddress string, spec *services.PortSpec) (string, error) {
	parts := strings.Split(format, ",")
	var resultParts []string
	isOnlyPiece := len(parts) == 1
	lastPart := parts[0]
	for _, part := range parts {
		if partIndex := expectedRelativeOrder[part]; partIndex < expectedRelativeOrder[lastPart] {
			return "", stacktrace.NewError("Found '%v' before '%v', which is not expected.", lastPart, part)
		}
	}
	for _, part := range parts {
		switch part {
		case protocolStr:
			if spec.GetMaybeApplicationProtocol() != "" {
				resultParts = append(resultParts, getApplicationProtocol(spec.GetMaybeApplicationProtocol(), isOnlyPiece))
			} else {
				// TODO(victor.colombo): What should we do here? Panic? Warn?
				// Left it as a debug for now so it doesn't pollute the output
				logrus.Debugf("Expected protocol but was empty, skipping")
			}
		case ipStr:
			resultParts = append(resultParts, ipAddress)
		case numberStr:
			resultParts = append(resultParts, getPortString(spec.GetNumber(), isOnlyPiece))
		default:
			return "", stacktrace.NewError("Invalid format piece '%v'", part)
		}
	}
	return strings.Join(resultParts, ""), nil
}

func getPortString(portNumber uint16, isOnlyPiece bool) string {
	if isOnlyPiece {
		return fmt.Sprintf("%d", portNumber)
	} else {
		return fmt.Sprintf(":%d", portNumber)
	}
}

func getApplicationProtocol(applicationProtocol string, isOnlyPiece bool) string {
	if isOnlyPiece {
		return applicationProtocol
	} else {
		return fmt.Sprintf("%s://", applicationProtocol)
	}
}
