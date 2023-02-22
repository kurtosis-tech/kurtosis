package stop

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/metrics_client_factory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	enclaveIdentifiersArgKey = "enclave-identifiers"
	isEnclaveIdArgOptional   = false
	isEnclaveIdArgGreedy     = true // The user can specify multiple enclaves to stop

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"
)

var EnclaveStopCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveStopCmdStr,
	ShortDescription:          "Stops enclaves",
	LongDescription:           "Stops the enclaves with the given UUIDs",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     nil,
	Args: []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIdentifierArg(
			enclaveIdentifiersArgKey,
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
	_ *flags.ParsedFlags,
	args *args.ParsedArgs,
) error {
	enclaveIdentifiers, err := args.GetGreedyArg(enclaveIdentifiersArgKey)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave identifiers arg using key '%v'", enclaveIdentifiersArgKey)
	}

	metricsClient, metricsClientCloser, err := metrics_client_factory.GetMetricsClient()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating metrics client, the metrics will not be recorded")
	}
	defer func() {
		err = metricsClientCloser()
		if err != nil {
			logrus.Warnf("An error occurred while closing the metrics client\n%s", err)
		}
	}()

	logrus.Info("Stopping enclaves...")
	stopEnclaveErrorStrs := []string{}
	for _, enclaveIdentifier := range enclaveIdentifiers {
		stopArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{EnclaveIdentifier: enclaveIdentifier}
		if err = metricsClient.TrackStopEnclave(enclaveIdentifier); err != nil {
			logrus.Errorf("An error occurred while logging the stop enclave event for enclave '%v'", enclaveIdentifier)
		}
		if _, err := engineClient.StopEnclave(ctx, stopArgs); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveIdentifier)
			stopEnclaveErrorStrs = append(stopEnclaveErrorStrs, wrappedErr.Error())
		}
	}

	if len(stopEnclaveErrorStrs) > 0 {
		joinedErrorsStr := strings.Join(
			stopEnclaveErrorStrs,
			"\n\n",
		)
		// We use this rather than stacktrace because stacktrace gets messy
		return fmt.Errorf(
			"One or more errors occurred when stopping enclaves:\n%v",
			joinedErrorsStr,
		)
	}

	logrus.Info("Enclaves stopped successfully")

	return nil
}
