package engine

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/restart"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/start"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/status"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine/stop"
	"github.com/spf13/cobra"
)

var EngineCmd = &cobra.Command{
	Use:   command_str_consts.EngineCmdStr,
	Short: "Manage the Kurtosis engine server",
	RunE:  nil,
}

func init() {
	EngineCmd.AddCommand(start.StartCmd)
	EngineCmd.AddCommand(status.StatusCmd)
	EngineCmd.AddCommand(stop.StopCmd)
	EngineCmd.AddCommand(restart.RestartCmd)
}
