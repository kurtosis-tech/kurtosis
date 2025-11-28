package version

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/spf13/cobra"
)

const (
	cliVersionKey = "CLI Version"
)

// VersionCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var VersionCmd = &cobra.Command{
	Use:   command_str_consts.VersionCmdStr,
	Short: "Prints the CLI version",
	Long: fmt.Sprintf(
		"Prints the version of the Kurtosis CLI (use '%v %v %v' to print the Kurtosis engine version)",
		command_str_consts.KurtosisCmdStr,
		command_str_consts.EngineCmdStr,
		command_str_consts.EngineStatusCmdStr,
	),
	RunE: run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(cliVersionKey, kurtosis_version.KurtosisVersion)
	keyValuePrinter.Print()

	fmt.Println()
	fmt.Printf(
		"To see the engine version (provided it is running): %v %v %v\n",
		command_str_consts.KurtosisCmdStr,
		command_str_consts.EngineCmdStr,
		command_str_consts.EngineStatusCmdStr,
	)

	return nil
}
