package engine

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/start"
	"github.com/spf13/cobra"
)

var EngineCmd = &cobra.Command{
	Use:   "engine",
	Short: "Manage the Kurtosis engine server",
	RunE:  nil,
}

func init() {
	EngineCmd.AddCommand(start.StartCmd)
}
