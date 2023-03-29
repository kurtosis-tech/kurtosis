package version

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/engine_manager"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/output_printers"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	cliVersionKey                       = "CLI Version"
	runningEngineVersionKey             = "Running Engine Version"
	errorDeterminingEngineVersionLogStr = "Ran into an error determining running engine version. Use `kurtosis engine status` to learn more"
)

// VersionCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var VersionCmd = &cobra.Command{
	Use:   command_str_consts.VersionCmdStr,
	Short: "Prints the CLI and Running Engine Version",
	Long:  "Prints the version of the CLI and if there is any running Engine then it prints that too",
	RunE:  run,
}

func init() {
	// No flags yet
}

func run(cmd *cobra.Command, args []string) error {
	keyValuePrinter := output_printers.NewKeyValuePrinter()
	keyValuePrinter.AddPair(cliVersionKey, kurtosis_version.KurtosisVersion)

	ctx := context.Background()

	engineManager, err := engine_manager.NewEngineManager(ctx)
	if err != nil {
		// if the engine manager can't be fetched; perhaps docker isn't alive we just print the CLI version
		logrus.Errorf(errorDeterminingEngineVersionLogStr)
		keyValuePrinter.Print()
		return nil
	}

	status, _, maybeEngineVersion, err := engineManager.GetEngineStatus(ctx)
	if err != nil {
		// if the engine status can't be fetched we just print the CLI version
		logrus.Errorf(errorDeterminingEngineVersionLogStr)
		keyValuePrinter.Print()
		return nil
	}
	if status == engine_manager.EngineStatus_Running {
		keyValuePrinter.AddPair(runningEngineVersionKey, maybeEngineVersion)
	}

	keyValuePrinter.Print()

	return nil
}
