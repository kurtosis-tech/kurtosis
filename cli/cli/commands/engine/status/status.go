package status

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/spf13/cobra"
)

const ()

var StatusCmd = &cobra.Command{
	Use:   command_str_consts.EngineStatusCmdStr,
	Short: "Reports the status of the Kurtosis engine",
	RunE:  run,
}

var WithKubernetes bool

func init() {
	// TODO Remove this in favor of actual Kubernetes info in the config file
	StatusCmd.Flags().BoolVarP(&WithKubernetes, "with-kubernetes", "k", false, "Operate on the engine in kubernetes")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// TODO Hack; remove when we read cluster state from disk
	var clusterName = "docker"
	if WithKubernetes {
		clusterName = "minikube"
	}
	engineManager, err := engine_manager.NewEngineManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating an engine manager connected to cluster '%v'", clusterName)
	}

	status, _, maybeApiVersion, err := engineManager.GetEngineStatus(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Kurtosis engine status")
	}
	prettyPrintingStatusVisitor := newPrettyPrintingEngineStatusVisitor(maybeApiVersion)
	if err := status.Accept(prettyPrintingStatusVisitor); err != nil {
		return stacktrace.Propagate(err, "An error occurred printing the engine status")
	}

	return nil
}
