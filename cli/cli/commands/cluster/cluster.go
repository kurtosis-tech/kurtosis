package cluster

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	get_cluster "github.com/kurtosis-tech/kurtosis/cli/cli/commands/cluster/get"
	ls_cluster "github.com/kurtosis-tech/kurtosis/cli/cli/commands/cluster/ls"
	set_cluster "github.com/kurtosis-tech/kurtosis/cli/cli/commands/cluster/set"
	"github.com/spf13/cobra"
)

var ClusterCmd = &cobra.Command{
	Use:                    command_str_consts.ClusterCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage Kurtosis cluster setting",
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
	ClusterCmd.AddCommand(set_cluster.SetCmd.MustGetCobraCommand())
	ClusterCmd.AddCommand(ls_cluster.LsCmd.MustGetCobraCommand())
	ClusterCmd.AddCommand(get_cluster.GetCmd.MustGetCobraCommand())
}
