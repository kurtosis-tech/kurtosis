package files

import (
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/download"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/inspect"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/rendertemplate"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/storeservice"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/storeweb"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files/upload"
	"github.com/spf13/cobra"
)

// FilesCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var FilesCmd = &cobra.Command{
	Use:   command_str_consts.FilesCmdStr,
	Short: "Manage files for an enclave",
	Long:  "Contains actions for managing the Kurtosis enclave filestore, used for sending around, in, and out of the enclave",
	RunE:  nil,
}

func init() {
	FilesCmd.AddCommand(upload.FilesUploadCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(storeweb.FilesStoreWebCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(storeservice.FilesStoreServiceCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(rendertemplate.RenderTemplateCommand.MustGetCobraCommand())
	FilesCmd.AddCommand(inspect.FilesInspectCmd.MustGetCobraCommand())
	FilesCmd.AddCommand(download.FilesDownloadCmd.MustGetCobraCommand())
}
