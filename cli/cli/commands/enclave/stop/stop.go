package stop

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/enclave_id_arg"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	enclaveIdArg     = "enclave-id"
	isEnclaveIdArgOptional = false
	isEnclaveIdArgGreedy = true // The user can specify multiple enclaves to stop

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey  = "engine-client"
)

var EnclaveStopCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EnclaveStopCmdStr,
	ShortDescription:          "Stops enclaves",
	LongDescription:           "Stops the enclaves with the given IDs",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Args:                      []*args.ArgConfig{
		enclave_id_arg.NewEnclaveIDArg(
			enclaveIdArg,
			engineClientCtxKey,
			isEnclaveIdArgOptional,
			isEnclaveIdArgGreedy,
		),
	},
	RunFunc:                   run,
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
	enclaveIds, err := args.GetGreedyArg(enclaveIdArg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the enclave IDs arg using key '%v'", enclaveIdArg)
	}

	logrus.Info("Stopping enclaves...")
	stopEnclaveErrorStrs := []string{}
	for _, enclaveId := range enclaveIds {
		stopArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{EnclaveId: enclaveId}
		if _, err := engineClient.StopEnclave(ctx, stopArgs); err != nil {
			wrappedErr := stacktrace.Propagate(err, "An error occurred stopping enclave '%v'", enclaveId)
			stopEnclaveErrorStrs = append(stopEnclaveErrorStrs, wrappedErr.Error())
		}
	}

	if len(stopEnclaveErrorStrs) > 0 {
		joinedErrorsStr := strings.Join(
			stopEnclaveErrorStrs,
			"\n\n",
		)
		// We use this rather than stacktrace because stacktrace gets messy
		return errors.New(
			fmt.Sprintf(
				"One or more errors occurred when stopping enclaves:\n%v",
				joinedErrorsStr,
			),
		)
	}

	logrus.Info("Enclaves stopped successfully")

	return nil
}
