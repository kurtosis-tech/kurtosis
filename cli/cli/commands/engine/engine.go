package engine

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/start"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/status"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/stop"
	"github.com/spf13/cobra"
)

const (
	CommandStr = "engine"
)

var EngineCmd = &cobra.Command{
	Use:   CommandStr,
	Short: "Manage the Kurtosis engine server",
	RunE:  nil,
}

func init() {
	EngineCmd.AddCommand(start.StartCmd)
	EngineCmd.AddCommand(status.StatusCmd)
	EngineCmd.AddCommand(stop.StopCmd)
}
