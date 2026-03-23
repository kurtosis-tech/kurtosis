package loki

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/loki/start"
	"github.com/spf13/cobra"
)

// LokiCmd suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var LokiCmd = &cobra.Command{
	Use:   command_str_consts.LokiCmdStr,
	Short: "Start Loki for log collection",
	RunE:  nil,
}

func init() {
	LokiCmd.AddCommand(start.LokiStartCmd.MustGetCobraCommand())
}
