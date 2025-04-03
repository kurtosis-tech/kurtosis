package grafloki

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/grafloki/start"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/grafloki/stop"
	"github.com/spf13/cobra"
)

// GraflokiCmd suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var GraflokiCmd = &cobra.Command{
	Use:   command_str_consts.GraflokiCmdStr,
	Short: "Start Grafana/Loki command for log collection",
	RunE:  nil,
}

func init() {
	GraflokiCmd.AddCommand(start.GraflokiStartCmd.MustGetCobraCommand())
	GraflokiCmd.AddCommand(stop.GraflokiStopCmd.MustGetCobraCommand())
}
