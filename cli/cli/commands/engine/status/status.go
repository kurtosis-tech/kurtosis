package status

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/backend_for_cmd"
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
	// No flags yet
	StatusCmd.Flags().BoolVarP(&WithKubernetes, "with-kubernetes", "k", false, "Operate on the engine in kubernetes")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	kurtosisBackend, err := backend_for_cmd.GetBackendForCmd(WithKubernetes)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting a Kurtosis backend")
	}
	engineManager := engine_manager.NewEngineManager(kurtosisBackend)
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
