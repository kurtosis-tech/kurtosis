package otel

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/otel/start"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/otel/stop"
	"github.com/spf13/cobra"
)

//nolint:exhaustruct
var OtelCmd = &cobra.Command{
	Use:   command_str_consts.OtelCmdStr,
	Short: "Start Docker-only OpenTelemetry side containers",
	RunE:  nil,
}

func init() {
	OtelCmd.AddCommand(start.OtelStartCmd.MustGetCobraCommand())
	OtelCmd.AddCommand(stop.OtelStopCmd.MustGetCobraCommand())
}
