package dump

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/backend_creator"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis-cli/commons/positional_arg_parser"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

const (
	enclaveIdArg     = "enclave-id"
	outputDirpathArg = "output-dirpath"
)

var positionalArgs = []string{
	enclaveIdArg,
	outputDirpathArg,
}

var EnclaveDumpCmd = &cobra.Command{
	Use:                   command_str_consts.EnclaveDumpCmdStr + " [flags] " + strings.Join(positionalArgs, " "),
	DisableFlagsInUseLine: true,
	Short:                 "Dumps all information about the given enclave into the given output dirpath",
	RunE:                  run,
}

func init() {
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	parsedPositionalArgs, err := positional_arg_parser.ParsePositionalArgsAndRejectEmptyStrings(positionalArgs, args)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the positional args")
	}
	enclaveId := parsedPositionalArgs[enclaveIdArg]
	enclaveOutputDirpath := parsedPositionalArgs[outputDirpathArg]

	// TODO REFACTOR: we should get this backend from the config!!
	var apiContainerModeArgs *backend_creator.APIContainerModeArgs = nil  // Not an API container
	kurtosisBackend, err := backend_creator.GetLocalDockerKurtosisBackend(apiContainerModeArgs)
	if err := kurtosisBackend.DumpEnclave(ctx, enclave.EnclaveID(enclaveId), enclaveOutputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to '%v'", enclaveId, enclaveOutputDirpath)
	}

	logrus.Infof("Dumped enclave '%v' to directory '%v'", enclaveId, enclaveOutputDirpath)
	return nil
}