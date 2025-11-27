package print_builtin

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	PrintBuiltinName = "print"
	// UsePlanFromKurtosisInstructionError We want users to use kurtosis_instruction.kurtosis_print as that adds to the instruction queue
	// and resolves future values
	UsePlanFromKurtosisInstructionError = "The default `print` function isn't supported please use the one that comes with the `plan` like `plan.print(...)` as that works with multi-phase execution and resolves future references."
)

// GeneratePrintBuiltin This only exists to throw a nice interpretation error when print without plan is used
func GeneratePrintBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return nil, startosis_errors.NewInterpretationError(UsePlanFromKurtosisInstructionError)
	}
}
