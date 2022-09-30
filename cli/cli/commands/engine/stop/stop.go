package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   command_str_consts.EngineStopCmdStr,
	Short: "Stops the Kurtosis engine",
	Long:  "Stops the Kurtosis engine, doing nothing if no engine is running",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	logrus.Infof("Stopping Kurtosis engine...")

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager")
	}

	if err := engineManager.StopEngineIdempotently(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the Kurtosis engine")
	}

	logrus.Info("Kurtosis engine successfully stopped")
	return nil
}
