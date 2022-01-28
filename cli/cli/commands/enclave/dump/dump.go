package dump

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/logrus_log_levels"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg = "enclave-id"
)

var positionalArgs = []string{
	enclaveIdArg,
}

var EnclaveDumpCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveDumpCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Lists detailed information about an enclave",
	RunE:                  run,
}

var kurtosisLogLevelStr string

func init() {
	EnclaveDumpCmd.Flags().StringVarP(
		&kurtosisLogLevelStr,
		kurtosisLogLevelArg,
		"l",
		defaultKurtosisLogLevel,
		fmt.Sprintf(
			"The log level that Kurtosis itself should log at (%v)",
			strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|"),
		),
	)

}
