package unpause

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg = "enclave-id"
	guidArg      = "guid"
)

var positionalArgs = []string{
	enclaveIdArg,
	guidArg,
}

var UnpauseCmd = &cobra.Command{
	Use:                   command_str_consts.ServiceUnpauseCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Unpause all processes running in a service inside of an enclave",
	RunE:                  run,
}

func run(cmd *cobra.Command, args []string) error {
	_ = context.Background()
	return nil
}
