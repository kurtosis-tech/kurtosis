package files

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/enclave/upload"
	"github.com/spf13/cobra"
)

var FilesCmd = &cobra.Command{
	Use:   command_str_consts.FilesCmdStr,
	Short: "Manage files for an enclave",
	Long: "Contains actions for managing the Kurtosis enclave filestore, used for sending around, in, and out of the enclave",
	RunE:  nil,
}

func init() {
	FilesCmd.AddCommand(upload.FilesUploadCmd.MustGetCobraCommand())
}
