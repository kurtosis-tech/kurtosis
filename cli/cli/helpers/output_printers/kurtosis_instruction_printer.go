package output_printers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
)

type ExecutionPrinter interface {
	Start() error

	Stop()

	PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error
}
