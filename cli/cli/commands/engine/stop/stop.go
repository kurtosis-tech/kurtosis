package stop

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
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

var WithKubernetes bool

func init() {
	// TODO Remove this in favor of actual Kubernetes info in the config file
	StopCmd.Flags().BoolVarP(&WithKubernetes, "with-kubernetes", "k", false, "Operate on the engine in kubernetes")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cmd.Flags().GetBool("with-kubernetes")
	logrus.Infof("Stopping Kurtosis engine...")

	// TODO Hack; remove when we read cluster state from disk
	var clusterName = "docker"
	if WithKubernetes {
		clusterName = "minikube"
	}
	engineManager, err := engine_manager.NewEngineManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager connected to cluster '%v'", clusterName)
	}

	if err := engineManager.StopEngineIdempotently(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping the Kurtosis engine")
	}

	logrus.Info("Kurtosis engine successfully stopped")
	return nil
}
