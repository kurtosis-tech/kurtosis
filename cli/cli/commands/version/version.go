package version

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// VersionCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var VersionCmd = &cobra.Command{
	Use:   command_str_consts.VersionCmdStr,
	Short: "Prints the CLI version",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(logrus.StandardLogger().Out, kurtosis_version.KurtosisVersion)
	return nil
}
