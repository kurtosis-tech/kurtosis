package logs

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/highlevel/engine_consuming_kurtosis_command"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	outputDirpathArg = "output-dirpath"

	kurtosisBackendCtxKey = "kurtosis-backend"
	engineClientCtxKey    = "engine-client"

	defaultEngineDumpDir    = "kurtosis-engine-logs"
	outputDirIsOptional     = true
	dumpDirTimeoutDelimiter = "--"
)

var EngineLogsCmd = &engine_consuming_kurtosis_command.EngineConsumingKurtosisCommand{
	CommandStr:                command_str_consts.EngineLogsCmdStr,
	ShortDescription:          "Dumps logs for all engines",
	LongDescription:           "Dumps logs for all engines to the given directory",
	KurtosisBackendContextKey: kurtosisBackendCtxKey,
	EngineClientContextKey:    engineClientCtxKey,
	Flags:                     nil,
	Args: []*args.ArgConfig{
		// TODO Create a NewFilepathArg that has filepath tab-completion & validation set up
		{
			Key:          outputDirpathArg,
			DefaultValue: defaultEngineDumpDir,
			IsOptional:   outputDirIsOptional,
		},
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
	outputDirpath, err := args.GetNonGreedyArg(outputDirpathArg)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting output dirpath using arg key '%v'", outputDirpathArg)
	}

	if outputDirpath == defaultEngineDumpDir {
		outputDirpath = fmt.Sprintf("%s%s%d", outputDirpath, dumpDirTimeoutDelimiter, time.Now().Unix())
	}

	if err := kurtosisBackend.GetEngineLogs(ctx, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping engine logs to '%v'", outputDirpath)
	}

	logrus.Infof("Dumped engine logs and information to directory '%v'", outputDirpath)
	return nil
}
