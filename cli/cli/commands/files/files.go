package files

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/rendertemplate"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/storeservice"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/storeweb"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/upload"
	"github.com/spf13/cobra"
)

var FilesCmd = &cobra.Command{
	Use:                    command_str_consts.FilesCmdStr,
	Aliases:                nil,
	SuggestFor:             nil,
	Short:                  "Manage files for an enclave",
	Long:                   "Contains actions for managing the Kurtosis enclave filestore, used for sending around, in, and out of the enclave",
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
	FilesCmd.AddCommand(upload.FilesUploadCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(storeweb.FilesStoreWebCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(storeservice.FilesStoreServiceCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(rendertemplate.RenderTemplateCommand.MustGetCobraCommand())
}
