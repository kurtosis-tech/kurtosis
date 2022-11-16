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
	Use:                    command_str_consts.EngineCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage the Kurtosis engine server",
	Long:                   "",
	Example:                "",
	ValidArgs:              nil,
	ValidArgsFunction:      nil,
	Args:                   nil,
	ArgAliases:             nil,
	BashCompletionFunction: "",
	Deprecated:             "",
	Annotations:            nil,
	Version:                "",
	PersistentPreRun:       nil,
	PersistentPreRunE:      nil,
	PreRun:                 nil,
	PreRunE:                nil,
	Run:                    nil,
	RunE:                   nil,
	PostRun:                nil,
	PostRunE:               nil,
	PersistentPostRun:      nil,
	PersistentPostRunE:     nil,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: false,
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd:   false,
		DisableNoDescFlag:   false,
		DisableDescriptions: false,
	},
	TraverseChildren:           false,
	Hidden:                     false,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
}

func init() {
	EngineCmd.AddCommand(start.StartCmd)
	EngineCmd.AddCommand(status.StatusCmd)
	EngineCmd.AddCommand(stop.StopCmd)
	EngineCmd.AddCommand(restart.RestartCmd)
}
