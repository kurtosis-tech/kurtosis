package connect

import (
	"context"
	"fmt"
	"strings"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/portal_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveIdentifierArgKey = "enclave"
	isEnclaveIdArgOptional  = false
	isEnclaveIdArgGreedy    = false

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	portMappingSeparatorForLogs = ", "
)

var EnclaveConnectCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveConnectCmdStr,
	ShortDescription:          "Connects enclave",
	LongDescription:           "Connects the enclave with the given name",
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
	},
	RunFunc: run,
}

func init() {
}

func run(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient,
	_ metrics_client.MetricsClient,
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifier, err := args.GetNonGreedyArg(enclaveIdentifierArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifier arg using key '%v'", enclaveIdentifierArgKey)
	}

	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err != nil {
		logrus.Warnf("Could not retrieve the current context. Kurtosis will assume context is local.")
		logrus.Debugf("Error was: %v", err.Error())
		return nil
	}

	if !store.IsRemote(currentContext) {
		logrus.Info("Current context is local, not mapping enclave service ports")
		return nil
	}

	logrus.Info("Connecting enclave...")
	if err = connectAllEnclaveServices(ctx, enclaveIdentifier); err != nil {
		return stacktrace.Propagate(err, "An error occurred connecting all enclave services")
	}

	logrus.Info("Enclave connected successfully")

	return nil
}

func connectAllEnclaveServices(ctx context.Context, enclaveIdentifier string) error {

	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating Kurtosis Context from local engine")
	}

	enclaveCtx, err := kurtosisCtx.GetEnclaveContext(ctx, enclaveIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting an enclave context from enclave info for enclave '%v'", enclaveIdentifier)
	}
	allServicesMap := map[string]bool{}
	serviceContexts, err := enclaveCtx.GetServiceContexts(allServicesMap)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving the service contexts")
	}

	portsMapping := map[uint16]*services.PortSpec{}
	for _, serviceCtx := range serviceContexts {
		for _, portSpec := range serviceCtx.GetPublicPorts() {
			portsMapping[portSpec.GetNumber()] = portSpec
		}
	}

	portalManager := portal_manager.NewPortalManager()
	successfullyMappedPorts, failedPorts, err := portalManager.MapPorts(ctx, portsMapping)
	if err != nil {
		var stringifiedPortMapping []string
		for localPort, remotePort := range failedPorts {
			stringifiedPortMapping = append(stringifiedPortMapping, fmt.Sprintf("%d:%d", localPort, remotePort.GetNumber()))
		}
		logrus.Warnf("The enclave was successfully run but the following port(s) could not be mapped locally: %s. "+
			"The associated service(s) will not be reachable on the local host",
			strings.Join(stringifiedPortMapping, portMappingSeparatorForLogs))
		logrus.Debugf("Error was: %v", err.Error())
		return nil
	}
	logrus.Infof("Successfully mapped %d ports. All services running inside the enclave are reachable locally on"+
		" their ephemeral port numbers", len(successfullyMappedPorts))
	return nil
}
