package version

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/kurtosis_cli_version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   command_str_consts.VersionCmdStr,
	Short: "Prints the CLI version",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(logrus.StandardLogger().Out, kurtosis_cli_version.KurtosisCLIVersion)
	return nil
}
