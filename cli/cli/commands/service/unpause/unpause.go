package unpause

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
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
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveIdStr := parsedPositionalArgs[enclaveIdArg]
	enclaveId := enclave.EnclaveID(enclaveIdStr)
	guidStr := parsedPositionalArgs[guidArg]
	guid := service.ServiceGUID(guidStr)

	kurtosisBackend, err := lib.GetLocalDockerKurtosisBackend()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting local Docker Kurtosis backend")
	}

	err = kurtosisBackend.UnpauseService(ctx, enclaveId, guid)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred trying to unpause service '%v' in enclave '%v'", guid, enclaveId)
	}

	return nil
}
